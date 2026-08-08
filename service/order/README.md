# order 订单服务

订单创建、取消、支付状态流转与消息事件发布的 gRPC 服务。

- 端口：`50054`
- 数据库：`shop_order`
- 服务名（Consul）：`shop.order.service`

## 本地运行

```bash
make build
./bin/order -conf configs
```

依赖 PostgreSQL、Redis、Consul、RabbitMQ、Jaeger，可通过根目录 `make infra-up` 启动基础设施。

## 配置

配置见 `configs/config.yaml`，敏感字段通过 `config.local.yaml` 覆盖，说明见 [docs/configuration.md](../../docs/configuration.md)。
