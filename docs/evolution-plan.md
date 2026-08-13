# 微服务商城演进方案：库存服务评估与消息队列落地规划

> 状态：可行性评估 + 执行方案（本期不改代码）
> 适用版本：当前仓库已完成 Go 1.26.5 / Kratos v2.9.2 / PostgreSQL 迁移后的基线

## 1. 背景与目标

当前项目已经可以运行：

- `shop.user.service` / `shop.goods.service` / `shop.cart.service` / `shop.order.service` 四个 gRPC 服务
- `shop.api`（BFF）与 `admin.api`（管理端）两个 HTTP 入口
- Consul 注册发现、PostgreSQL、Redis、Elasticsearch、Jaeger 均已打通

但离“完整可用的微服务商城”还有明显差距。本方案回答三个问题：

1. 是否需要独立的库存服务？
2. 除库存外还缺什么？
3. 如何用消息队列把订单、库存、支付串成一个可靠闭环？

## 2. 现状盘点

### 2.1 已完成

| 模块 | 状态 |
| --- | --- |
| 基础设施 | PostgreSQL 18、Redis 7、Consul、Elasticsearch 7、Jaeger、RabbitMQ 已具备 |
| 用户服务 | 用户 CRUD、密码校验、收货地址可用 |
| 商品服务 | 分类/品牌/类型/规格/属性/商品/SKU 的建表与基础写入可用，ES 搜索有雏形 |
| 购物车服务 | 创建、查询可用 |
| 订单服务 | 能启动，跨服务调用链可跑通 |
| BFF / 管理端 | 登录、注册、地址接口可用 |

### 2.2 主要缺口

| 缺口 | 现状 | 影响 |
| --- | --- | --- |
| 订单没有真正落库 | `CreateOrder` 只做跨服务查询和打印，不写 `orders` 表，也不返回订单号 | 无法形成订单闭环 |
| 订单状态机缺失 | 无待支付/已支付/已发货/已完成/已取消状态流转 | 支付、超时、售后无从谈起 |
| 购物车不完整 | `UpdateCart` / `DeleteCart` / `GetCart` 是注释掉的空实现 | 用户无法改数量、删除商品 |
| 商品 SKU 查询为空 | `GoodsService.SkuList` 返回空响应 | 订单无法校验价格与库存 |
| ES 搜索未闭环 | mapping 写死了 `ik_max_word`，本地 ES 没有 IK 插件；索引初始化与商品索引同步不完整 | 搜索不可用或报错 |
| 没有库存服务 | 库存表 `goods_inventories` 挂在 goods 服务里，只有创建库存，没有扣减/锁定/释放 | 超卖、并发下单无法控制 |
| 没有消息队列接入 | RabbitMQ 容器已在 compose 中，代码零接入 | 无法异步解耦、削峰、可靠事件 |
| 没有支付服务 | 无支付单、回调、对账 | 订单只能停在待支付 |
| 分布式一致性 | 无 outbox、无幂等、无补偿 | 跨服务操作失败后会数据不一致 |
| 运维设施 | 无 Prometheus/Grafana、无 CI/CD、无统一部署编排 | 只适合本地学习 |
| Proto 重复维护 | user/goods/cart proto 在多个服务里复制 | 改接口要同步多处，易出错 |

## 3. 库存服务评估

### 3.1 现状问题

库存目前“长在” goods 服务里：

- `goods_skus.inventory`：SKU 上的库存冗余字段
- `goods_inventories`：库存表，仅有 `Create` 方法
- 没有任何 `TryLock / Deduct / Release / Query` 语义

这带来三个问题：

1. 订单下单要扣库存时，如果直接调 goods 服务，库存和商品元数据耦合，库存的并发锁、流水、对账逻辑会越写越重。
2. 订单、支付、售后都要操作库存，库存本质上是独立业务域，不应由商品服务代理。
3. 后续做秒杀/热库存时，库存需要独立的 Redis 预热、异步落库能力，放在 goods 里会互相干扰。

### 3.2 三个可选方案

#### 方案 A：库存继续放在 goods 服务（最小改动）

优点：

- 不用新增服务，改动最小
- 现有 `goods_inventories` 表可以直接扩展

缺点：

- 库存逻辑与商品元数据耦合
- 订单/支付/售后都依赖商品服务，商品服务会变成“上帝服务”
- 后续拆分成本更高

适合：只想快速验证订单闭环的学习阶段。

#### 方案 B：独立 `service/inventory` 库存服务（推荐）

优点：

- 库存作为独立业务域，拥有自己的数据库（`shop_inventory`）
- 提供明确的库存接口：预占、确认扣减、释放、查询
- 通过 RabbitMQ 消费订单事件，天然适合异步削峰
- 后续可独立扩展 Redis 热库存、库存流水、对账

