package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"techmind/internal/agent/mcp"
	aiModel "techmind/internal/ai/model"
	mysqlDAO "techmind/internal/dao/mysql"
	"techmind/internal/model"
	"techmind/internal/monitor"
	"techmind/internal/pkg/snowflake"
)

// DiagnoseInput 触发诊断的参数
type DiagnoseInput struct {
	AlertID     int64  // 0=手动触发
	TriggerType string // manual / alert
	Service     string // 相关服务名，可为空
	AlertName   string // 告警名称，用于选择工作流
}

// Diagnose 执行诊断并将报告写入数据库，返回报告 ID
func Diagnose(ctx context.Context, input DiagnoseInput) (int64, error) {
	start := time.Now()

	// 1. 创建占位报告（status=running）
	reportID := snowflake.GenID()
	report := &model.OpsReport{
		ID:          reportID,
		AlertID:     input.AlertID,
		TriggerType: input.TriggerType,
		Status:      "running",
	}
	if err := mysqlDAO.CreateOpsReport(report); err != nil {
		return 0, fmt.Errorf("diagnose: create report: %w", err)
	}

	// 2. 采集证据
	evidence := collectEvidence(ctx, input)

	// 3. 查询最近变更
	changes := mcp.RecentChangesQuery(ctx, input.Service, 10)

	// 4. RAG 检索 Runbook 和历史报告
	ragResult, _ := IncidentRAGSkill(ctx, input.AlertName, input.Service,
		input.AlertName+" "+input.Service)

	// 5. 识别诊断类型选择工作流，组装 LLM Prompt（编号修正）
	workflowName := identifyWorkflow(input.AlertName)
	evidenceText := formatEvidence(evidence, changes)

	// 把 RAG 结果附加到证据文本，帮助 LLM 参考历史处理经验
	if ragResult != nil {
		if len(ragResult.RunbookSummaries) > 0 {
			evidenceText += "\n相关 Runbook：\n- " + strings.Join(ragResult.RunbookSummaries, "\n- ")
		}
		if len(ragResult.ReportSummaries) > 0 {
			evidenceText += "\n历史报告：\n- " + strings.Join(ragResult.ReportSummaries, "\n- ")
		}
	}

	// 6. 调用 LLM 汇总诊断报告
	summaryResult, err := callDiagnoseSkill(ctx, workflowName, input.AlertName, evidenceText)
	duration := time.Since(start)
	monitor.ObserveAICall("ops_diagnose", duration, err)

	status := "done"
	if err != nil {
		status = "failed"
		_ = mysqlDAO.UpdateOpsReportStatus(reportID, status)
		return reportID, fmt.Errorf("diagnose: llm failed: %w", err)
	}

	// 6. 组装结构化报告字段
	evidenceList := make(model.JSONSlice, 0)
	for k, v := range evidence {
		evidenceList = append(evidenceList, fmt.Sprintf("%s: %v", k, v))
	}

	changesList := make(model.JSONSlice, 0)
	if rc, ok := changes["recent_changes"].([]string); ok {
		for _, c := range rc {
			changesList = append(changesList, c)
		}
	}

	// 7. 更新报告完整内容
	finalReport := &model.OpsReport{
		ID:             reportID,
		AlertID:        input.AlertID,
		TriggerType:    input.TriggerType,
		Summary:        summaryResult.Summary,
		Evidence:       evidenceList,
		RootCause:      summaryResult.RootCause,
		Impact:         summaryResult.Impact,
		Suggestions:    toJSONSlice(summaryResult.Suggestions),
		RelatedChanges: changesList,
		ToolCalls:      toJSONSlice(summaryResult.ToolCalls),
		Status:         status,
	}
	if err := mysqlDAO.DB.Save(finalReport).Error; err != nil {
		return reportID, fmt.Errorf("diagnose: save report: %w", err)
	}

	return reportID, nil
}

