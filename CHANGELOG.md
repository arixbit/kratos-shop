# Changelog

本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。未发布正式版本前，变更记录维护在 GitHub Releases 中。

## [Unreleased]

### 新增

- README：完整的三套启动方式（Docker 一键、本地手动、多服务器独立部署）
- README/docs：可观测性与辅助组件（Prometheus、Grafana、RabbitMQ、Jaeger、Traefik、Loki）使用说明
- 配置：`config.local.yaml` 本地覆盖机制与 `config.local.example.yaml` 模板
- 运维：服务 `/healthz` 健康检查、Consul 健康检查、Compose healthcheck
- 运维：Docker 镜像构建（GHCR）与 Release 工作流
- 运维：Dependabot 依赖更新
- 运维：数据库版本化迁移脚本、备份/恢复脚本
- 运维：Prometheus 告警规则与 Alertmanager
- 部署：Kubernetes Kustomize 清单
- 测试：订单、支付、用户密码等关键单元测试

### 修复

- 移除文档中的本机绝对路径，统一使用项目内 `deploy/docker-compose.yml`
- `scripts/init-db.sh` 默认容器名与 Compose 保持一致
