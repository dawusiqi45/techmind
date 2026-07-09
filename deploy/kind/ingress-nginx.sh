# Nginx Ingress Controller for kind
# 安装后 Ingress 资源才能生效
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

kubectl apply -f "${SCRIPT_DIR}/ingress-nginx-kind.yaml"
