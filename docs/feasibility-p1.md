# P1 生产就绪可行性报告

> 本报告记录 kratos-shop 从“本地可运行”走向“可发布、可部署”的 P1 改造决策，供后续维护者参考。P2（前端与商城业务功能扩展）不在本报告范围内。

## 1. 配置：config.local.yaml 本地覆盖

**方案**：保留仓库内 `config.yaml` 作为 Demo 默认值；用户复制 `config.local.example.yaml` 为 `config.local.yaml` 覆盖敏感配置。

**依据**：Kratos `file.NewSource(目录)` 会合并目录内所有非隐藏 YAML，文件名排序靠后的覆盖靠前，因此 `config.local.yaml` 自动覆盖 `config.yaml` / `registry.yaml`，无需引入环境变量机制。

**落地**：
- 16 个配置目录均提供 `config.local.example.yaml`
- `.gitignore` 忽略 `**/config.local.yaml`
- 文档见 [configuration.md](configuration.md)

## 2. 健康检查

**方案**：
- Consul 注册开启 TTL 健康检查（`WithHealthCheck(true)`）
- 每个服务在指标端口（9101 ~ 9108）提供 `/healthz`
- Docker Compose 与 K8s 探针均使用 `/healthz`

**落地**：8 个服务已全部开启；Dockerfile 运行时镜像安装 `curl` 供 healthcheck 使用。

## 3. Docker 镜像与版本发布

**决策**：
- 镜像仓库：GHCR（`ghcr.io/arixbit/kratos-shop/<service>`）
- 多架构：`linux/amd64` + `linux/arm64`（buildx + QEMU）
- 触发方式：推送 `v*` tag 自动构建并发布

**落地**：`.github/workflows/release.yml`、`make build-image`、Dockerfile 支持 `VERSION`/`GOPROXY` 参数。

## 4. CI 与依赖更新

- CI：build + vet + test（PostgreSQL 服务）
- 依赖更新：Dependabot（Go 模块、Docker 基础镜像、GitHub Actions）
- 覆盖率：可后续接入 Codecov 或 Actions 报告

## 5. 数据库版本化迁移

**决策**：使用 golang-migrate 目录结构，按 6 个库拆分。

**落地**：
- `sql/migrations/<db>/000001_init.up.sql` / `.down.sql`
- `scripts/migrate.sh`（优先本机 `migrate` CLI，否则使用 `migrate/migrate` Docker 镜像）
- `make migrate-up` / `make migrate-down`

## 6. 备份与恢复

**落地**：
- `scripts/backup.sh`：按 6 个库 `pg_dump`，默认保留 7 天
- `scripts/restore.sh`：恢复前必须输入 `YES` 确认

## 7. 告警

**落地**：
- Prometheus 规则：应用服务与基础设施下线告警
- Alertmanager 配置模板（webhook / email 示例，接收方需自行填写）

## 8. Kubernetes

**决策**：先提供 Kustomize 基础清单，Helm 作为后续可选演进。

**落地**：
- `deploy/k8s/base/`：8 个 Deployment + Service + Ingress + Namespace
- `scripts/k8s-config.sh`：从 `deploy/configs/` 生成 ConfigMap/Secret
- 文档见 [deploy/k8s/README.md](../deploy/k8s/README.md)

## 9. 测试

新增关键单元测试：

- 用户密码加密/校验（PBKDF2-SHA512）
- 订单创建校验、金额计算、事件 payload、取消/支付幂等
- 支付回调全流程（金额校验、订单状态、重复回调）
- `/healthz` 健康检查
- 订单领域工具方法

## 后续可选项（未包含在本次范围内）

- Helm Chart
- 真实告警渠道（钉钉/企业微信/邮件）
- Secret 管理（Sealed Secrets、External Secrets Operator）
- 镜像漏洞扫描（Trivy）
- 测试覆盖率看板
- 前端与商城业务功能（P2）