缺点：

- 需要把现有 `goods_inventories` 表和写入逻辑迁移到新服务
- 需要处理订单-库存之间的一致性（outbox + 补偿）

#### 方案 C：Redis 热库存 + 异步落库（作为 B 的增强）

在高并发场景下，库存扣减先在 Redis 完成，再由库存服务异步写入 PostgreSQL。

优点：

- 抗高并发、支持秒杀

缺点：

- 需要非常严格的对账和补偿机制
- 不适合作为唯一方案，必须与方案 B 的数据库库存配合

### 3.3 评估结论

**建议采用“方案 B 为主 + 方案 C 为后续增强”。**

理由：

- 当前代码量小，库存逻辑还没写复杂，现在拆独立服务成本最低；
- 用户明确要“微服务版本商城”，独立库存服务更符合目标架构；
- 你已经具备 RabbitMQ，库存服务天然应该消费订单事件，而不是被同步调用扣库存。

推荐的新增服务：

| 服务 | 职责 | 端口建议 |
| --- | --- | --- |
| `service/inventory` | 库存查询、预占、确认扣减、释放、库存流水 | 50055 |

## 4. 目标架构

```mermaid
flowchart LR
    U["用户"] --> A["admin.api :9099"]
    U --> S["shop.api :8097"]
    S --> G["API 网关（可选）"]
    A --> G
    G --> US["user :50051"]
    G --> GS["goods :50052"]
    G --> CS["cart :50053"]
    G --> OS["order :50054"]
    G --> IS["inventory :50055"]
    G --> PS["payment :50056（新增）"]

    OS -->|gRPC| US
    OS -->|gRPC| CS
    OS -->|gRPC| GS
    OS -->|gRPC| IS
    IS -->|gRPC| OS
    PS -->|gRPC| OS

    MQ["RabbitMQ :5672"]
    OS -->|outbox 发布事件| MQ
    MQ --> IS
    MQ --> GS
    MQ --> PS
    MQ --> OS

    US --> P1[("PostgreSQL shop_user")]
    GS --> P2[("PostgreSQL shop_goods")]
    CS --> P3[("PostgreSQL shop_cart")]
    OS --> P4[("PostgreSQL shop_order")]
    IS --> P5[("PostgreSQL shop_inventory")]
    PS --> P6[("PostgreSQL shop_payment")]
    IS --> R[("Redis 热库存")]
    GS --> E["Elasticsearch"]
```

## 5. 队列方案：RabbitMQ

### 5.1 为什么选 RabbitMQ

项目自带的 `deploy/docker-compose.yml` 已经包含：

```yaml
rabbitmq:
  image: rabbitmq:3-management-alpine
  environment:
    RABBITMQ_DEFAULT_USER: root
    RABBITMQ_DEFAULT_PASS: root
```

连接地址：

- AMQP：`amqp://root:root@127.0.0.1:5672/`
- 管理台：`http://127.0.0.1:15672`

因此不需要引入 Kafka/RocketMQ 等新组件，直接用 RabbitMQ 即可。

### 5.2 消息拓扑

采用 **Topic Exchange + 独立队列 + 死信队列**：

| Exchange | 类型 | 用途 |
| --- | --- | --- |
| `order.exchange` | topic | 订单生命周期事件 |
| `inventory.exchange` | topic | 库存预占/扣减/释放结果 |
| `payment.exchange` | topic | 支付结果事件 |
| `mall.dlx` | topic | 死信队列，统一重试 |

路由键设计：

```text
order.created     // 订单创建成功
order.paid        // 支付成功
order.cancelled   // 用户取消
order.timeout     // 超时未支付
order.completed   // 交易完成

inventory.locked      // 库存预占成功
inventory.lock.failed // 库存预占失败
inventory.deducted    // 库存已扣减
inventory.released    // 库存已释放

payment.created
payment.success
payment.failed
```

队列与消费关系：

| 队列 | 绑定路由键 | 消费者 |
| --- | --- | --- |
| `q.order.created` | `order.created` | inventory：预占库存 |
| `q.order.paid` | `order.paid` | inventory：确认扣减；goods：更新销量 |
| `q.order.cancel.timeout` | `order.cancelled`、`order.timeout` | inventory：释放库存；order：更新状态 |
| `q.inventory.result` | `inventory.*` | order：更新预占结果 |
| `q.payment.result` | `payment.success`、`payment.failed` | order：更新支付状态 |

### 5.3 消息体设计

统一事件消息：

```json
{
  "event_id": "8f14e45f-ea2b-4b3d-9f6e-2d7c1a0b9e4f",
  "event_type": "order.created",
  "order_sn": "20260808123456789",
  "user_id": 1,
  "occurred_at": "2026-08-08T12:00:00+08:00",
  "trace_id": "trace-xxx",
  "payload": {
    "skus": [
      { "sku_id": 1, "num": 2, "price": 899900 }
    ],
    "address_id": 1
  }
}
```

