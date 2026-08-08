# 安全说明

## 报告漏洞

kratos-shop 是教学/示例性质的开源项目，但安全问题同样重要。

如果发现漏洞，请**不要**直接在 Issue 中公开细节，而是：

1. 在仓库页面打开 **Security → Report a vulnerability**（GitHub Security Advisory）；
2. 或发送邮件至维护者邮箱（见 GitHub 个人主页），主题注明 `[SECURITY]`；
3. 描述：影响范围、复现步骤、影响与修复建议（如已知）。

我们会在确认后尽快处理，并在修复发布前与报告者协调披露时间。

## 默认凭据提醒

仓库内 `config.yaml` 中的数据库、Redis、RabbitMQ、Grafana 等凭据仅用于本地演示。任何公网部署前必须通过 `config.local.yaml`、Kubernetes Secret 或部署时生成配置替换默认值，具体见 [docs/configuration.md](docs/configuration.md)。
