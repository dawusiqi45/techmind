# TechMind SRE Agent V2：循环推理与证据链诊断设计

## 1. 目标

将当前“收集固定数据后交给 LLM 总结”的诊断实现，升级为一个**只读、可循环取证、可审计**的 SRE 诊断 Agent。

## 当前实现状态（2026-07）

已实现：firing 告警按 `alert_id + startsAt` 自动、原子去重入队；诊断任务携带 `task_key` 与证据时间窗，Worker 重试和 stale claim 复用同一份报告；告警诊断自动聚合到 `incident`、`ops_report.incident_id` 回链；慢请求、错误、Prometheus Range 与部署变更按告警窗口查询；基础取证（MySQL/Redis/Prometheus/Kubernetes/Helm）与模型规划的追加取证；最多 5 轮追加只读查询；可配置 120 秒总超时、Kubernetes 10 秒请求超时；受限 Pod 日志；`ops_tool_call` 审计和管理台证据链；Runbook 索引进入可靠 Worker 队列；最终报告生成结构化的只读排查命令、需审批修改方案、验证命令和回滚命令，并经过危险命令与敏感信息过滤。

尚未实现、保留为后续演进：独立的 `diagnosis_run` / `diagnosis_step` / `hypothesis` 表、工具调用显式 success/error 状态、服务级日志/错误关联、更强的历史报告相似度检索和 Milvus 向量检索的集群部署。

它要回答的不只是“发生了什么”，还要回答：

- 异常从何时开始，影响了什么；
- Agent 查了哪些系统、使用了哪些参数；
- 每条根因结论由哪些实际证据支撑；
- 证据不足时下一步应该继续查什么，而不是编造结论。
- 后续人员可以复制哪些只读命令，修改什么、如何验证以及如何回滚。

## 2. 不做的内容

- 不构建通用 Agent 平台，不引入 CRD、Controller Reconcile、A2A 或 Python/Go 双运行时。
- 不允许 Agent 直接执行任何报告命令；它可以生成需人工审批的扩缩容、发布或回滚建议，但不能生成删除资源、读取 Secret、清库或主机重启命令。
- 不提供任意 Shell/SSH 工具；Helm 和日志能力必须封装成参数受限的只读工具。

这保证方案与 TechMind 的 Go + Eino + Redis Stream + MySQL 架构保持一致。

## 3. 总体流程

```text
Alertmanager 告警 / 管理员手动触发
        ↓
创建 Incident 与 DiagnosisRun
        ↓
AgentLoop（最多 N 轮、总时限 M 秒）
  1. LLM 读取当前上下文与已有证据
  2. 选择一个允许的只读工具及结构化参数
  3. Tool Executor 执行、限流、脱敏、压缩结果
  4. 写入 DiagnosisStep + OpsToolCall + Evidence
  5. LLM 判断：继续取证 / 输出最终结论
        ↓
结构化报告：影响、时间线、根因候选、证据、只读排查命令、修改/验证/回滚方案
```

当前停止条件：模型输出 `final`、模型输出不可解析、达到 5 轮预算或达到诊断总超时。证据置信度阈值和“连续两轮无新增信息”仍属于后续增强。

## 4. AgentLoop

### 4.1 状态

```go
type DiagnosisState struct {
    IncidentID      int64
    RunID           int64
    AlertName       string
    Service         string
    Namespace       string
    TimeRange       TimeRange
    Evidence        []EvidenceItem
    Hypotheses      []Hypothesis
    RemainingBudget int
}
```

每轮只允许模型返回以下两种动作：

```json
{"action":"tool","tool":"prometheus_range_query","arguments":{...},"reason":"确认异常开始时间"}
```

```json
{"action":"final","summary":"...","hypotheses":[...],"confidence":0.78}
```

模型输出必须经过 JSON Schema 校验；工具名称与参数不在白名单内时拒绝执行并把错误写入下一轮上下文。

### 4.2 初始上下文

由现有 `AlertEvent`、`AlertEnrichment`、`DeploymentChange` 和手动输入构成：

- 告警名、服务、端点、严重度、首次/最近发生时间；
- 默认诊断窗口：以 Alertmanager `startsAt` 为锚点前后各 15 分钟，未来部分截断到当前；手动诊断默认最近 30 分钟；所有窗口最多 60 分钟；
- 当前镜像、namespace、关联发布变更；
- 已存在的慢请求、错误事件、队列与 Runbook 摘要。

### 4.3 推理策略

固定的首轮不是固定结论，而是最低限度的证据采集：

1. Prometheus Range Query 确定 QPS、错误率、P95/P99 的异常开始时间；
2. 根据服务名查询 Pod/Deployment/Event；
3. 根据异常窗口查询近期 Helm 发布与 `deployment_change`；
4. 后续由模型依据证据选择日志、依赖或 Runbook 工具。

对 `SearchLatencyHigh`，典型路径是：Prometheus 延迟趋势 → Server Pod/Event → Milvus/AI 调用指标 → 相关日志 → 发布变更。

## 5. 只读工具集

