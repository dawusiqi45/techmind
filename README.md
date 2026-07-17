# TechMind

TechMind 是一个基于 Go/Gin 的技术论坛与云原生智能可观测平台。系统以技术论坛作为真实业务场景，支持文章发布、评论互动、热榜排行、关键词/语义搜索和搜索结果 AI 总结；后台围绕论坛业务构建可观测闭环，采集 HTTP 延迟、错误率、慢请求、缓存命中率、Redis Stream 积压、Milvus 检索耗时和 AI 调用状态。系统接入 Prometheus 和 Alertmanager，实现告警中心、告警去重、告警增强，并提供只读、可循环取证的 SRE Agent。

## 当前已实现的核心能力

- **论坛与用户**：文章、评论、标签、点赞、收藏、个人中心、热榜、JWT 鉴权与管理员后台。
- **AI 内容能力**：文章摘要、AI 标签、搜索总结；Milvus 与 Embedding 可用时启用文章语义搜索和 Runbook 语义检索，不可用时保留关键词/数据库降级路径。
- **可观测与告警**：Prometheus 指标、慢请求和错误事件归档、Redis Stream 队列观测、Alertmanager Webhook、告警去重与增强。
- **SRE Agent**：firing 告警自动去重入队，也支持管理员手动触发；诊断按告警时间窗采集 Prometheus、慢请求、错误和部署变更，默认总时限 120 秒；基础取证后由 LLM 最多追加 5 轮只读查询；Worker 重试复用同一报告，生成 Incident、结构化报告和可审计证据链。
- **部署**：Docker Compose、Helm、kind 部署；Worker 使用最小只读 RBAC 查询 Kubernetes。Milvus/MinIO/etcd 在当前 kind 配置中默认不部署。

## 技术栈

| 类型 | 技术 | 用途 |
|---|---|---|
| 语言 | Go 1.25 | 后端主语言 |
| Web 框架 | Gin | HTTP API、Middleware、路由 |
| 配置 | Viper | YAML 配置、环境变量覆盖 |
| 日志 | Zap + Lumberjack | 结构化日志和日志切割 |
| 数据库 | MySQL 8 + GORM | 用户、文章、评论、告警、诊断报告 |
| 缓存 | Redis 7 | 缓存、热榜、限流、Stream 队列 |
| 向量库 | Milvus | 语义搜索、Runbook RAG |
| AI 框架 | CloudWeGo Eino | Agent/工具调用 |
| LLM | DeepSeek / OpenAI 兼容接口 | 摘要、搜索总结、诊断报告 |
| Embedding | DashScope text-embedding-v4 | 文章和 Runbook 向量化 |
| 指标 | Prometheus | HTTP、缓存、队列、AI、Milvus 指标 |
| 告警 | Alertmanager | 告警推送和路由 |
| K8s | Helm + kind + client-go | 本地集群部署、服务暴露、HPA 验证与 Agent 只读查询 |
| 部署 | Docker Compose + Helm | 本地演示和 K8s 部署 |

## 项目结构

```text
TechMind/
├── cmd/
│   ├── server/main.go          API Server 入口
│   └── worker/main.go          异步任务 Worker 入口
├── internal/
│   ├── controller/             HTTP 协议层
│   ├── logic/                  业务编排层
│   ├── dao/                    数据访问层
│   │   ├── mysql/              MySQL GORM DAO
│   │   ├── redis/              Redis + Stream
│   │   └── milvus/             Milvus 向量检索
│   ├── monitor/                Prometheus 指标定义
│   ├── alert/                  告警接收、去重、增强
│   ├── agent/                  SRE Agent、MCP、RAG
│   ├── ai/                     LLM、Embedding、Prompt
│   ├── worker/                 异步任务消费、重试、死信
│   ├── middleware/             Gin 中间件
│   ├── model/                  数据模型
│   └── pkg/                    通用包（jwt/snowflake/response/logger/settings）
├── frontend/                   React 19 + Vite + TypeScript 前端
│   ├── src/
│   │   ├── api/                Axios API 模块（9个）
│   │   ├── store/              Zustand 状态管理（auth/theme）
│   │   ├── layouts/            ForumLayout（浅色）/ AdminLayout（深色）
│   │   ├── pages/forum/        论坛页（Home/Login/Search/ArticleDetail/Editor/UserProfile）
│   │   ├── pages/admin/        管理后台页（Monitor/Alert/Ops/Runbook/Deployment 等）
│   │   ├── components/         公共组件（ArticleCard/StatCard/AlertBadge）
│   │   └── utils/              token 管理 + 时间格式化
│   ├── nginx.conf              /api/ 反代 + SPA fallback
│   └── Dockerfile              多阶段构建（Node→Nginx）
├── config/
│   ├── config.yaml             本地开发配置
│   └── config.example.yaml     配置样例
├── deploy/
│   ├── docker/docker-compose.yml   Docker Compose 全栈编排
│   ├── helm/techmind/              Helm Chart（server/worker/frontend）
│   ├── kind/                       kind 集群部署
│   │   ├── cluster.yaml            集群定义
│   │   ├── deploy.sh               一键部署脚本（8步）
│   │   ├── mysql.yaml / redis.yaml 基础设施
│   │   ├── prometheus.yaml / alertmanager.yaml 监控
│   │   ├── metrics-server.yaml     HPA 依赖
│   │   └── values-kind.yaml        Helm 专用配置
│   └── prometheus/                 Prometheus + Alertmanager 配置
├── scripts/
│   ├── sql/init.sql            数据库初始化（20张表）
│   └── seed_data.go            演示数据种子
├── docs/                       项目文档
│   ├── architecture.md         架构和 API 文档
│   ├── operations.md           运维与增量更新手册
│   └── 项目范围.md             模块边界说明
└── Dockerfile                  后端多阶段构建（server/worker 两个 target）
```

