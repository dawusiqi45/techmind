# TechMind 架构文档

## 系统概述

TechMind 是一个技术论坛与云原生智能可观测平台。以技术论坛作为真实业务场景，后台围绕论坛业务构建可观测闭环，接入 Prometheus + Alertmanager，并实现 SRE Agent 自动诊断。

---

## 技术栈

| 层 | 技术 | 用途 |
|---|---|---|
| 前端 | React 19 + Vite 8 + TypeScript | 论坛（浅色）+ 管理后台（深色）双主题 SPA |
| HTTP 框架 | Go + Gin | API Server，路由/中间件/响应 |
| 配置 | Viper | YAML 配置 + 环境变量覆盖 |
| 日志 | Zap + Lumberjack | 结构化日志 + 切割 |
| 数据库 | MySQL 8 + GORM | 业务数据（用户/文章/告警/报告等） |
| 缓存/队列 | Redis 7 | 热榜 ZSet + Redis Stream 异步队列 + 限流 |
| 向量库 | Milvus | 文章语义搜索 + Runbook RAG |
| AI | CloudWeGo Eino | LLM / Embedding 封装与 Skill 调用 |
| LLM | OpenAI 兼容接口（默认 DeepSeek） | 摘要生成、搜索总结、诊断规划与报告 |
| Embedding | DashScope text-embedding-v4 | 文章和 Runbook 向量化 |
| 指标 | Prometheus | HTTP/缓存/队列/AI/Milvus 指标 |
| 告警 | Alertmanager | 告警推送和路由 |
| ID 生成 | Snowflake | 分布式唯一 ID |

---

## 系统分层

```
┌─────────────────────────────────────────────────────┐
│                    浏览器                             │
│   /forum/*  浅色论坛主题                              │
│   /admin/*  深色管理后台（JWT + 服务端 role=1 鉴权）  │
└───────────────────┬─────────────────────────────────┘
                    │ HTTP
                    ▼
┌─────────────────────────────────────────────────────┐
│             Nginx (frontend Pod)                     │
│   静态文件托管 + /api/* 反代到 techmind-server:8080   │
└───────────────────┬─────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────┐
│              API Server (techmind-server)            │
│                                                     │
│  中间件链：CORS → RequestID → Logger → Metrics      │
│            → RateLimit → SlowRequest → Recovery     │
│                                                     │
│  router ─→ controller ─→ logic ─→ dao               │
└──────┬──────────┬───────────────┬───────────────────┘
       │          │               │
       ▼          ▼               ▼
    MySQL       Redis           Milvus
   (业务数据)  (热榜/限流/    (向量搜索)
               Stream队列)
                    │
                    │ 写入 ai_tasks / ops_tasks Stream
                    ▼
┌─────────────────────────────────────────────────────┐
│              Worker (techmind-worker)                │
│                                                     │
│  AIWorker  ──消费 ai_tasks Stream──→                │
│    article.summary  → LLM 生成摘要                  │
│    article.tag      → LLM 提取标签                  │
│    article.index    → Embedding + 写入 Milvus        │
│                                                     │
│  OpsWorker ──消费 ops_tasks Stream──→               │
│    AgentLoop：基础取证 + 最多 5 轮追加只读查询        │
│    DB/Redis/Prometheus/K8s/变更/Runbook → LLM → 报告 │
│    告警诊断：Incident ← OpsReport ← OpsToolCall      │
│  /metrics:9091 暴露 Worker 指标                      │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│              Prometheus + Alertmanager               │
│                                                     │
│  Prometheus 每 5s 抓取 /metrics                     │
│  触发告警 → Alertmanager → Webhook                   │
│           → POST /api/v1/alerts/webhook（Bearer）    │
│           → 告警入库 → firing 告警自动去重触发诊断  │
└─────────────────────────────────────────────────────┘
```

---

## 模块职责

