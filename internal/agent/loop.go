package agent

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"techmind/internal/agent/mcp"
	aiModel "techmind/internal/ai/model"
)

// maxEvidenceRounds 限制 Agent 在基础取证后可自主追加的只读查询轮数，避免无限调用。
const maxEvidenceRounds = 5

type loopDecision struct {
	Action string `json:"action"`
	Tool   string `json:"tool"`
	Reason string `json:"reason"`
}

// runEvidenceLoop 让模型根据已有证据决定是否继续调用受限的只读工具。
func runEvidenceLoop(ctx context.Context, input DiagnoseInput, evidence mcp.Evidence, recorder *toolRecorder) []string {
	calls := []string{"prometheus_snapshot", "prometheus_range_query", "kubernetes_snapshot", "slow_request_query", "error_event_query", "redis_stream_stats", "alert_query"}
	for round := 0; round < maxEvidenceRounds; round++ {
		decision, ok := decideNextTool(ctx, input, evidence)
		if !ok || decision.Action == "final" {
			break
		}
		var next mcp.Evidence
		switch decision.Tool {
		case "prometheus_range_query":
			next = recorder.execute(ctx, decision.Tool, decision.Reason, func() mcp.Evidence { return mcp.PrometheusRangeSnapshot(ctx, input.AlertName, time.Now()) })
		case "kubernetes_snapshot":
			next = recorder.execute(ctx, decision.Tool, decision.Reason, func() mcp.Evidence { return mcp.KubernetesSnapshot(ctx, "techmind", input.Service) })
		case "kubernetes_logs":
			next = recorder.execute(ctx, decision.Tool, decision.Reason, func() mcp.Evidence { return mcp.KubernetesLogSnapshot(ctx, "techmind", input.Service) })
		case "slow_request_query":
			next = recorder.execute(ctx, decision.Tool, decision.Reason, func() mcp.Evidence { return mcp.SlowRequestQuery(ctx, 20) })
		case "error_event_query":
			next = recorder.execute(ctx, decision.Tool, decision.Reason, func() mcp.Evidence { return mcp.ErrorEventQuery(ctx, "", 20) })
		case "redis_stream_stats":
			next = recorder.execute(ctx, decision.Tool, decision.Reason, func() mcp.Evidence { return mcp.RedisStreamStats(ctx) })
		default:
			break
		}
		if next == nil {
			break
		}
		merge(evidence, next)
		calls = append(calls, decision.Tool+": "+decision.Reason)
	}
	return calls
}

func decideNextTool(ctx context.Context, input DiagnoseInput, evidence mcp.Evidence) (loopDecision, bool) {
	prompt := `你是 TechMind 的只读 SRE 诊断规划器。你的职责是根据已有证据决定下一步最有价值的查询，不是直接猜测根因。
规则：
1. 只能使用给定的只读工具；不得建议或执行写操作、Shell、跨 namespace 查询。
2. 优先补充能区分不同根因的证据；日志仅在指标、Pod 或 Event 有异常线索时查询。
3. 没有足够证据时继续查询；证据已能支持结论或没有有效下一步时返回 final。
4. 只能返回一行合法 JSON，不得使用 Markdown：
{"action":"tool"或"final","tool":"prometheus_range_query|kubernetes_snapshot|kubernetes_logs|slow_request_query|error_event_query|redis_stream_stats","reason":"不超过40字的证据缺口说明"}`
	raw, err := aiModel.Chat(ctx, "ops_planner", prompt, "告警："+input.AlertName+"\n服务："+input.Service+"\n现有证据：\n"+formatEvidence(evidence, nil))
	if err != nil {
		return loopDecision{}, false
	}
	start, end := strings.Index(raw, "{"), strings.LastIndex(raw, "}")
	if start < 0 || end < start {
		return loopDecision{}, false
	}
	var decision loopDecision
	if err := json.Unmarshal([]byte(raw[start:end+1]), &decision); err != nil || (decision.Action != "tool" && decision.Action != "final") {
		return loopDecision{}, false
	}
	return decision, true
}