### 5.4 可靠性设计

1. **Outbox 模式**：订单服务在本地事务里同时写 `orders` 和 `order_event_outbox`，再由 Outbox Relay 发布到 RabbitMQ，避免“先写库再发消息，消息丢失”的双写问题。
2. **消费幂等**：每个消费者维护 `consumed_event(event_id)` 表，重复消息直接跳过。
3. **手动 ACK + 重试**：消费失败不自动确认，进入重试队列；超过 N 次进入死信队列告警。
4. **补偿**：
   - 库存预占失败 → 订单标记取消
   - 支付超时/用户取消 → 发布 `order.timeout` / `order.cancelled` → 库存释放
   - 支付成功但扣减失败 → 告警 + 人工/自动对账
5. **超时关单**：优先使用 RabbitMQ Delayed Message Exchange 插件实现延迟消息；若不想引入插件，用定时任务扫描 `待支付 + 超过 N 分钟` 的订单兜底。

## 6. 核心链路设计

### 6.1 下单主流程

```mermaid
sequenceDiagram
    participant U as 用户
    participant S as shop.api
    participant O as order
    participant C as cart
    participant G as goods
    participant I as inventory
    participant M as RabbitMQ

    U->>S: 提交订单
    S->>O: CreateOrder
    O->>C: 校验购物车
    O->>G: 校验 SKU 与价格
    O->>O: 本地事务：写订单 + 写 outbox
    O-->>M: 发布 order.created
    M->>I: 消费 order.created
    I->>I: 预占库存（Redis + PostgreSQL）
    I-->>M: 发布 inventory.locked / lock.failed
    M->>O: 更新订单状态（待支付 / 已取消）
    O-->>S: 返回 order_sn
```

### 6.2 支付成功流程

```mermaid
sequenceDiagram
    participant P as payment
    participant M as RabbitMQ
    participant I as inventory
    participant G as goods
    participant O as order

    P->>M: 发布 payment.success
    M->>I: 确认扣减库存
    M->>G: 更新销量/热度
    M->>O: 订单置为已支付
```

### 6.3 取消/超时流程

```mermaid
sequenceDiagram
    participant O as order
    participant M as RabbitMQ
    participant I as inventory

    O->>M: 发布 order.cancelled / order.timeout
    M->>I: 释放库存
    I->>M: 发布 inventory.released
    M->>O: 订单置为已取消
```

## 7. 数据模型规划

### 7.1 新增/调整的库

| 数据库 | 说明 |
| --- | --- |
| `shop_inventory`（新增） | 库存服务独立库 |
| `shop_payment`（新增） | 支付服务独立库 |
| `shop_order`（调整） | 增加 outbox 表、事件消费表、状态机字段 |

### 7.2 关键新表

库存服务：

```text
inventories              -- 实时库存（sku_id 唯一）
inventory_locks          -- 预占记录（order_sn + sku_id）
inventory_flows          -- 库存流水（入/出/冻结/释放）
consumed_event           -- 消息消费幂等表
```

订单服务：

```text
orders                   -- 补全状态、金额、支付信息
order_items              -- 订单商品快照（名称/价格/图片/规格）
order_event_outbox       -- 待发布事件
consumed_event           -- 消息消费幂等表
```

## 8. 分阶段执行计划

> 估算按“一个人全职”的口径，实际可按团队调整。

### 当前实施进度（2026-08-08）

| 阶段 | 状态 |
| --- | --- |
| Phase 0：基线 | ✅ 已完成 |
| Phase 1：订单与购物车闭环 | ✅ 已完成（购物车增删改查、SkuList 真实查询、订单落库与查询/列表、下单清购物车、商品更新/删除/上下架、admin/shop 商品代理、商品详情含 SKU/图片、SKU 库存实时来自 inventory 服务、ES 启动自动重建索引并组合分类/品牌/SKU、GormList 扫描修复、RabbitMQ 客户端、端到端测试脚本、基础单元测试） |
| Phase 2：库存服务 + 消息队列 | ✅ 完成（独立 inventory 服务、outbox + 重试上限、消费幂等、Redis 热库存、order.created 预占、order.paid 确认扣减、order.cancelled/超时释放、商品销量同步与幂等、消费者重试退避与 DLQ 死信队列、inventory 数据层集成测试） |
| Phase 3：支付服务 | ✅ 本地模拟完成（创建支付单、金额与订单校验、回调幂等、回调校验订单待支付状态、payment.success → 订单已支付 → 库存确认扣减、payment 数据层测试）；✅ shop BFF 已补齐购物车/下单/模拟支付/订单列表与详情的 HTTP 接口；✅ admin BFF 订单管理（列表/详情/发货/退款）、商品新增、用户列表与地址管理、运营看板统计（用户数/订单量/成交额/近30天趋势），运营后台使用 Ant Design Pro v6（Umi Max 4 + React 19 + antd 6）；✅ `make seed-demo` 可一键生成演示数据；待补：支付宝/微信真实对接、退款库存/支付补偿 |
| Phase 4：工程化与运维 | ✅ 完成（统一 compose、Prometheus/Grafana、应用级 /metrics、Loki/Promtail 日志采集、Traefik API 网关、GitHub Actions CI、shop/admin 限流、JWT 刷新令牌、管理员角色校验）；真实镜像仓库发布需仓库凭据，列为部署增强 |