| 模块 | 路径 | 职责 |
|---|---|---|
| 入口 | `cmd/server` `cmd/worker` | HTTP Server 和 Worker 的启动、优雅退出 |
| 路由 | `internal/router` | 注册 API v1 37 条路由和 3 条系统路由，全局中间件链 |
| 控制器 | `internal/controller` | 参数绑定、调用 logic、统一响应 |
| 业务逻辑 | `internal/logic` | 业务编排，跨 DAO 协调，不涉及 HTTP |
| 数据访问 | `internal/dao/mysql` | GORM DAO，覆盖 20 张业务和可观测数据表 |
| 数据访问 | `internal/dao/redis` | 热榜 ZSet、Stream、限流计数 |
| 数据访问 | `internal/dao/milvus` | 向量写入、删除、ANN 搜索 |
| 模型 | `internal/model` | 14 个 GORM 模型，定义表结构 |
| 中间件 | `internal/middleware` | CORS / BodyLimit / RequestID / JWT / Admin / Logger / Metrics / RateLimit / SlowRequest / Recovery |
| 指标 | `internal/monitor` | 14 个 Prometheus 指标定义和记录函数 |
| 告警 | `internal/alert` | Alertmanager Webhook 解析、SHA256 指纹去重、告警增强 |
| Agent | `internal/agent` | SRE 循环诊断、Incident 关联、真实工具调用审计与报告汇总 |
| Agent MCP | `internal/agent/mcp` | 进程内受限只读工具：Prometheus Instant/Range、慢请求/错误、队列、告警、变更、Pod/Event/Deployment、Pod 日志；尚未独立为标准 MCP ToolServer |
| Agent RAG | `internal/agent/rag` | Runbook 精确召回 + Milvus 语义召回 + 历史报告检索 |
| AI 模型 | `internal/ai/model` | OpenAI 兼容 LLM 调用、Skill 标签与调用审计 |
| AI 向量 | `internal/ai/embedding` | Ark Embedding 单文本和批量向量 |
| AI Prompt | `internal/ai/prompt` | 摘要/标签/搜索总结 Skill Prompt |
| Worker | `internal/worker` | AIWorker（3次重试→死信）+ OpsWorker（3次重试→ops死信） |
| 工具包 | `internal/pkg` | jwt / snowflake / response / settings / logger / health / validator |

---

## 数据库表

| 分类 | 表名 |
|---|---|
| 论坛业务 | `user` `article` `tag` `article_tag` `comment` `favorite` `user_like` |
| AI 与向量 | `article_chunk` `ai_task` `ai_call_record` |
| SRE | `runbook` `ops_report` `ops_tool_call` `deployment_change` |
| 可观测 | `monitor_slow_request` `monitor_error_event` |
| 告警 | `alert_event` `alert_enrichment` `incident` `incident_alert` |
| 诊断幂等键 | `ops_report.task_key`（唯一；Worker 重试与 stale claim 复用同一报告） |

---

## 前端路由

| 路由 | 页面 | 主题 |
|---|---|---|
| `/` | 文章列表 + 热榜 + 标签云 | 浅色 |
| `/login` | 登录/注册（Modal 弹窗） | 浅色 |
| `/search` | AI 搜索总结 + 结果列表 | 浅色 |
| `/articles/:id` | 文章详情 + 评论树 | 浅色 |
| `/articles/new` `/articles/:id/edit` | Markdown 分屏编辑器 | 浅色 |
| `/user/profile` | 个人资料编辑 + 我的文章/收藏/点赞 | 浅色 |
| `/admin/monitor` | 4 指标卡 + 慢请求/错误表格 | 深色 |
| `/admin/monitor/slow` `errors` `queues` `ai` | 各监控子页 | 深色 |
| `/admin/alerts` | 告警列表（状态筛选） | 深色 |
| `/admin/alerts/:id` | 告警详情 + 确认/诊断 | 深色 |
| `/admin/ops/reports` | 诊断报告列表 | 深色 |
| `/admin/ops/reports/:id` | 摘要、根因、建议、Incident 回链与真实工具证据链 | 深色 |
| `/admin/ops/diagnose` | 手动触发诊断 | 深色 |
| `/admin/incidents` `/admin/incidents/:id` | 故障事件、关联告警与人工关闭 | 深色 |
| `/admin/runbooks` | Runbook 列表 + 新建 | 深色 |
| `/admin/deployments` | 部署变更列表 + 记录 | 深色 |

