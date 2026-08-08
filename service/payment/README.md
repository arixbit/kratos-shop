# payment 支付服务

支付单创建与回调的 gRPC 服务，目前为本地模拟支付（`channel: mock`），支付成功通过 RabbitMQ 发布事件。

- 端口：`50056`
- 数据库：`shop_payment`
- 服务名（Consul）：`shop.payment.service`

## 本地运行

```bash
make build
./bin/payment -conf configs
```

依赖 PostgreSQL、Consul、RabbitMQ、Jaeger，可通过根目录 `make infra-up` 启动基础设施。

## 配置

配置见 `configs/config.yaml`，敏感字段通过 `config.local.yaml` 覆盖，说明见 [docs/configuration.md](../../docs/configuration.md)。