### Phase 0：基线（已完成）

- 依赖升级、PostgreSQL 迁移、SQL 与 Mock 数据、docs
- 退出标准：6 个服务可运行、可测试

### Phase 1：订单与购物车闭环（预计 5~8 人日）

任务：

- 实现 order 服务真正落库：订单号、金额、状态、收货地址快照
- 实现 `CreateOrder` 返回真实 `order_sn`
- 补齐 cart 的 `UpdateCart` / `DeleteCart` / `GetCart`
- 实现 goods 的 `SkuList` 真实查询（按 SKU ID 返回价格、名称、上下架状态）
- 修复 ES：索引初始化、IK 分词插件或改为标准分词
- 引入公共库：RabbitMQ client、JSON 序列化、trace 透传

退出标准：

- 能创建订单并持久化
- 订单金额与购物车/SKU 价格一致
- 购物车增删改查全部可用

### Phase 2：库存服务 + 消息队列（预计 6~10 人日）

任务：

- 新建 `service/inventory`，迁移 `goods_inventories` 数据
- 实现库存接口：`Query` / `TryLock` / `ConfirmDeduct` / `Release`
- 引入 Redis 热库存（可选先做数据库版）
- order 服务接入 outbox + RabbitMQ
- inventory 消费 `order.created`，发布预占结果
- 实现订单超时/取消的库存释放

退出标准：

- 下单 → 库存预占 → 超时释放全链路可用
- RabbitMQ 管理台能看到 exchange/queue 和消息
- 重复消费不产生重复扣减

### Phase 3：支付服务（预计 5~8 人日）

任务：

- 新建 `service/payment`，先做模拟支付，再对接支付宝/微信
- 支付单创建、回调验签、支付状态事件
- 支付成功后驱动库存确认扣减与订单状态流转

退出标准：

- 完整生命周期：创建订单 → 支付 → 扣库存 → 订单完成
- 取消/超时链路不产生库存泄漏

### Phase 4：工程化与运维（预计 5~8 人日）

任务：

- Prometheus + Grafana 指标采集
- 统一 Docker Compose（当前 infra + 新增服务 + 网关）
- CI/CD（GitHub Actions：lint、test、build、镜像）
- API 网关（Kratos Gateway / APISIX）与统一鉴权
- 公共 proto 模块，消除重复维护

退出标准：

- 一条命令拉起整套环境
- CI 自动构建、测试、发布镜像

### Phase 5（可选）：高并发增强

- Redis 热库存 + 异步对账
- 秒杀场景独立队列与限流
- 搜索服务独立拆分
- 分库分表与分布式 ID

## 9. 风险与应对

| 风险 | 影响 | 应对 |
| --- | --- | --- |
| 订单与库存一致性 | 超卖/库存泄漏 | outbox + 幂等 + 补偿 + 对账任务 |
| RabbitMQ 消息丢失 | 状态不同步 | 手动 ACK、持久化队列、死信队列 |
| ES 分词器不匹配 | 搜索不可用 | 本地装 IK 插件，或先去掉 `ik_max_word` |
| 现有 proto 重复 | 接口不一致 | 抽公共 proto 模块或 buf workspace |
| 订单代码过于骨架 | 队列方案无处落地 | 先完成 Phase 1 的真实落库，再接 MQ |
| 服务越多运维越重 | 本地启动成本高 | 统一 compose + Makefile 一键启动 |

## 10. 可行性结论

**结论：可行，且当前正是做架构演进的最佳时机。**

- 基础设施已经齐备（Postgres、Redis、Consul、ES、Jaeger、RabbitMQ）；
- 服务骨架已经能跑，改造成本主要在业务闭环而不是基础设施；
- 独立库存服务建议现在就规划，越晚拆分成本越高；
- RabbitMQ 是当前最顺的队列选择，无需引入新中间件；
- 总工作量粗估 20~35 人日即可达到“完整可用的微服务商城”水平。
