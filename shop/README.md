# shop 商城 BFF

面向 C 端的商城 HTTP 聚合服务（Backend For Frontend），通过 gRPC 服务发现调用 user、goods 等服务，并负责签发 JWT。

已暴露购物车（`/api/cart/*`）、订单（`/api/order/*`）、模拟支付（`/api/payment/*`）HTTP 接口，前端可走 BFF 完成 加购 → 下单 → 模拟支付 → 订单查询 的完整链路。

- HTTP 端口：`8097`
- gRPC 端口：`9001`
- 数据库：`shop_user`（复用用户库）
- 服务名（Consul）：`shop.api`

## 本地运行

```bash
make build
./bin/shop -conf configs
```

依赖 PostgreSQL、Redis、Consul、Jaeger，可通过根目录 `make infra-up` 启动基础设施。

## 配置

配置见 `configs/config.yaml`，敏感字段（含 `auth.jwt_key`）通过 `config.local.yaml` 覆盖，说明见 [docs/configuration.md](../docs/configuration.md)。
