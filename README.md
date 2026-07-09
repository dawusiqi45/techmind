# TechMind

TechMind 是一个基于 Go/Gin 的技术论坛与云原生智能可观测平台。系统以技术论坛作为真实业务场景，支持文章发布、评论互动、热榜排行、关键词/语义搜索和搜索结果 AI 总结；后台围绕论坛业务构建可观测闭环，采集 HTTP 延迟、错误率、慢请求、缓存命中率、Redis Stream 积压、Milvus 检索耗时和 AI 调用状态。系统接入 Prometheus 和 Alertmanager，实现告警中心、告警去重、Robusta 式告警增强，并参考 HolmesGPT 设计 SRE Agent，通过 MCP 只读工具和 Runbook RAG 聚合指标、日志、队列、数据库、K8s 事件和历史故障，生成结构化诊断报告。

## 技术栈

| 类型 | 技术 | 用途 |
|---|---|---|
| 语言 | Go 1.24+ | 后端主语言 |
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
| K8s | client-go | Pod/Deployment/Event 查询 |
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
├── frontend/                   React 18 + Vite + TypeScript 前端
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
│   ├── sql/init.sql            数据库初始化（15张表）
│   └── seed_data.go            演示数据种子
├── docs/                       项目文档
│   ├── 项目进度.md             完整开发日志（阶段1-13）
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
| techmind-worker | - | 异步任务 Worker |
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
| GET | `/api/v1/admin/monitor/endpoints` | 接口性能统计 |
| GET | `/api/v1/admin/monitor/slow-requests` | 慢请求列表 |
| GET | `/api/v1/admin/monitor/errors` | 错误事件列表 |
| GET | `/api/v1/admin/monitor/queues` | Redis Stream 队列状态 |
| GET | `/api/v1/admin/monitor/ai-calls` | AI 调用观测 |
| POST | `/api/v1/admin/alerts/webhook` | Alertmanager Webhook |
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
| POST | `/api/v1/admin/runbooks` | 新增 Runbook |
| GET | `/api/v1/admin/runbooks` | Runbook 列表 |
| POST | `/api/v1/admin/deployment-changes` | 记录部署变更 |

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

详见 [docs/项目进度.md](docs/项目进度.md)（共 13 个阶段，全部完成）。

- **后端**（阶段1-10）：Go/Gin API Server + Worker，论坛业务 + AI + 可观测 + 告警 + SRE Agent。
- **前端**（阶段11-12）：React 18 + Vite + TypeScript，论坛浅色主题 + 管理后台深色主题，前后端接口已对齐。
- **部署**（阶段13）：Docker Compose + Helm Chart + kind 集群一键部署脚本。