---

## 后端 API

### 认证与用户

| 方法 | 路径 | 鉴权 | 说明 |
|---|---|---|---|
| POST | `/api/v1/auth/register` | 无 | 注册 |
| POST | `/api/v1/auth/login` | 无 | 登录 |
| POST | `/api/v1/auth/refresh` | 无 | 刷新Token |
| GET  | `/api/v1/user/profile` | JWT | 当前用户信息 |
| PUT  | `/api/v1/user/profile` | JWT | 更新用户资料 |
| POST | `/api/v1/user/avatar` | JWT | 上传头像 |
| GET  | `/api/v1/user/favorites` | JWT | 我的收藏 |
| GET  | `/api/v1/user/likes` | JWT | 我的点赞 |
| GET  | `/api/v1/user/articles` | JWT | 我的文章 |

### 论坛

| 方法 | 路径 | 鉴权 |
|---|---|---|
| GET  | `/api/v1/articles` | 无 |
| GET  | `/api/v1/articles/hot` | 无 |
| GET  | `/api/v1/articles/:id` | 无 |
| POST | `/api/v1/articles` | JWT |
| PUT  | `/api/v1/articles/:id` | JWT |
| DELETE | `/api/v1/articles/:id` | JWT |
| POST | `/api/v1/articles/:id/like` | JWT |
| POST | `/api/v1/articles/:id/favorite` | JWT |
| GET  | `/api/v1/articles/:id/comments` | 无 |
| POST | `/api/v1/articles/:id/comments` | JWT |
| GET  | `/api/v1/tags` | 无 |
| GET  | `/api/v1/search` | 无 |

### Alertmanager Webhook（独立 Bearer Token）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/v1/alerts/webhook` | Alertmanager Webhook（Bearer Token） |

### 管理后台（JWT + 服务端管理员角色校验）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET  | `/api/v1/admin/monitor/overview` | 监控总览 |
| GET  | `/api/v1/admin/monitor/slow-requests` | 慢请求列表 |
| GET  | `/api/v1/admin/monitor/errors` | 错误事件列表 |
| GET  | `/api/v1/admin/monitor/queues` | Redis Stream 队列状态 |
| GET  | `/api/v1/admin/monitor/ai-calls` | AI 调用记录 |
| GET  | `/api/v1/admin/alerts` | 告警列表 |
| GET  | `/api/v1/admin/alerts/:id` | 告警详情 |
| POST | `/api/v1/admin/alerts/:id/ack` | 确认告警 |
| POST | `/api/v1/admin/alerts/:id/diagnose` | 对告警触发诊断 |
| POST | `/api/v1/admin/ops/diagnose` | 手动触发诊断 |
| GET  | `/api/v1/admin/ops/reports` | 诊断报告列表 |
| GET  | `/api/v1/admin/ops/reports/:id` | 诊断报告详情 |
| GET  | `/api/v1/admin/ops/reports/:id/timeline` | 诊断真实工具调用证据链 |
| GET  | `/api/v1/admin/incidents` | 故障事件列表 |
| GET  | `/api/v1/admin/incidents/:id` | 故障事件和关联告警 |
| POST | `/api/v1/admin/incidents/:id/resolve` | 人工关闭故障事件，不修改原告警状态 |
| POST | `/api/v1/admin/runbooks` | 新增 Runbook |
| GET  | `/api/v1/admin/runbooks` | Runbook 列表 |
| POST | `/api/v1/admin/deployment-changes` | 记录部署变更 |
| GET  | `/api/v1/admin/deployment-changes` | 部署变更列表 |

