package agent

import (
	"context"
	"regexp"
	"strings"
	"time"

	"techmind/internal/agent/mcp"
	mysqlDAO "techmind/internal/dao/mysql"
	"techmind/internal/model"

	"go.uber.org/zap"
)

const maxAuditTextLength = 4096

var (
	bearerCredentialPattern = regexp.MustCompile(`(?i)\bbearer\s+[a-z0-9._~+/=-]+`)
	namedSecretPattern      = regexp.MustCompile(`(?i)\b(authorization|password|passwd|token|api[_-]?key|secret|cookie)\b\s*[:=]\s*[^\s,;]+`)
)

// toolRecorder 将每一次只读取证持久化，供报告详情页显示完整证据链。
type toolRecorder struct {
	reportID int64
}

func (r *toolRecorder) execute(ctx context.Context, name, reason string, fn func() mcp.Evidence) mcp.Evidence {
	start := time.Now()
	output := fn()
	input := model.JSONMap{"reason": sanitizeAuditString(reason)}
	if err := mysqlDAO.CreateOpsToolCall(ctx, &model.OpsToolCall{
		ReportID:   r.reportID,
		ToolName:   name,
		Input:      input,
		Output:     sanitizeAuditEvidence(output),
		DurationMs: int(time.Since(start).Milliseconds()),
	}); err != nil {
		zap.L().Warn("persist ops tool call failed", zap.Int64("report_id", r.reportID), zap.String("tool", name), zap.Error(err))
	}
	return output
}

// sanitizeAuditEvidence 防止审计表和管理端证据链意外保存敏感信息或无限长输出。
// 它不修改传给诊断模型的原始证据，避免审计展示逻辑影响诊断结果。
func sanitizeAuditEvidence(input mcp.Evidence) model.JSONMap {
	result := make(model.JSONMap, len(input))
	for key, value := range input {
		if isSensitiveAuditKey(key) {
			result[key] = "[redacted]"
			continue
		}
		result[key] = sanitizeAuditValue(value)
	}
	return result
}

func sanitizeAuditValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return sanitizeAuditString(v)
	case []string:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = sanitizeAuditString(item)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = sanitizeAuditValue(item)
		}
		return result
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for key, item := range v {
			if isSensitiveAuditKey(key) {
				result[key] = "[redacted]"
				continue
			}
			result[key] = sanitizeAuditValue(item)
		}
		return result
	case model.JSONMap:
		result := make(model.JSONMap, len(v))
		for key, item := range v {
			if isSensitiveAuditKey(key) {
				result[key] = "[redacted]"
				continue
			}
			result[key] = sanitizeAuditValue(item)
		}
		return result
	case mcp.Evidence:
		return sanitizeAuditEvidence(v)
	default:
		return value
	}
}

func sanitizeAuditString(value string) string {
	value = bearerCredentialPattern.ReplaceAllString(value, "Bearer [redacted]")
	value = namedSecretPattern.ReplaceAllString(value, "$1=[redacted]")
	if len(value) > maxAuditTextLength {
		return strings.TrimSpace(value[:maxAuditTextLength]) + "... [truncated]"
	}
	return value
}

func isSensitiveAuditKey(key string) bool {
	key = strings.ToLower(key)
	for _, fragment := range []string{"authorization", "password", "passwd", "token", "api_key", "api-key", "secret", "cookie"} {
		if strings.Contains(key, fragment) {
			return true
		}
	}
	return false
}
