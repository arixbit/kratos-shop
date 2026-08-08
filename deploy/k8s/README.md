# Kubernetes 部署（Kustomize）

本目录提供一套最小可用的 Kustomize 清单，用于把 8 个应用服务部署到 Kubernetes。

## 前提

- 已发布镜像（`ghcr.io/arixbit/kratos-shop/<service>:<tag>`），可通过仓库的 Release 工作流构建，或本地 `make build-image` 后自行推送。
- 集群内可访问外部依赖：PostgreSQL、Redis、RabbitMQ、Consul、Elasticsearch、Jaeger。依赖地址通过配置注入（见下文），本项目默认不把基础设施打包进集群。
- 已安装 `kubectl`。

## 部署步骤

1. 创建命名空间与配置：

```bash
./scripts/k8s-config.sh
```

脚本会把 `deploy/configs/<service>/` 下的 `config.yaml`、`registry.yaml` 写入 ConfigMap。如果存在 `config.local.yaml`，会写入同名 Secret，Deployment 会自动挂载到 `/data/conf/config.local.yaml`，由 Kratos 合并覆盖。

2. 应用清单：

```bash
kubectl apply -k deploy/k8s/base
```

3. 查看状态：

```bash
kubectl -n kratos-shop get pods
kubectl -n kratos-shop get svc
```

4. 验证：

- Consul UI 中应能看到 8 个服务注册；
- `kubectl -n kratos-shop port-forward svc/shop 8097:8097` 后访问 `http://127.0.0.1:8097`；
- Pod 就绪探针依赖各服务 `/healthz`（91xx 端口）。

## 访问入口

Ingress 使用 `traefik` IngressClass：

- 商城 BFF：`http://shop.local`（需要把 `shop.local` 解析到集群入口）
- 后台管理：`http://admin.local`

也可以去掉 Ingress，改用 `kubectl port-forward` 或 LoadBalancer。

## 覆盖生产配置

不要直接修改 `deploy/configs/` 下的 Demo 配置。在 `deploy/configs/<service>/config.local.yaml` 中填写真实凭据后重新执行：

```bash
./scripts/k8s-config.sh
kubectl -n kratos-shop rollout restart deployment --all
```

Secret 由本机 `kubectl` 创建，不会进入 Git；生产环境建议改用外部密钥管理（如 Sealed Secrets、External Secrets Operator）。
