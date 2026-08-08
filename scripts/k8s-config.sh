#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NS="${NS:-kratos-shop}"
SERVICES=(user goods cart order inventory payment shop admin)

kubectl create namespace "$NS" --dry-run=client -o yaml | kubectl apply -f -

for svc in "${SERVICES[@]}"; do
  kubectl create configmap "$svc-config" -n "$NS" \
    --from-file=config.yaml="$ROOT_DIR/deploy/configs/$svc/config.yaml" \
    --from-file=registry.yaml="$ROOT_DIR/deploy/configs/$svc/registry.yaml" \
    --dry-run=client -o yaml | kubectl apply -f -

  if [ -f "$ROOT_DIR/deploy/configs/$svc/config.local.yaml" ]; then
    kubectl create secret generic "$svc-local-config" -n "$NS" \
      --from-file=config.local.yaml="$ROOT_DIR/deploy/configs/$svc/config.local.yaml" \
      --dry-run=client -o yaml | kubectl apply -f -
    echo "==> $svc: configmap + local secret 已更新"
  else
    echo "==> $svc: configmap 已更新（未发现 config.local.yaml）"
  fi
done