## 快速开始

### 本地开发

1. 安装依赖服务（MySQL 8、Redis 7、Milvus 2.4+）
2. 初始化数据库：

```bash
mysql -u root -p < scripts/sql/init.sql
```

3. 修改 `config/config.yaml` 中的连接信息和 AI API Key
4. 启动 Server：

```bash
go run cmd/server/main.go
```

5. 启动 Worker（新终端）：

```bash
go run cmd/worker/main.go
```

6. 写入演示数据：

```bash
go run scripts/seed_data.go
```

### Docker Compose 一键启动

```bash
cd deploy/docker
docker-compose up -d
```

服务列表：

| 服务 | 端口 | 说明 |
|---|---|---|
| techmind-server | 8080 | API Server |
| techmind-worker | 9091 | 异步任务 Worker，暴露 Prometheus metrics |
| mysql | 3306 | MySQL 8.0 |
| redis | 6379 | Redis 7 |
| milvus-standalone | 19530 | Milvus 向量库 |
| prometheus | 9090 | Prometheus 指标采集 |
| alertmanager | 9093 | Alertmanager 告警路由 |

启动后写入演示数据：

```bash
go run scripts/seed_data.go
```

### Kubernetes 部署（Helm）

```bash
cd deploy/helm/techmind
helm install techmind . -n techmind --create-namespace \
  --set externalMySQL.host=your-mysql-host \
  --set externalMySQL.password=your-password \
  --set externalRedis.host=your-redis-host \
  --set externalMilvus.host=your-milvus-host
```

### kind 本地完整部署验证

`deploy/kind/deploy.sh` 会创建名为 `techmind` 的 kind 集群，预加载 registry.k8s.io 插件镜像，安装固定版本的 ingress-nginx、metrics-server、MySQL、Redis、Prometheus、Alertmanager，并通过 Helm 部署 TechMind。

```bash
cd deploy/kind
bash deploy.sh
```

如果之前的 kind 集群已有半安装失败的插件，建议先重建本地验证集群：

```bash
kind delete cluster --name techmind
bash deploy.sh
```

部署完成后验证：

```bash
kubectl get pods -A
kubectl top nodes
```

访问地址：

```text
前端:       http://<虚拟机IP>:30000
Prometheus: http://<虚拟机IP>:30909
```

## API 概览

