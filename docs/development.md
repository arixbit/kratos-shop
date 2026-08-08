# 本地开发指南

## 环境要求

- Go 1.26.5（与本机保持一致）
- Docker Desktop / OrbStack
- PostgreSQL 18（Docker）
- Redis 7（Docker）
- Consul（服务发现）
- Elasticsearch（goods 服务搜索）
- Jaeger（可选，链路追踪）

## 1. 启动基础设施

项目核心依赖的 Postgres/Redis 定义在 `/Users/arix/src/infra/docker-compose.yml`：

```bash
cd /Users/arix/src/infra
docker compose up -d postgres redis
```

还需要启动服务发现与商品搜索依赖：

```bash
docker run -d --name consul -p 8500:8500 hashicorp/consul:latest agent -dev -client=0.0.0.0
docker run -d --name elasticsearch -p 9200:9200 \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  docker.elastic.co/elasticsearch/elasticsearch:7.17.22
docker run -d --name jaeger -p 14268:14268 -p 16686:16686 jaegertracing/all-in-one:latest
```

> Elasticsearch 镜像中如启用了 IK 分词插件，请替换为包含 `ik_max_word` 分词器的镜像。

## 2. 初始化数据库

```bash
./scripts/init-db.sh
```

详细说明见 [database.md](database.md)。

## 3. 构建全部服务

```bash
make build
```

或分别构建：

```bash
cd service/user && go build -o bin/user ./cmd/user
cd service/goods && go build -o bin/goods ./cmd/goods
cd service/cart && go build -o bin/cart ./cmd/cart
cd service/order && go build -o bin/order ./cmd/order
cd shop && go build -o bin/shop ./cmd/shop
cd admin && go build -o bin/admin ./cmd/admin
```

## 4. 启动服务

每个服务从自己的目录启动，默认读取 `configs/`：

```bash
cd service/user && ./bin/user -conf configs
cd service/goods && ./bin/goods -conf configs
cd service/cart && ./bin/cart -conf configs
cd service/order && ./bin/order -conf configs
cd service/inventory && ./bin/inventory -conf configs
cd service/payment && ./bin/payment -conf configs
cd shop && ./bin/shop -conf configs
cd admin && ./bin/admin -conf configs
```

> 实际运行时建议分别开 8 个终端，先启动 6 个 gRPC 服务，再启动 `shop` 与 `admin`。

## 5. 验证

检查 Consul UI：<http://127.0.0.1:8500/ui>，应能看到 8 个应用服务注册。

通过 HTTP 验证商城 BFF：

```bash
curl -X POST http://127.0.0.1:8097/api/users/login \
  -H 'Content-Type: application/json' \
  -d '{"mobile":"13800138000","password":"12345678"}'
```

通过 gRPC 工具验证用户服务：

```bash
grpcurl -plaintext -d '{"mobile":"13800138000"}' \
  127.0.0.1:50051 user.v1.User/GetUserByMobile
```

### 一键启动整套环境

项目提供统一编排 `deploy/docker-compose.yml`：

```bash
make up
```

包含全部 8 个应用服务与基础设施（Postgres、Redis、RabbitMQ、Consul、Elasticsearch、Jaeger），以及：

- Prometheus：<http://127.0.0.1:9090>
- Grafana：<http://127.0.0.1:3000>（admin / admin）
- Loki：<http://127.0.0.1:3100>（日志存储，Grafana 已配置数据源）
- RabbitMQ 管理台：<http://127.0.0.1:15672>（root / root）
- Consul：<http://127.0.0.1:8500>
- Jaeger：<http://127.0.0.1:16686>
- API 网关（Traefik）：<http://localhost:8080>（Dashboard）
- 服务指标：各服务 `http://127.0.0.1:91XX/metrics`（user 9101 ~ admin 9108）
- 日志：Promtail 自动采集全部容器日志写入 Loki，可在 Grafana Explore 中按 `container` / `service` 查询

> 注意：`make up` 会占用 5432/6379/5672/8500/9200 等端口；如果本机已经用 `/Users/arix/src/infra/docker-compose.yml` 启动了相同服务，请先停掉避免端口冲突。

通过网关访问：

- 商城 BFF：`http://localhost/api/users/login`
- 管理端：`http://admin.localhost/api/users/login`（需在 `/etc/hosts` 加 `127.0.0.1 admin.localhost`）

### 端到端测试

服务全部启动后，可运行自动化端到端测试（覆盖下单 → 支付 → 扣库存 → 加销量 → 取消释放）：

```bash
make e2e
```

脚本会创建随机测试用户与购物车数据，全部通过后输出 `E2E 全部通过`。

### 接口限流

`shop` 与 `admin` 已内置基于 IP 的令牌桶限流（默认 20 req/s，突发 40），超限返回 `429 RATE_LIMITED`，可用于防刷。

### JWT 刷新与管理员权限

- `POST /api/users/refresh`：用旧 token 换取新的 30 天 token（shop/admin 均支持）
- admin 除登录/注册/验证码/刷新外，其余接口仅允许 `AuthorityId=2` 的管理员访问，普通用户返回 `403 ADMIN_ONLY`

## 6. 依赖升级说明

全部模块统一到：

- Go `1.26.5`
- `github.com/go-kratos/kratos/v2` 最新稳定版
- gRPC / Protobuf / OpenTelemetry / GORM 等依赖均为当前最新版本

如需重新整理依赖：

```bash
cd service/user
GOPROXY=https://proxy.golang.org,direct go mod tidy
```

## 7. 常见问题

### Postgres 连不上

确认 `infra` 的 compose 已启动且 5432 端口未被占用：

```bash
docker ps | grep postgres
```

### Consul 未启动

服务启动时会注册到 Consul，若 Consul 未启动，`app.Run()` 会失败，请先启动 Consul。

### goods 服务启动失败

goods 启动时会连接 Elasticsearch，请确保 9200 可用；也可以暂时把 `configs/config.yaml` 中 elastic 地址改到可用的实例。
