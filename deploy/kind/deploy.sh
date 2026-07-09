#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
CLUSTER_NAME="techmind"
NAMESPACE="techmind"

pull_image_with_fallback() {
  local target_image="$1"
  shift

  if docker image inspect "${target_image}" >/dev/null 2>&1; then
    return 0
  fi

  for source_image in "$@"; do
    echo "  拉取 ${source_image}"
    if timeout 180 docker pull "${source_image}"; then
      if [ "${source_image}" != "${target_image}" ]; then
        docker tag "${source_image}" "${target_image}"
      fi
      return 0
    fi
    echo "  拉取 ${source_image} 失败，尝试下一个镜像源"
  done

  echo "  无法拉取 ${target_image}，请检查 Docker 网络或手动拉取该镜像" >&2
  return 1
}

preload_registry_k8s_image() {
  local image="$1"

  pull_image_with_fallback "${image}" \
    "m.daocloud.io/${image}" \
    "${image}"

  echo "  注入 ${image} 到 kind 集群"
  kind load docker-image "${image}" --name "${CLUSTER_NAME}"
}

preload_docker_image() {
  local image="$1"
  local candidates=()

  if [[ "${image}" == */* ]]; then
    candidates=(
      "docker.1ms.run/${image}"
      "docker.m.daocloud.io/${image}"
      "docker.xuanyuan.me/${image}"
      "${image}"
    )
  else
    candidates=(
      "docker.1ms.run/library/${image}"
      "docker.m.daocloud.io/library/${image}"
      "docker.xuanyuan.me/library/${image}"
      "docker.1ms.run/${image}"
      "docker.m.daocloud.io/${image}"
      "docker.xuanyuan.me/${image}"
      "${image}"
    )
  fi

  pull_image_with_fallback "${image}" "${candidates[@]}"

  echo "  注入 ${image} 到 kind 集群"
  kind load docker-image "${image}" --name "${CLUSTER_NAME}"
}

echo "=========================================="
echo "  TechMind Kind 集群一键部署"
echo "=========================================="

# ============================================
# 1. 创建 kind 集群
# ============================================
echo ""
echo "[1/8] 创建 kind 集群..."
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
  echo "  集群 ${CLUSTER_NAME} 已存在，跳过"
else
  kind create cluster --name ${CLUSTER_NAME} --config "${SCRIPT_DIR}/cluster.yaml"
  echo "  集群创建完成"
fi

# ============================================
# 2. 安装集群中间件
# ============================================
echo ""
echo "[2/8] 安装集群中间件 (Ingress Controller + Metrics Server)..."

# registry.k8s.io 在部分网络环境不可用，先从可用镜像源拉取并注入 kind 节点。
preload_registry_k8s_image "registry.k8s.io/ingress-nginx/controller:v1.15.1"
preload_registry_k8s_image "registry.k8s.io/ingress-nginx/kube-webhook-certgen:v1.6.9"
preload_registry_k8s_image "registry.k8s.io/metrics-server/metrics-server:v0.7.1"

# Nginx Ingress Controller. Admission jobs are recreated so failed/old jobs do not block re-runs.
if kubectl get ns ingress-nginx >/dev/null 2>&1; then
  kubectl delete job ingress-nginx-admission-create ingress-nginx-admission-patch \
    -n ingress-nginx --ignore-not-found=true
fi
kubectl apply --validate=false -f "${SCRIPT_DIR}/ingress-nginx-kind.yaml"
echo "  等待 Ingress Admission Job 完成..."
kubectl wait --namespace ingress-nginx \
  --for=condition=complete job/ingress-nginx-admission-create \
  --timeout=120s
kubectl wait --namespace ingress-nginx \
  --for=condition=complete job/ingress-nginx-admission-patch \
  --timeout=120s
echo "  等待 Ingress Controller 就绪..."
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=180s

# Metrics Server
kubectl apply --validate=false -f "${SCRIPT_DIR}/metrics-server.yaml"
echo "  等待 Metrics Server 就绪..."
kubectl rollout status deployment/metrics-server -n kube-system --timeout=120s

# ============================================
# 3. 构建后端镜像 (server + worker)
# ============================================
echo ""
echo "[3/8] 构建后端镜像..."
cd "${PROJECT_DIR}"
docker build --target server -t techmind-server:latest .
docker build --target worker -t techmind-worker:latest .
echo "  后端镜像构建完成 (techmind-server, techmind-worker)"

# ============================================
# 4. 构建前端镜像
# ============================================
echo ""
echo "[4/8] 构建前端镜像..."
cd "${PROJECT_DIR}/frontend"
docker build -t techmind-frontend:latest .
echo "  前端镜像构建完成 (techmind-frontend)"

# ============================================
# 5. 加载镜像到 kind 节点
# ============================================
echo ""
echo "[5/8] 加载镜像到 kind 节点..."
kind load docker-image techmind-server:latest --name ${CLUSTER_NAME}
kind load docker-image techmind-worker:latest --name ${CLUSTER_NAME}
kind load docker-image techmind-frontend:latest --name ${CLUSTER_NAME}
preload_docker_image "mysql:8.0"
preload_docker_image "redis:7-alpine"
preload_docker_image "prom/prometheus:v2.55.0"
preload_docker_image "prom/alertmanager:v0.27.0"
preload_docker_image "grafana/grafana:11.2.0"
echo "  镜像加载完成"

# ============================================
# 6. 部署基础设施 (MySQL + Redis + Prometheus + Alertmanager)
# ============================================
echo ""
echo "[6/8] 部署基础设施..."

# Namespace
kubectl apply -f "${SCRIPT_DIR}/namespace.yaml"

# MySQL (含 init.sql ConfigMap)
kubectl create configmap mysql-init-sql \
  --from-file=init.sql="${PROJECT_DIR}/scripts/sql/init.sql" \
  --namespace=${NAMESPACE} \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f "${SCRIPT_DIR}/mysql.yaml"

# Redis
kubectl apply -f "${SCRIPT_DIR}/redis.yaml"

# Prometheus + Alertmanager (ConfigMap + Deployment)
kubectl apply -f "${SCRIPT_DIR}/prometheus-config.yaml"
kubectl apply -f "${SCRIPT_DIR}/alertmanager-config.yaml"
kubectl apply -f "${SCRIPT_DIR}/prometheus.yaml"
kubectl apply -f "${SCRIPT_DIR}/alertmanager.yaml"
kubectl apply -f "${SCRIPT_DIR}/grafana.yaml"

echo "  等待 MySQL 就绪 (可能需要 1-2 分钟)..."
kubectl rollout status statefulset/mysql -n ${NAMESPACE} --timeout=360s
echo "  等待 Redis 就绪..."
kubectl rollout status deployment/redis -n ${NAMESPACE} --timeout=60s

# ============================================
# 7. Helm 部署应用 (server + worker + frontend)
# ============================================
echo ""
echo "[7/8] Helm 部署 TechMind 应用..."
cd "${PROJECT_DIR}"

helm upgrade --install techmind ./deploy/helm/techmind \
  --namespace ${NAMESPACE} \
  -f "${SCRIPT_DIR}/values-kind.yaml"

# ============================================
# 8. 等待就绪
# ============================================
echo ""
echo "[8/8] 等待所有 Pod 就绪..."
kubectl rollout status deployment/techmind-server -n ${NAMESPACE} --timeout=120s
kubectl rollout status deployment/techmind-worker -n ${NAMESPACE} --timeout=120s
kubectl rollout status deployment/techmind-frontend -n ${NAMESPACE} --timeout=120s

echo ""
echo "=========================================="
echo "  部署完成！"
echo "=========================================="
echo ""
echo "  查看所有 Pod："
echo "    kubectl get pods -n ${NAMESPACE}"
echo ""
echo "  访问地址："
echo "    前端：        http://<虚拟机IP>:30000"
echo "    Prometheus：  http://<虚拟机IP>:30909"
echo ""
echo "  后端 API 端口转发："
echo "    kubectl port-forward svc/techmind-server 8080:8080 -n ${NAMESPACE}"
echo ""
echo "  查看日志："
echo "    kubectl logs -f deployment/techmind-server -n ${NAMESPACE}"
echo "    kubectl logs -f deployment/techmind-worker -n ${NAMESPACE}"
echo ""
echo "  删除集群："
echo "    kind delete cluster --name ${CLUSTER_NAME}"