### 论坛业务

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/v1/auth/register` | 注册 |
| POST | `/api/v1/auth/login` | 登录 |
| POST | `/api/v1/auth/refresh` | 刷新 Token |
| GET | `/api/v1/user/profile` | 当前用户信息 |
| PUT | `/api/v1/user/profile` | 更新用户资料 |
| POST | `/api/v1/user/avatar` | 上传头像 |
| GET | `/api/v1/user/favorites` | 我的收藏列表 |
| GET | `/api/v1/user/likes` | 我的点赞列表 |
| POST | `/api/v1/articles` | 发布文章 |
| GET | `/api/v1/articles` | 文章列表 |
| GET | `/api/v1/articles/:id` | 文章详情 |
| PUT | `/api/v1/articles/:id` | 编辑文章 |
| DELETE | `/api/v1/articles/:id` | 删除文章 |
| POST | `/api/v1/articles/:id/like` | 点赞 |
| POST | `/api/v1/articles/:id/favorite` | 收藏 |
| GET | `/api/v1/articles/hot` | 热榜 |
| POST | `/api/v1/articles/:id/comments` | 评论 |
| GET | `/api/v1/articles/:id/comments` | 评论列表 |
| GET | `/api/v1/tags` | 标签列表 |
| GET | `/api/v1/search` | 关键词+语义搜索+AI 总结 |

### 监控与告警

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/admin/monitor/overview` | 监控总览 |
| GET | `/api/v1/admin/monitor/slow-requests` | 慢请求列表 |
| GET | `/api/v1/admin/monitor/errors` | 错误事件列表 |
| GET | `/api/v1/admin/monitor/queues` | Redis Stream 队列状态 |
| GET | `/api/v1/admin/monitor/ai-calls` | AI 调用观测 |
| POST | `/api/v1/alerts/webhook` | Alertmanager Webhook（Bearer Token） |
| GET | `/api/v1/admin/alerts` | 告警列表 |
| GET | `/api/v1/admin/alerts/:id` | 告警详情 |
| POST | `/api/v1/admin/alerts/:id/ack` | 确认告警 |

### SRE Agent

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/v1/admin/ops/diagnose` | 手动触发诊断 |
| POST | `/api/v1/admin/alerts/:id/diagnose` | 对告警触发诊断 |
| GET | `/api/v1/admin/ops/reports` | 诊断报告列表 |
| GET | `/api/v1/admin/ops/reports/:id` | 诊断报告详情 |
| GET | `/api/v1/admin/ops/reports/:id/timeline` | 诊断真实工具调用证据链 |
| GET | `/api/v1/admin/incidents` | 故障事件列表 |
| GET | `/api/v1/admin/incidents/:id` | 故障事件与关联告警 |
| POST | `/api/v1/admin/incidents/:id/resolve` | 管理员关闭故障事件（不修改告警状态） |
| POST | `/api/v1/admin/runbooks` | 新增 Runbook |
| GET | `/api/v1/admin/runbooks` | Runbook 列表 |
| POST | `/api/v1/admin/deployment-changes` | 记录部署变更 |
| GET | `/api/v1/admin/deployment-changes` | 部署变更列表 |

## 使用监控与 SRE Agent

监控后台和 SRE Agent 都在管理员登录后的前端管理台中：进入 `/admin/monitor` 查看慢请求、错误事件、队列与 AI 调用。Alertmanager 推送 firing 告警后，Server 会按 `alert_id + startsAt` 原子去重并自动写入 `ops_tasks`；也可在 **SRE 诊断 → 手动触发**（`/admin/ops/diagnose`）主动提交。任务由 `techmind-worker` 异步执行，失败重试继续使用同一个 `task_key` 和报告；在 **SRE 诊断 → 诊断报告**（`/admin/ops/reports`）查看状态、时间窗证据、根因、可复制的只读排查命令、需审批的修改方案、修改后验证命令和回滚命令。

Agent 只生成操作手册，不执行其中任何命令。排查与验证命令必须通过只读白名单和单命令安全过滤；修改、扩缩容、发布与回滚建议始终标记为“需要人工审批”。删除资源、读取 Secret、清库、命令管道/重定向以及包含凭据的命令不会保存到报告。

Prometheus 用于核实采集是否正常：访问 `http://<虚拟机IP>:30909/targets`，确认 `techmind-server` 和 `techmind-worker` 目标为 **UP**；在 Graph 页面查询 `http_requests_total` 或 `http_request_duration_seconds_count`。访问论坛、搜索和管理端 API 后，这些指标应增长。告警规则满足持续时间后由 Alertmanager 携带内部 Webhook Bearer Token 回调 `/api/v1/alerts/webhook`，并出现在 **告警中心**；也可从单条告警详情直接触发诊断。

SRE 行为可通过 `ops.autoDiagnose`、`ops.diagnoseTimeoutSec`、`ops.evidenceWindowMin` 配置，或分别使用 `TECHMIND_OPS_AUTO_DIAGNOSE`、`TECHMIND_OPS_DIAGNOSE_TIMEOUT_SEC`、`TECHMIND_OPS_EVIDENCE_WINDOW_MIN` 覆盖。

### 配置 AI 与 Webhook 密钥

