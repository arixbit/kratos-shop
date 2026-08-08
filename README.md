# kratos-shop

基于 [go-kratos](https://github.com/go-kratos/kratos) v2 的微服务商城示例项目。

## 技术栈

- Go 1.26.5
- Kratos v2.9.2
- PostgreSQL 18
- Redis 7
- Consul 服务注册/发现
- Elasticsearch 商品搜索
- Jaeger 链路追踪
- GORM + wire

## 目录结构

| 目录 | 说明 |
| --- | --- |
| `service/user` | 用户 gRPC 服务 |
| `service/goods` | 商品 gRPC 服务 |
| `service/cart` | 购物车 gRPC 服务 |
| `service/order` | 订单 gRPC 服务 |
| `service/inventory` | 库存 gRPC 服务 |
| `service/payment` | 支付 gRPC 服务（本地模拟） |
| `shop` | 商城 BFF HTTP 服务 |
| `admin` | 后台管理 HTTP 服务 |
| `sql` | PostgreSQL 建表与 Mock 数据 |
| `scripts` | 数据库初始化脚本 |
| `docs` | 架构与使用文档 |

## 快速开始

```bash
# 1. 启动基础设施（Postgres/Redis 来自 /Users/arix/src/infra/docker-compose.yml）
cd /Users/arix/src/infra
docker compose up -d postgres redis

# 2. 初始化数据库与 Mock 数据
cd /Users/arix/src/Personal/kratos-shop
./scripts/init-db.sh

# 3. 构建并启动全部服务
make build

# 或者一键拉起整套环境（含基础设施、Prometheus、Grafana）
make up
```

详细说明见 [docs](docs/README.md)。