// collectEvidence 根据工作流类型采集对应证据
func collectEvidence(ctx context.Context, input DiagnoseInput) mcp.Evidence {
	combined := make(mcp.Evidence)

	// 公共证据：慢请求 + 错误事件 + 队列状态 + 告警列表
	merge(combined, mcp.SlowRequestQuery(ctx, 5))
	merge(combined, mcp.ErrorEventQuery(ctx, "", 10))
	merge(combined, mcp.RedisStreamStats(ctx))
	merge(combined, mcp.AlertQuery(ctx, 5))

	return combined
}

func merge(dst, src mcp.Evidence) {
	for k, v := range src {
		dst[k] = v
	}
}

// identifyWorkflow 根据告警名称匹配内置工作流
func identifyWorkflow(alertName string) string {
	workflows := map[string]string{
		"SearchLatencyHigh":     "SearchLatencyWorkflow",
		"RedisStreamBacklogHigh": "WorkerBacklogWorkflow",
		"CacheHitRateLow":       "CacheMissWorkflow",
		"PodRestartHigh":        "PodRestartWorkflow",
		"AICallFailureHigh":     "AIFailureWorkflow",
	}
	if w, ok := workflows[alertName]; ok {
		return w
	}
	return "GeneralDiagnoseWorkflow"
}

// diagnoseResult 是 LLM 汇总结果的结构化表示
type diagnoseResult struct {
	Summary     string
	RootCause   string
	Impact      string
	Suggestions []string
	ToolCalls   []string
}

// callDiagnoseSkill 调用 LLM 对采集到的证据进行诊断汇总
func callDiagnoseSkill(ctx context.Context, workflow, alertName, evidenceText string) (*diagnoseResult, error) {
	system := `你是一个 SRE 诊断助手。根据告警名称和采集到的系统证据，输出一份结构化诊断报告。
格式：
摘要：（1-2句话）
根因：（1-2句话）
影响：（1句话）
建议：（3条，每条一行，以"-"开头）`

	user := fmt.Sprintf("工作流：%s\n告警：%s\n\n证据：\n%s", workflow, alertName, evidenceText)

	raw, err := aiModel.ChatOnce(ctx, system, user)
	if err != nil {
		return &diagnoseResult{
			Summary:   fmt.Sprintf("诊断失败，告警名称：%s", alertName),
			RootCause: "LLM 调用失败，请检查 AI 配置",
		}, err
	}

	return parseRawDiagnose(raw, evidenceText), nil
}

// parseRawDiagnose 从 LLM 自由文本中提取结构化字段（宽松解析）
func parseRawDiagnose(raw, evidenceText string) *diagnoseResult {
	result := &diagnoseResult{
		Summary:   extractSection(raw, "摘要"),
		RootCause: extractSection(raw, "根因"),
		Impact:    extractSection(raw, "影响"),
	}
	if result.Summary == "" {
		result.Summary = raw
	}

	// 提取建议行
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			result.Suggestions = append(result.Suggestions, strings.TrimPrefix(line, "- "))
		}
	}
	return result
}

// extractSection 从 LLM 输出中提取指定标签后的内容
func extractSection(text, label string) string {
	prefix := label + "："
	idx := strings.Index(text, prefix)
	if idx < 0 {
		return ""
	}
	after := text[idx+len(prefix):]
	end := strings.Index(after, "\n")
	if end < 0 {
		return strings.TrimSpace(after)
	}
	return strings.TrimSpace(after[:end])
}

// formatEvidence 将 Evidence map 格式化为 LLM 可读文本
func formatEvidence(ev mcp.Evidence, changes mcp.Evidence) string {
	var sb strings.Builder
	for k, v := range ev {
		sb.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
	}
	for k, v := range changes {
		sb.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
	}
	return sb.String()
}

func toJSONSlice(ss []string) model.JSONSlice {
	result := make(model.JSONSlice, len(ss))
	for i, s := range ss {
		result[i] = s
	}
	return result
}