仓库不保存真实 API Key。kind 更新或首次部署前，请在 Ubuntu 终端设置自己的值，并以 Helm Secret 注入：

```bash
export TECHMIND_LLM_API_KEY='你的 DeepSeek/OpenAI 兼容 API Key'
export TECHMIND_ALERT_WEBHOOK_TOKEN='随机生成的长令牌'

helm upgrade --install techmind ./deploy/helm/techmind \
  -n techmind \
  -f deploy/kind/values-kind.yaml \
  --set-string secrets.llmApiKey="$TECHMIND_LLM_API_KEY" \
  --set-string secrets.alertWebhookToken="$TECHMIND_ALERT_WEBHOOK_TOKEN"
```

同时将 `deploy/kind/alertmanager-config.yaml` 中的 `credentials` 改为同一个 Webhook Token，再执行 `kubectl apply -f deploy/kind/alertmanager-config.yaml` 和 `kubectl rollout restart deployment/alertmanager -n techmind`。生产环境应改用外部 Secret/Vault，而非命令历史或 values 文件。

完整的增量更新、迁移和验收步骤见 [docs/operations.md](docs/operations.md)。

循环推理、Kubernetes 只读工具与证据链诊断的后续实现方案见 [docs/sre-agent-v2-design.md](docs/sre-agent-v2-design.md)。

### 系统

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/healthz` | 存活检查 |
| GET | `/readyz` | 就绪检查（MySQL + Redis + Milvus） |
| GET | `/metrics` | Prometheus Metrics |

## 告警规则

项目内置以下 Prometheus 告警规则（`deploy/prometheus/techmind-rules.yml` 和 Helm Chart `prometheusrule.yaml`）：

| 告警 | 条件 | 级别 |
|---|---|---|
| APIHighErrorRate | 5xx 错误率 > 5%，持续 5 分钟 | critical |
| APILatencyHigh | P95 延迟 > 1s，持续 5 分钟 | warning |
| CacheHitRateLow | 缓存命中率 < 60%，持续 5 分钟 | warning |
| RedisStreamBacklogHigh | 队列 pending > 100，持续 3 分钟 | warning |
| AICallFailureHigh | AI 调用失败率 > 10%，持续 5 分钟 | critical |
| MilvusSearchLatencyHigh | Milvus P95 > 500ms，持续 5 分钟 | warning |
| OpsDiagnoseDurationHigh | 诊断报告 P95 > 120s，持续 5 分钟 | warning |

## 构建说明

本项目依赖 Milvus SDK 和 bytedance/sonic，需要 64 位 Go 编译环境：

```bash
go env -w GOARCH=amd64
go build ./...
```

## 开发进度

当前项目范围见 [docs/项目范围.md](docs/项目范围.md)，架构和 API 见 [docs/architecture.md](docs/architecture.md)，部署运维见 [docs/operations.md](docs/operations.md)。

- **后端**（阶段1-10）：Go/Gin API Server + Worker，论坛业务 + AI + 可观测 + 告警 + SRE Agent。
- **前端**（阶段11-12）：React 19 + Vite + TypeScript，论坛浅色主题 + 管理后台深色主题，前后端接口已对齐。
- **部署**（阶段13）：Docker Compose + Helm Chart + kind 集群一键部署脚本。

### 2026-07-17 SRE 可靠性迭代

- 告警 webhook 对 firing 告警自动诊断，Redis Lua 原子去重。
- 诊断任务新增 `task_key` 和时间窗，Worker 重试/崩溃接管复用同一报告。
- Prometheus Range、慢请求、错误事件和部署变更按告警窗口取证；补齐主要告警的 PromQL 模板。
- Agent 增加可配置总超时，Kubernetes 客户端增加 10 秒请求超时。
- Runbook 索引改由 AI Stream Worker 执行，获得重试、死信和 stale claim 能力。
- 修复 Compose Prometheus 地址、规则挂载、Embedding Key 注入及管理端队列/变更展示契约。
- 新增数据库迁移 `002_sre_agent_reliability.sql`；升级已有环境前必须执行。
- 报告新增结构化排查命令、修改方案、验证与回滚步骤；修复旧自由文本解析把证据误当建议的问题。
- 操作指令经过只读白名单、危险命令和敏感信息过滤，所有变更与回滚均强制人工审批，Agent 不自动执行。
- 新增数据库迁移 `003_sre_action_guidance.sql`；升级已有环境时在 `002` 之后执行。
