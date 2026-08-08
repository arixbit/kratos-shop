# goods 商品服务

商品、SKU、分类、品牌、规格与 Elasticsearch 搜索的 gRPC 服务。

- 端口：`50052`
- 数据库：`shop_goods`
- 服务名（Consul）：`shop.goods.service`

## 本地运行

```bash
make build
./bin/goods -conf configs
```

依赖 PostgreSQL、Redis、Consul、Elasticsearch、RabbitMQ、Jaeger，可通过根目录 `make infra-up` 启动基础设施。

## 配置

配置见 `configs/config.yaml`，敏感字段通过 `config.local.yaml` 覆盖，说明见 [docs/configuration.md](../../docs/configuration.md)。
