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
	defer func() { monitor.ObserveOpsDiagnose(time.Since(start)) }()
	incidentID := int64(0)
	if input.AlertID > 0 {
		incident, err := mysqlDAO.EnsureOpenIncidentForAlert(input.AlertID)
		if err != nil {
			return 0, fmt.Errorf("diagnose: ensure incident: %w", err)
		}
		incidentID = incident.ID
	}

	// 1. 创建占位报告（status=running）
	reportID := snowflake.GenID()
	report := &model.OpsReport{
		ID:          reportID,
		AlertID:     input.AlertID,
		IncidentID:  incidentID,
		TriggerType: input.TriggerType,
		Status:      "running",
	}
	if err := mysqlDAO.CreateOpsReport(report); err != nil {
		return 0, fmt.Errorf("diagnose: create report: %w", err)
	}

	// 2. 采集证据并持久化每一次工具调用
	recorder := &toolRecorder{reportID: reportID}
	evidence := collectEvidence(ctx, input, recorder)
	toolCalls := runEvidenceLoop(ctx, input, evidence, recorder)

	// 3. 查询最近变更
	changes := recorder.execute(ctx, "recent_changes_query", "检查告警窗口内的部署变更", func() mcp.Evidence {
		return mcp.RecentChangesQuery(ctx, input.Service, 10)
	})

	// 4. RAG 检索 Runbook 和历史报告
	var ragResult *RAGResult
	ragEvidence := recorder.execute(ctx, "incident_rag_search", "检索相关 Runbook 与相似历史诊断", func() mcp.Evidence {
		result, err := IncidentRAGSkill(ctx, input.AlertName, input.Service, input.AlertName+" "+input.Service)
		if err != nil {
			return mcp.Evidence{"rag_error": err.Error()}
		}
		ragResult = result
		return mcp.Evidence{
			"runbook_matches":           result.RunbookSummaries,
			"historical_report_matches": result.ReportSummaries,
		}
	})

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
	if errText, ok := ragEvidence["rag_error"].(string); ok && errText != "" {
		evidenceText += "\nRAG 检索异常：" + errText
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
	toolCalls = append(toolCalls, "recent_changes_query", "incident_rag_search")

	// 7. 更新报告完整内容
	finalReport := &model.OpsReport{
		ID:             reportID,
		AlertID:        input.AlertID,
		IncidentID:     incidentID,
		TriggerType:    input.TriggerType,
		Summary:        summaryResult.Summary,
		Evidence:       evidenceList,
		RootCause:      summaryResult.RootCause,
		Impact:         summaryResult.Impact,
		Suggestions:    toJSONSlice(summaryResult.Suggestions),
		RelatedChanges: changesList,
		ToolCalls:      toJSONSlice(append(toolCalls, summaryResult.ToolCalls...)),
		Status:         status,
	}
	if err := mysqlDAO.DB.Save(finalReport).Error; err != nil {
		return reportID, fmt.Errorf("diagnose: save report: %w", err)
	}

	return reportID, nil
}

// collectEvidence 根据工作流类型采集对应证据
func collectEvidence(ctx context.Context, input DiagnoseInput, recorder *toolRecorder) mcp.Evidence {
	combined := make(mcp.Evidence)

	// 公共证据：慢请求 + 错误事件 + 队列状态 + 告警列表
	merge(combined, recorder.execute(ctx, "slow_request_query", "获取最近慢请求样本", func() mcp.Evidence { return mcp.SlowRequestQuery(ctx, 5) }))
	merge(combined, recorder.execute(ctx, "error_event_query", "获取最近错误事件", func() mcp.Evidence { return mcp.ErrorEventQuery(ctx, "", 10) }))
	merge(combined, recorder.execute(ctx, "redis_stream_stats", "检查异步任务积压", func() mcp.Evidence { return mcp.RedisStreamStats(ctx) }))
	merge(combined, recorder.execute(ctx, "alert_query", "读取当前 firing 告警", func() mcp.Evidence { return mcp.AlertQuery(ctx, 5) }))
	merge(combined, recorder.execute(ctx, "prometheus_snapshot", "获取当前聚合指标", func() mcp.Evidence { return mcp.PrometheusSnapshot(ctx, input.AlertName) }))
	merge(combined, recorder.execute(ctx, "prometheus_range_query", "识别最近三十分钟趋势", func() mcp.Evidence { return mcp.PrometheusRangeSnapshot(ctx, input.AlertName, time.Now()) }))
	merge(combined, recorder.execute(ctx, "kubernetes_snapshot", "检查 Pod、Event、Deployment 与 Helm 历史", func() mcp.Evidence { return mcp.KubernetesSnapshot(ctx, "techmind", input.Service) }))

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
		"SearchLatencyHigh":      "SearchLatencyWorkflow",
		"RedisStreamBacklogHigh": "WorkerBacklogWorkflow",
		"CacheHitRateLow":        "CacheMissWorkflow",
		"PodRestartHigh":         "PodRestartWorkflow",
		"AICallFailureHigh":      "AIFailureWorkflow",
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
	system := `你是 TechMind 的高级 SRE 诊断报告 Skill。你只能依据提供的工具证据给出结论，绝不能把猜测写成事实。
约束：
1. 工具均为只读；不得建议直接执行删除 Pod、扩容、发布或任意 Shell 命令。
2. 先区分“已确认”“高可能”“证据不足”；证据冲突或不足时必须说明缺口与下一步查询方向。
3. 根因必须关联至少一条具体证据（指标、Pod/Event、日志、发布、队列或 Runbook）。
4. 建议必须按风险排序，优先给出验证动作和可回滚动作。
严格按以下格式输出：
摘要：（1-2句话）
影响：（1句话）
根因：（1-2句话，明确置信度：高/中/低）
证据：
- <证据及来源>
- <证据及来源>
未确认项：<没有则写无>
建议：
- <低风险验证动作>
- <处理建议>
- <后续预防建议>`

	user := fmt.Sprintf("工作流：%s\n告警：%s\n\n证据：\n%s", workflow, alertName, evidenceText)

	raw, err := aiModel.Chat(ctx, "ops_report", system, user)
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