### 系统

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/healthz` | 存活探针 |
| GET | `/readyz` | 就绪探针（检查 MySQL + Redis + Milvus） |
| GET | `/metrics` | Prometheus 指标 |

---

## 告警规则（7条）

| 告警名 | 触发条件 | 级别 |
|---|---|---|
| APIHighErrorRate | 5xx 错误率 > 5%，持续 5 分钟 | critical |
| APILatencyHigh | P95 延迟 > 1s，持续 5 分钟 | warning |
| RedisStreamBacklogHigh | Stream pending > 100，持续 3 分钟 | warning |
| AICallFailureHigh | AI 调用失败率 > 10%，持续 5 分钟 | critical |
| SearchLatencyHigh | 搜索接口平均延迟 > 1s，持续 2 分钟 | warning |
| CacheHitRateLow | 缓存命中率 < 60%，持续 5 分钟 | warning |
| OpsDiagnoseDurationHigh | 诊断报告 P95 > 120s，持续 5 分钟 | warning |

---

## 部署方式

| 方式 | 入口 | 适用场景 |
|---|---|---|
| 本地开发 | `go run cmd/server/main.go` + `npm run dev` | 开发调试 |
| Docker Compose | `cd deploy/docker && docker compose up -d` | 单机演示 |
| Kind K8s | `bash deploy/kind/deploy.sh` | 模拟集群，验证 Helm + Pod |
| Helm | `helm install techmind ./deploy/helm/techmind` | 生产 K8s |

### Kind 部署组件清单

| Pod | 镜像 | 说明 |
|---|---|---|
| techmind-server | techmind-server:latest（自建） | API Server |
| techmind-worker | techmind-worker:latest（自建） | 异步 Worker，9091 暴露 `/metrics` |
| techmind-frontend | techmind-frontend:latest（自建） | Nginx + 前端静态文件 |
| mysql-0 | mysql:8.0 | StatefulSet + PVC 5Gi |
| redis | redis:7-alpine | Deployment |
| prometheus | prom/prometheus:v2.55.0 | NodePort 30909 |
| alertmanager | prom/alertmanager:v0.27.0 | ClusterIP |
| metrics-server | metrics-server:v0.7.1 | kube-system，HPA 依赖 |
| ingress-nginx | nginx-ingress-controller | ingress-nginx namespace |

---

## 依赖关系

- **MySQL**：必须，所有业务数据存储
- **Redis**：必须，热榜 ZSet + Stream 任务队列 + 限流
- **Milvus**：可选，失败自动降级为 MySQL 关键词搜索
- **AI（LLM + Embedding）**：可选，失败不阻断启动，相关功能静默降级
- **Prometheus**：Agent 可选依赖；不可用时会把查询异常写入证据，无法完成趋势分析
- **Kubernetes API**：Agent 的 Pod/Event/Deployment/日志工具依赖 Worker 的只读 ServiceAccount；仅允许 `techmind` 命名空间且不授予 Secret 权限
- **安全边界**：非本地模式拒绝默认/弱 JWT Secret；HTTP 业务错误返回真实 4xx/5xx；请求体、头像、Server 超时和可信代理均有限制

## SRE Agent 运行闭环

```text
Alertmanager firing Webhook 自动去重入队 / 管理员手动触发
        ↓
Redis Stream: ops_tasks（task_key + 时间窗）→ OpsWorker
        ↓
基础取证：时间窗内慢请求、错误、Prometheus、部署变更 + 当前队列、告警、Kubernetes
        ↓
LLM 规划：最多再选择 5 次受限只读工具调用（可提前 final）
        ↓
最近部署变更 + Runbook/历史报告检索
        ↓
LLM 结构化报告 → ops_report（根因 + 只读排查命令 + 需审批修改/回滚方案）
        ├── alert 触发时关联/复用 incident
        └── 每一次工具调用写入 ops_tool_call，管理端展示证据链
```

Agent 默认总时限 120 秒、单次 Kubernetes 请求 10 秒，Prometheus Range 最多查询 60 分钟；Agent 只读访问集群。生成的排查/验证命令经过只读白名单，修改与回滚强制人工审批；Agent 不执行报告内命令。关闭 Incident 也是管理员操作，不会自动修改告警状态。
