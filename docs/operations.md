# TechMind 运维与增量更新手册

## 诊断链路

```text
API Server / Worker 埋点 → Prometheus → Alertmanager
                                        ↓ Bearer Webhook
告警中心 ← MySQL ← TechMind Server ← /api/v1/alerts/webhook
                                        ↓
firing 告警自动原子去重入队 / 管理员手动触发 → ops_tasks → Worker
→ 慢请求、错误事件、Redis、Prometheus、Kubernetes、Helm、变更、Runbook → LLM → ops_report
                                                                        ↓
                          只读排查命令 + 需审批修改方案 + 验证/回滚命令 + ops_tool_call 审计
```

Agent 是只读诊断报告生成器：它不修改集群或业务数据。firing 告警按 `alert_id + startsAt` 在 Redis 中原子去重并自动入队；告警诊断创建或复用开放的 `incident`。任务携带唯一 `task_key`，失败重试和 stale claim 复用同一份 `ops_report`。慢请求、错误、Prometheus Range 和部署变更按告警时间窗查询；默认诊断总时限 120 秒。模型最多再选择 5 次只读查询。报告为 `done` 才代表成功；失败最多重试 3 次，随后进入 `tm:stream:ops_tasks:dead`。

报告操作区分为四类：`verification_commands` 和 `validation_commands` 只接受白名单内的单条只读命令；`change_plan` 和 `rollback_commands` 仅作为建议，后端强制标记 `approval_required=true`。命令连接、管道、重定向、命令替换、读取 Secret、删除资源、清库、主机重启和含凭据命令会被过滤。管理员复制执行前仍必须核对环境、命名空间、资源名、版本、容量和备份。

```bash
kubectl logs -n techmind deployment/techmind-worker --tail=100
kubectl get pods -n techmind
```

Prometheus `http://<VM-IP>:30909/targets` 中的 `techmind-server:8080`、`techmind-worker:9091` 均应为 `UP`。Grafana 用于查看趋势和大盘，不负责调用 Agent；Agent 使用 Prometheus API 获取受限的聚合证据。

## 首次部署前的密钥

真实 LLM API Key 和 Alertmanager Webhook Token 不得提交到 Git。通过 Helm Secret 注入：

```bash
export TECHMIND_LLM_API_KEY='你的 Key'
export TECHMIND_ALERT_WEBHOOK_TOKEN='随机长令牌'
```

令牌需要同时写入 `deploy/kind/alertmanager-config.yaml` 的 `credentials` 与 Helm 的 `secrets.alertWebhookToken`。LLM Key 写入 `secrets.llmApiKey`，Embedding Key 如需语义检索则写入 `secrets.embeddingApiKey`。完成 Helm 更新后重启 Alertmanager。

## 常规代码增量更新

不要删除 kind 集群，也不要为普通代码改动重跑完整 `deploy/kind/deploy.sh`。使用不可变镜像标签：

```bash
cd ~/projects/techmind
git pull --ff-only
VERSION=$(git rev-parse --short HEAD)

docker build --target server -t techmind-server:$VERSION .
docker build --target worker -t techmind-worker:$VERSION .
docker build -t techmind-frontend:$VERSION ./frontend

kind load docker-image techmind-server:$VERSION --name techmind
kind load docker-image techmind-worker:$VERSION --name techmind
kind load docker-image techmind-frontend:$VERSION --name techmind

helm upgrade techmind ./deploy/helm/techmind -n techmind \
  -f deploy/kind/values-kind.yaml \
  --set server.image.tag=$VERSION \
  --set worker.image.tag=$VERSION \
  --set frontend.image.tag=$VERSION

kubectl rollout status deployment/techmind-server -n techmind
kubectl rollout status deployment/techmind-worker -n techmind
kubectl rollout status deployment/techmind-frontend -n techmind
```

若本次版本包含数据库升级，在 `helm upgrade` 前执行对应的、可重复运行的 migration。当前 SRE Agent 证据链与 Incident 关联使用：

```bash
kubectl exec -i -n techmind mysql-0 -- \
  mysql --protocol=TCP -h 127.0.0.1 -P 3306 -utechmind -ptechmind techmind \
  < scripts/sql/migrations/001_sre_agent_audit_and_incident.sql

kubectl exec -i -n techmind mysql-0 -- \
  mysql --protocol=TCP -h 127.0.0.1 -P 3306 -utechmind -ptechmind techmind \
  < scripts/sql/migrations/002_sre_agent_reliability.sql

kubectl exec -i -n techmind mysql-0 -- \
  mysql --protocol=TCP -h 127.0.0.1 -P 3306 -utechmind -ptechmind techmind \
  < scripts/sql/migrations/003_sre_action_guidance.sql
```

`deploy/kind/deploy.sh` 已在 MySQL 就绪后依次执行三份 migration；日常增量部署需要在升级镜像前显式执行。`002` 会为既有报告生成 `legacy:<id>` 幂等键，再创建 `uk_task_key` 唯一索引；`003` 为既有报告补四个空 JSON 数组后增加结构化操作字段。三份迁移均可重复运行。

LLM 兼容服务地址和模型名也可随 Helm 更新覆盖；API Key 始终经 `Secret` 注入：

```bash
helm upgrade techmind ./deploy/helm/techmind -n techmind \
  -f deploy/kind/values-kind.yaml \
  --set-string secrets.llmApiKey="$TECHMIND_LLM_API_KEY" \
  --set-string ai.llmBaseURL='https://api.deepseek.com/v1' \
  --set-string ai.llmModel='deepseek-chat'
```

## 特殊更新

- 数据库表结构：编写并执行显式 migration；现有 MySQL PVC 不会重新执行 `init.sql`。
- Server/Worker ConfigMap 或 Secret：执行 `helm upgrade` 后执行 `kubectl rollout restart deployment/techmind-server deployment/techmind-worker -n techmind`。
- Prometheus/Grafana/Alertmanager 配置：`kubectl apply -f deploy/kind/<文件>.yaml` 后重启相应 Deployment；Worker 指标通过 Helm 的 `worker-service.yaml` 暴露到 9091，Prometheus 也可通过 lifecycle reload。
- 演示数据：运行 `go run scripts/seed_data.go` 会删除并重建演示 ID 范围内的数据，不可用于生产。
