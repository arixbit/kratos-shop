# kratos-shop
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

### 方式一：一键启动（推荐）

环境需要 Docker Desktop / OrbStack，在项目根目录执行：

```bash
make up
```

该命令通过 `deploy/docker-compose.yml` 一次性启动全部 8 个应用服务以及基础设施（Postgres、Redis、RabbitMQ、Consul、Elasticsearch、Jaeger、Prometheus、Grafana 等）。

停止环境：

```bash
make down
```

### 方式二：手动启动

1. 先启动依赖服务（Postgres、Redis、Consul、Elasticsearch 等），具体命令见 [docs/development.md](docs/development.md)。
2. 初始化数据库与 Mock 数据：

```bash
./scripts/init-db.sh
```

3. 构建全部服务：

```bash
make build
```

4. 分别启动各个服务（每个服务在自己的目录下执行）：

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

启动后可通过以下地址验证：

- 商城 BFF：`http://127.0.0.1:8097`
- 后台管理：`http://127.0.0.1:9099`
- Consul 服务列表：`http://127.0.0.1:8500/ui`

更详细的本地开发说明见 [docs/development.md](docs/development.md)。

---

* 有任何建议，请扫码添加我微信进行交流。

![扫码提建议](https://cdn.jsdelivr.net/gh/aliliin/blog-image@main/uPic/扫码_搜索联合传播样式-白色版.png)

---

> 技术栈、快速开始与更详细的说明见 [docs/README.md](docs/README.md)。
