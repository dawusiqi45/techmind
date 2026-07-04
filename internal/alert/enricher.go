package alert

import (
	"context"
	"fmt"
	"time"

	mysqlDAO "techmind/internal/dao/mysql"
	redisDAO "techmind/internal/dao/redis"
	"techmind/internal/model"
)

// EnrichAlert 根据告警类型自动补充上下文，结果写入 alert_enrichment 表。
// 失败不影响主流程，只记录错误。
func EnrichAlert(ctx context.Context, event *model.AlertEvent) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	enrichCtx := collectContext(ctx, event)
	if len(enrichCtx) == 0 {
		return
	}

	enrichment := &model.AlertEnrichment{
		AlertID: event.ID,
		Context: enrichCtx,
	}
	_ = mysqlDAO.CreateAlertEnrichment(enrichment)
}

// collectContext 根据告警名称采集对应上下文证据
func collectContext(ctx context.Context, event *model.AlertEvent) model.JSONMap {
	result := make(model.JSONMap)

	switch event.AlertName {
	case "SearchLatencyHigh":
		enrichSearchLatency(ctx, result)
	case "RedisStreamBacklogHigh":
		enrichRedisBacklog(ctx, result)
	case "AICallFailureHigh":
		enrichAIFailure(ctx, result)
	case "APIHighErrorRate":
		enrichAPIErrors(ctx, result)
	case "WorkerConsumeLagHigh":
		enrichWorkerLag(ctx, result)
	default:
		// 通用增强：补充最近慢请求和错误事件数
		enrichGeneral(ctx, result)
	}

	// 所有告警都补充最近部署变更（10分钟内）
	changes, err := mysqlDAO.GetRecentChanges(event.Service, event.LastSeenAt, 10)
	if err == nil && len(changes) > 0 {
		changeDescs := make([]string, 0, len(changes))
		for _, c := range changes {
			changeDescs = append(changeDescs, fmt.Sprintf(
				"%s: %s -> %s at %s",
				c.Service, c.OldImage, c.Image, c.ChangedAt.Format("15:04:05"),
			))
		}
		result["recent_changes"] = changeDescs
	}

	return result
}

func enrichSearchLatency(ctx context.Context, result model.JSONMap) {
	slowList, total, _ := mysqlDAO.ListSlowRequests(1, 5)
	result["slow_request_total"] = total
	samples := make([]string, 0, len(slowList))
	for _, s := range slowList {
		samples = append(samples, fmt.Sprintf("%s %s %dms", s.Method, s.Path, s.DurationMs))
	}
	result["slow_request_samples"] = samples

	pending, _ := redisDAO.PendingAITasks(ctx)
	result["ai_task_pending"] = pending
}

func enrichRedisBacklog(ctx context.Context, result model.JSONMap) {
	pending, _ := redisDAO.PendingAITasks(ctx)
	length, _ := redisDAO.StreamLen(ctx, redisDAO.StreamAITasks)
	deadLen, _ := redisDAO.StreamLen(ctx, redisDAO.StreamAIDeadLetter)
	result["ai_task_pending"] = pending
	result["ai_stream_length"] = length
	result["ai_dead_letter_length"] = deadLen
}

func enrichAIFailure(ctx context.Context, result model.JSONMap) {
	errList, total, _ := mysqlDAO.ListErrorEvents("ai", 1, 5)
	result["ai_error_total"] = total
	samples := make([]string, 0, len(errList))
	for _, e := range errList {
		samples = append(samples, e.Message)
	}
	result["ai_error_samples"] = samples
}

func enrichAPIErrors(ctx context.Context, result model.JSONMap) {
	errList, total, _ := mysqlDAO.ListErrorEvents("http", 1, 10)
	result["http_error_total"] = total
	topPaths := make(map[string]int)
	for _, e := range errList {
		topPaths[e.Path] += e.Count
	}
	result["top_error_paths"] = topPaths
}

func enrichWorkerLag(ctx context.Context, result model.JSONMap) {
	pending, _ := redisDAO.PendingAITasks(ctx)
	result["worker_pending"] = pending
	errList, total, _ := mysqlDAO.ListErrorEvents("worker", 1, 5)
	result["worker_error_total"] = total
	samples := make([]string, 0, len(errList))
	for _, e := range errList {
		samples = append(samples, e.Message)
	}
	result["worker_error_samples"] = samples
}

func enrichGeneral(ctx context.Context, result model.JSONMap) {
	_, slowTotal, _ := mysqlDAO.ListSlowRequests(1, 1)
	_, errTotal, _ := mysqlDAO.ListErrorEvents("", 1, 1)
	result["slow_request_total"] = slowTotal
	result["error_event_total"] = errTotal
}
