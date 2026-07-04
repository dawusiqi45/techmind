// Package mcp 提供 SRE Agent 可调用的只读诊断工具集。
// 每个工具函数采集一类证据，返回结构化 map，供 OpsDiagnoseSkill 聚合后交给 LLM 汇总。
// 所有工具均为只读，不修改任何数据。
package mcp

import (
	"context"
	"fmt"
	"time"

	mysqlDAO "techmind/internal/dao/mysql"
	redisDAO "techmind/internal/dao/redis"
	"techmind/internal/model"
)

// Evidence 是工具采集到的证据
type Evidence map[string]interface{}

// SlowRequestQuery 查询最近 topN 条慢请求
func SlowRequestQuery(ctx context.Context, topN int) Evidence {
	list, total, err := mysqlDAO.ListSlowRequests(1, topN)
	if err != nil {
		return Evidence{"error": err.Error()}
	}
	samples := make([]string, 0, len(list))
	for _, s := range list {
		samples = append(samples, fmt.Sprintf("[%s] %s %s %dms",
			s.CreatedAt.Format("15:04:05"), s.Method, s.Path, s.DurationMs))
	}
	return Evidence{
		"slow_request_total":   total,
		"slow_request_samples": samples,
	}
}

// ErrorEventQuery 查询最近错误事件
func ErrorEventQuery(ctx context.Context, source string, topN int) Evidence {
	list, total, err := mysqlDAO.ListErrorEvents(source, 1, topN)
	if err != nil {
		return Evidence{"error": err.Error()}
	}
	samples := make([]string, 0, len(list))
	for _, e := range list {
		samples = append(samples, fmt.Sprintf("[%s] %s: %s (count=%d)",
			e.Source, e.Path, e.Message, e.Count))
	}
	return Evidence{
		"error_event_total":   total,
		"error_event_samples": samples,
	}
}

// RedisStreamStats 查询 AI 任务队列状态
func RedisStreamStats(ctx context.Context) Evidence {
	pending, err1 := redisDAO.PendingAITasks(ctx)
	length, err2 := redisDAO.StreamLen(ctx, redisDAO.StreamAITasks)
	deadLen, err3 := redisDAO.StreamLen(ctx, redisDAO.StreamAIDeadLetter)
	ev := Evidence{
		"ai_task_pending":     pending,
		"ai_stream_length":    length,
		"ai_dead_letter_length": deadLen,
	}
	if err1 != nil { ev["pending_error"] = err1.Error() }
	if err2 != nil { ev["length_error"] = err2.Error() }
	if err3 != nil { ev["dead_error"] = err3.Error() }
	return ev
}

// AlertQuery 查询 firing 状态告警（最近 topN 条）
func AlertQuery(ctx context.Context, topN int) Evidence {
	list, total, err := mysqlDAO.ListAlertEvents(model.AlertStatusFiring, 1, topN)
	if err != nil {
		return Evidence{"error": err.Error()}
	}
	summaries := make([]string, 0, len(list))
	for _, a := range list {
		summaries = append(summaries, fmt.Sprintf("[%s] %s (severity=%s, repeat=%d)",
			a.AlertName, a.Service, a.Severity, a.RepeatCount))
	}
	return Evidence{
		"firing_alert_total": total,
		"alert_summaries":    summaries,
	}
}

// RecentChangesQuery 查询最近 windowMin 内的部署变更
func RecentChangesQuery(ctx context.Context, service string, windowMin int) Evidence {
	changes, err := mysqlDAO.GetRecentChanges(service, time.Now(), windowMin)
	if err != nil {
		return Evidence{"error": err.Error()}
	}
	descs := make([]string, 0, len(changes))
	for _, c := range changes {
		descs = append(descs, fmt.Sprintf("[%s] %s %s->%s at %s",
			c.Service, c.Source, c.OldImage, c.Image, c.ChangedAt.Format("15:04:05")))
	}
	return Evidence{
		"recent_change_count": len(changes),
		"recent_changes":      descs,
	}
}
