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
| AI | CloudWeGo Eino | Agent 框架，LLM + Embedding 封装 |
| LLM | DeepSeek（OpenAI 兼容接口） | 摘要生成、搜索总结、诊断报告 |
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
│    agent.Diagnose() → Prometheus/DB/Redis证据 → 报告 │
│  /metrics:9091 暴露 Worker 指标                      │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│              Prometheus + Alertmanager               │
│                                                     │
│  Prometheus 每 5s 抓取 /metrics                     │
│  触发告警 → Alertmanager → Webhook                   │
│           → POST /api/v1/alerts/webhook（Bearer）    │
│           → 告警入库 → 可触发 OpsWorker 诊断         │
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
| 中间件 | `internal/middleware` | CORS / RequestID / JWT / Admin / Logger / Metrics / RateLimit / SlowRequest / Recovery |
| 指标 | `internal/monitor` | 14 个 Prometheus 指标定义和记录函数 |
| 告警 | `internal/alert` | Alertmanager Webhook 解析、SHA256 指纹去重、告警增强 |
| Agent | `internal/agent` | SRE 诊断主流程（采集证据 → RAG → LLM → 写报告） |
| Agent MCP | `internal/agent/mcp` | 只读诊断工具（Prometheus、慢请求/错误事件、队列、告警、变更查询） |
| Agent RAG | `internal/agent/rag` | Runbook 精确召回 + Milvus 语义召回 + 历史报告检索 |
| AI 模型 | `internal/ai/model` | DeepSeek LLM 单次对话封装 |
| AI 向量 | `internal/ai/embedding` | Ark Embedding 单文本和批量向量 |
| AI Prompt | `internal/ai/prompt` | 摘要/标签/搜索总结 Skill Prompt |
| Worker | `internal/worker` | AIWorker（3次重试→死信）+ OpsWorker（3次重试→ops死信） |
| 工具包 | `internal/pkg` | jwt / snowflake / response / settings / logger / health / validator |

---

## 数据库表（20张）

| 分类 | 表名 |
|---|---|
| 论坛业务 | `user` `article` `tag` `article_tag` `comment` `favorite` `user_like` |
| AI 与向量 | `article_chunk` `ai_task` `ai_call_record` |
| SRE | `runbook` `ops_report` `ops_tool_call` `deployment_change` |
| 可观测 | `monitor_slow_request` `monitor_error_event` |
| 告警 | `alert_event` `alert_enrichment` `incident` `incident_alert` |

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
| `/admin/ops/reports/:id` | 摘要/证据/根因/建议 | 深色 |
| `/admin/ops/diagnose` | 手动触发诊断 | 深色 |
| `/admin/runbooks` | Runbook 列表 + 新建 | 深色 |
| `/admin/deployments` | 部署变更列表 + 记录 | 深色 |

---

## 后端 API（API v1 37条 + 系统路由 3条）

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
