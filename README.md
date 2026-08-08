# kratos-shop
[![CI](https://github.com/arixbit/kratos-shop/actions/workflows/ci.yml/badge.svg)](https://github.com/arixbit/kratos-shop/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.26.5-00ADD8)](https://go.dev)
[![License: MIT](https://img.shields.io/github/license/arixbit/kratos-shop)](LICENSE)

kratos 框架写商品微服务

本项目是一个使用 Kratos 框架创建的很简单的微服务商城项目。
> 注: 本项目中但凡 kratos 提供包,就不会自己封装第三方的包。

主要是为了学习 kratos 如何使用,尤其各种中间件之间的调用,包括微服务的一些技术点。

项目具体目录结构初步设计如下:

```
|-- kratos-shop
    |-- service
        |-- user // 用户服务 grpc
        |-- goods // 商品服务 grpc
        |-- cart // 购物车服务 grpc
        |-- order // 订单服务 grpc
        |-- inventory // 库存服务服务 grpc
    |-- shop // shop 商城服务 http (后期会考虑把订单单独拆出来)
        ├── api  // 商城 api
        │   ├── service
        │   │   └── user
        │   │       └── v1 // 用户服务的 proto
        │   │   └── goods
        │   │       └── v1 // 商品服务的 proto
        │   │
        │   └── shop
        │       └── v1
        │           ├── error_reason.proto
        │           ├── shop.proto
        │── cmd
        │── internal
        │.....
    |-- admin // 后端管理系统 web
```

## 服务启动

项目一共包含 8 个应用服务：user、goods、cart、order、inventory、payment、shop、admin。下面介绍三种启动方式，可以按场景选择。

### 方式一：Docker 一键启动（本地最省事，推荐）

环境需要 Docker Desktop / OrbStack，在项目根目录执行：

```bash
make up
```

等价于：

```bash
docker compose -f deploy/docker-compose.yml up -d --build
```

该命令会一次性启动全部 8 个应用服务，以及它们依赖的基础设施（Postgres、Redis、RabbitMQ、Consul、Elasticsearch、Jaeger）和可观测性组件（Prometheus、Grafana、Loki、Promtail、Traefik 网关）。

启动完成后可访问：

| 组件 | 地址 |
| --- | --- |
| 商城 BFF | http://127.0.0.1:8097 |
| 后台管理 | http://127.0.0.1:9099 |
| Consul 服务列表 | http://127.0.0.1:8500/ui |
| Prometheus | http://127.0.0.1:9090 |
| Grafana | http://127.0.0.1:3000（admin / admin） |
| RabbitMQ 管理台 | http://127.0.0.1:15672（root / root） |
| Jaeger | http://127.0.0.1:16686 |
| Traefik Dashboard | http://localhost:8080 |

停止并清理环境：

```bash
make down
```

> 注意：`make up` 会占用 5432/6379/5672/8500/9200/80 等端口，如果本机已经有服务占用这些端口，请先停掉。

### 方式二：本地手动启动（适合开发调试）

手动方式适合需要看单个服务日志、打断点调试的场景。需要 Go 1.26.5 和 Docker。

1. 只启动基础设施：

```bash
make infra-up
```

等价于：

```bash
docker compose -f deploy/docker-compose.yml up -d postgres redis rabbitmq consul elasticsearch jaeger
```

2. 初始化数据库与 Mock 数据（脚本幂等，可重复执行）：

```bash
./scripts/init-db.sh
```

3. 构建全部服务：

```bash
make build
```

4. 分别启动各个服务。建议每个服务开一个终端，先启动 6 个 gRPC 服务，再启动 `shop` 和 `admin`：

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

5. 验证服务是否正常：

```bash
curl -X POST http://127.0.0.1:8097/api/users/login \
  -H 'Content-Type: application/json' \
  -d '{"mobile":"13800138000","password":"12345678"}'
```

也可以在 Consul UI（http://127.0.0.1:8500/ui）确认 8 个服务都已注册。

停止基础设施：

```bash
make infra-down
```

### 方式三：微服务独立部署（多服务器 / 生产环境）

微服务拆分的目的就是可以按服务独立部署：每个服务都是一个独立二进制，可以运行在不同的服务器上，通过同一个 Consul 互相发现和调用。

1. 构建某个服务：

```bash
cd service/user
make build
```

如果目标服务器不是当前系统，可以先交叉编译：

```bash
cd service/user
GOOS=linux GOARCH=amd64 make build
```

2. 把 `bin/user` 和 `configs/` 目录一起拷贝到目标服务器。

3. 修改目标服务器上的配置：

- `configs/config.yaml`：把数据库、Redis、RabbitMQ、Elasticsearch、Jaeger 地址改成实际可达的地址（不要用 `127.0.0.1`，除非依赖就在本机）
- `configs/registry.yaml`：把 `consul.address` 改成 Consul 所在服务器的地址，所有服务必须注册到同一个 Consul

4. 启动服务：

```bash
./bin/user -conf configs
```

其他服务（goods、cart、order、inventory、payment、shop、admin）按同样方式部署。

5. 验证：在 Consul UI（http://<consul-ip>:8500/ui）里应能看到已注册的服务。服务之间通过 `discovery:///shop.xxx.service` 调用，只要 Consul 配置一致，服务跑在哪台机器上不影响调用关系。

各服务默认端口：

| 服务 | 端口 |
| --- | --- |
| user | 50051 |
| goods | 50052 |
| cart | 50053 |
| order | 50054 |
| inventory | 50055 |
| payment | 50056 |
| shop | 8097（HTTP）/ 9001（gRPC） |
| admin | 9099 |

更详细的多服务器部署说明见 [docs/development.md](docs/development.md)。

## 可观测性与辅助组件

`make up` 启动的整套环境里，除了 8 个应用服务，还包含以下组件：

| 组件 | 作用 | 访问地址 | 默认账号 |
| --- | --- | --- | --- |
| Prometheus | 采集 8 个服务及 Postgres/Redis/Consul 的指标 | http://127.0.0.1:9090 | 无 |
| Grafana | 指标与日志可视化，已预置 Prometheus/Loki 数据源 | http://127.0.0.1:3000 | admin / admin |
| RabbitMQ | 订单/支付/库存/商品之间的异步消息队列 | http://127.0.0.1:15672 | root / root |
| Jaeger | 分布式链路追踪，查看请求经过哪些服务 | http://127.0.0.1:16686 | 无 |
| Traefik | API 网关，统一入口转发到 shop 与 admin | http://localhost:8080 | 无 |
| Loki + Promtail | 容器日志采集与存储，在 Grafana 中查询 | 通过 Grafana Explore 使用 | 无 |

- Prometheus：已自动抓取 8 个服务的 `/metrics`（user 9101 ~ admin 9108）以及 Postgres、Redis、Consul，可在 Status → Targets 确认抓取正常。
- Grafana：预置了 Prometheus 和 Loki 两个数据源，登录后可在 Explore 中查指标或日志。
- RabbitMQ：order/payment 发布消息，goods/inventory/order 消费消息；管理台可查看队列、连接和消息积压。
- Jaeger：各服务通过 Collector（14268）上报 trace，UI 中按 Service 搜索即可查看调用链。
- Traefik：`localhost` 路由到商城 BFF，`admin.localhost` 路由到后台管理（需在 `/etc/hosts` 添加 `127.0.0.1 admin.localhost`）。
- Loki + Promtail：自动采集所有容器日志，Grafana Explore 中选择 Loki 数据源，可按 `container` 或 `service` 标签过滤。

> 默认账号仅适用于本地环境，公网部署请务必修改密码并限制管理端口暴露。

更详细的组件说明与验证方法见 [docs/development.md](docs/development.md)。

## 演示账号

- 普通用户：`13800138000` / `12345678`
- 管理员：`13501167215` / `12345678`

> 以上账号由 `sql/` 初始化脚本写入，仅用于本地演示。

## 支付说明

`service/payment` 目前为本地模拟支付（`channel: mock`），不会发起真实扣款，适合学习和联调。如需接入真实渠道（支付宝、微信支付、Stripe 等），需要自行扩展支付服务。

## 更多文档

- [配置说明](docs/configuration.md)：Demo 配置与 `config.local.yaml` 本地覆盖机制
- [架构与微服务关系](docs/architecture.md)：架构图、服务职责、部署形态
- [数据库说明](docs/database.md)：连接信息、初始化与迁移
- [本地开发指南](docs/development.md)：环境要求、启动、验证、可观测性
- [Kubernetes 部署](deploy/k8s/README.md)：Kustomize 清单与部署步骤
- [OpenAPI 接口文档](shop/openapi.yaml)：各服务 `openapi.yaml` 可用 Swagger UI / Redoc 查看，例如 `npx @redocly/cli preview-docs shop/openapi.yaml`

---

* 有任何建议，请扫码添加我微信进行交流。

![扫码提建议](https://cdn.jsdelivr.net/gh/aliliin/blog-image@main/uPic/扫码_搜索联合传播样式-白色版.png)

---

> 技术栈、快速开始与更详细的说明见 [docs/README.md](docs/README.md)。
