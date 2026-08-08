# inventory 库存服务

库存查询、预占、确认扣减与释放的 gRPC 服务，通过 RabbitMQ 消费订单事件。

- 端口：`50055`
- 数据库：`shop_inventory`
- 服务名（Consul）：`shop.inventory.service`

## 本地运行

```bash
make build
./bin/inventory -conf configs
```

依赖 PostgreSQL、Redis、Consul、RabbitMQ、Jaeger，可通过根目录 `make infra-up` 启动基础设施。

## 配置

配置见 `configs/config.yaml`，敏感字段通过 `config.local.yaml` 覆盖，说明见 [docs/configuration.md](../../docs/configuration.md)。