| 工具 | 参数约束 | 输出摘要 | 用途 |
|---|---|---|---|
| `prometheus_instant_query` | PromQL 白名单、5 秒超时 | 当前标量/向量 | 当前 QPS、错误率、队列状态 |
| `prometheus_range_query` | 最大 60 分钟、最大 120 点 | 趋势、峰值、异常起点 | P95/P99、错误率基线对比 |
| `k8s_get_pods` | 允许 namespace/service label | 状态、重启、Ready、资源使用 | 判断 Pod 是否异常 |
| `k8s_get_events` | 最大 50 条、限定时间窗 | Warning Event 摘要 | OOM、拉镜像、调度失败 |
| `k8s_get_deployment` | 单服务只读 | 副本、镜像、可用副本 | 判断发布与容量变化 |
| `k8s_get_logs` | 单 Pod/容器、最近 10 分钟、最多 200 行 | 错误模式与少量样本 | 关联运行时异常 |
| `helm_release_history` | 固定 namespace/release | 版本、时间、镜像/values 差异 | 发布关联 |
| `slow_request_query` | 最大 20 条 | 路径、耗时、时间 | 使用现有表 |
| `error_event_query` | 最大 20 条 | 来源、错误、次数 | 使用现有表 |
| `redis_stream_stats` | 固定 Stream | pending、lag、死信 | Worker 堆积 |
| `runbook_search` | Top 5 | Runbook/历史报告摘要 | 给出标准处理建议 |

工具执行器必须完成：namespace 白名单、RBAC、最大返回量、超时、敏感字段脱敏、结果截断和错误归一化。日志中禁止返回 Token、密码、Authorization Header、完整 Cookie。

## 6. 数据模型与现有表的衔接

保留现有 `alert_event`、`alert_enrichment`、`ops_report`、`deployment_change`、`runbook`。当前用 `incident` + `incident_alert` 聚合告警，用 `ops_report.incident_id` 关联报告，用 `ops_tool_call` 保存实际取证调用。

新增或补全以下记录：

| 实体 | 关键字段 | 说明 |
|---|---|---|
| `incident` | `id, status, severity` + `incident_alert` | 当前已实现的故障生命周期，可关联多次诊断 |
| `ops_report` | `id, incident_id, trigger_type, status, verification_commands, change_plan, validation_commands, rollback_commands` | 当前的一次诊断运行、最终报告与人工操作手册 |
| `ops_tool_call` | `id, report_id, tool_name, input, output, duration_ms` | 当前已实现，保存真实工具调用 |
| `diagnosis_step` | `id, run_id, round, action, reason, status, duration_ms` | 后续拆分时引入 |
| `evidence_item` | `id, run_id, source, observed_at, content, importance` | 可引用、可排序的证据 |
| `hypothesis` | `id, run_id, rank, statement, confidence, supporting_evidence_ids` | 根因候选与置信度 |

现有 `OpsReport` 继续作为最终展示对象，并新增关联 `incident_id`。其 `ToolCalls` 仍保留兼容字段；前端以 `ops_tool_call` 的真实审计记录作为证据链来源。

## 7. 前端设计

在现有“诊断报告详情”页新增四个区域：

1. **故障概览**：服务、告警、影响窗口、当前状态、最终置信度。
2. **诊断时间线**：第几轮、为什么要查、调用什么工具、耗时、成功/失败。
3. **证据面板**：Prometheus 趋势摘要、Pod/Event、日志模式、Helm 变更，可展开查看受限原始结果。
4. **根因与建议**：按置信度区分结论，明确标注证据不足的部分。
5. **操作手册**：可复制只读排查命令、需审批的修改步骤、修改后验证命令与回滚命令；每项展示目的、风险和判断标准。

管理员可看到“停止原因”：得到结论、无新增证据、达到工具预算或超时。

## 8. 实施顺序

### Phase 1：循环与审计基础

- 新增 `Incident`、`DiagnosisRun`、`DiagnosisStep`、真实 `OpsToolCall` 持久化。
- 把当前 `Diagnose` 改为 `AgentLoop`；保留 Redis Stream 异步执行与失败重试。
- 将报告和每一步状态通过 API 返回前端。

### Phase 2：Prometheus Range 与 Kubernetes 只读工具

- 先实现 Prometheus Instant/Range Query、Pod、Event、Deployment。
- 使用 `client-go` 及命名空间受限 ServiceAccount；日志工具最后加入。
- 为每个告警定义首轮工具模板，但允许后续模型自行选择工具。

### Phase 3：Helm、日志与证据时间线

- 增加只读 Helm 发布历史与差异摘要，不提供任意 CLI。
- 增加有行数/时间窗限制的 Pod 日志工具及敏感信息脱敏。
- 完成诊断时间线、根因置信度与证据引用 UI。

## 9. 验收标准

- 对 `SearchLatencyHigh`，报告至少包含 Prometheus 时间范围证据、Pod/Event 结果、一次变更关联判断。
- 每次工具调用在数据库和前端时间线都有名称、参数摘要、耗时、状态和结果摘要。
- Agent 无法调用未注册工具、跨 namespace 查询、写操作、无限时间范围或超过预算的日志查询。
- Agent 生成的只读命令必须通过白名单；修改和回滚必须标记人工审批，危险、复合或含凭据命令不得入库。
- LLM、Prometheus、K8s 任一依赖失败时，报告明确标注“证据不足/工具失败”，不编造根因。
- 正常或异常案例均可重复演示，并能在 Grafana/Prometheus 中交叉验证。
