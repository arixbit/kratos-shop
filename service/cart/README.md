# cart 购物车服务

购物车增删改查的 gRPC 服务。

- 端口：`50053`
- 数据库：`shop_cart`
- 服务名（Consul）：`shop.cart.service`

## 本地运行

```bash
make build
./bin/cart -conf configs
```

依赖 PostgreSQL、Redis、Consul、Jaeger，可通过根目录 `make infra-up` 启动基础设施。

## 配置

配置见 `configs/config.yaml`，敏感字段通过 `config.local.yaml` 覆盖，说明见 [docs/configuration.md](../../docs/configuration.md)。
