# kratos-shop 文档

本项目是基于 [go-kratos](https://github.com/go-kratos/kratos) v2 的微服务商城示例，仓库文档统一放在这里。

## 文档索引

- [架构与微服务关系](architecture.md)：服务拆分、调用链、端口、注册发现与中间件
- [数据库说明](database.md)：PostgreSQL 建表 SQL、连接信息、Mock 数据
- [本地开发指南](development.md)：环境要求、初始化、构建、启动与验证
- [微服务商城演进方案](evolution-plan.md)：库存服务评估、消息队列设计与分阶段执行计划

## 快速导航

| 目录 | 说明 |
| --- | --- |
| `admin/` | 后台管理 HTTP 服务（9099） |
| `shop/` | 商城 BFF HTTP 服务（8097） |
| `service/user/` | 用户 gRPC 服务（50051） |
| `service/goods/` | 商品 gRPC 服务（50052） |
| `service/cart/` | 购物车 gRPC 服务（50053） |
| `service/order/` | 订单 gRPC 服务（50054） |
| `sql/` | PostgreSQL 建表与 Mock 数据 |
| `scripts/` | 数据库初始化等辅助脚本 |

> 注意：`web/` 目录不在本次维护范围内。
