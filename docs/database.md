# 数据库说明

本项目数据层已从 MySQL 迁移到 PostgreSQL，本地通过 Docker 运行。

## 连接信息

连接信息来自项目自带的 `deploy/docker-compose.yml`（通过 `make infra-up` 启动）：

| 项目 | 值 |
| --- | --- |
| PostgreSQL | `127.0.0.1:5432` |
| 用户 | `postgres` |
| 密码 | `root` |
| Redis | `127.0.0.1:6379` |
| Redis 密码 | `root` |

六个业务库：`shop_user`、`shop_goods`、`shop_cart`、`shop_order`、`shop_inventory`、`shop_payment`。

## 初始化数据库

先确保 Postgres 已启动（项目自带的 Compose 方案）：

```bash
make infra-up
```

然后在项目根目录执行：

```bash
chmod +x scripts/init-db.sh
./scripts/init-db.sh
```

脚本会按顺序执行：

1. `sql/01_init_databases.sql`：创建 6 个数据库
2. `sql/02_user.sql`：用户表 + Mock 数据
3. `sql/03_goods.sql`：商品域全部表 + Mock 数据
4. `sql/04_cart.sql`：购物车表 + Mock 数据
5. `sql/05_order.sql`：订单域表 + Mock 数据
6. `sql/06_inventory.sql`：库存域表 + Mock 数据
7. `sql/07_payment.sql`：支付表 + Mock 数据

> 如果使用 `make up` 一键启动整套环境，Postgres 首次启动时会自动执行 `sql/` 下的初始化脚本，无需手动执行本步骤；手动重复执行也是安全的（脚本幂等）。

也可以手动执行：

```bash
psql -h 127.0.0.1 -U postgres -f sql/01_init_databases.sql
psql -h 127.0.0.1 -U postgres -d shop_user -f sql/02_user.sql
# ...其余文件同理
```

## 各服务配置中的 DSN

```yaml
data:
  database:
    driver: postgres
    source: postgres://postgres:root@127.0.0.1:5432/shop_user?sslmode=disable&TimeZone=Asia/Shanghai
  redis:
    addr: 127.0.0.1:6379
    password: root
```

goods 服务额外依赖：

```yaml
data:
  elastic:
    addr: http://127.0.0.1:9200
```

## Mock 数据

SQL 中已内置以下基础数据：

- 用户：`13800138000`（普通用户）、`13501167215`（管理员），密码均为 `12345678`
- 收货地址：每个用户各 1~2 条
- 商品域：品牌、分类、类型、规格、属性、3 个商品、4 个 SKU、库存与图片
- 购物车：用户 1 两条购物车记录
- 订单：用户 1 一条已完成支付的示例订单

## 表清单

| 数据库 | 表 |
| --- | --- |
| `shop_user` | `users`、`user_addresses` |
| `shop_goods` | `brands`、`categories`、`goods_category_brands`、`goods_types`、`goods_type_brands`、`specifications_attrs`、`specifications_attr_values`、`goods_attr_groups`、`goods_attrs`、`goods_attr_values`、`goods`、`goods_skus`、`goods_specification_skus`、`goods_inventories`、`goods_images` |
| `shop_cart` | `shop_carts` |
| `shop_order` | `orders`、`order_address`、`order_goods`、`order_pays` |
| `shop_inventory` | `inventories`、`inventory_locks`、`inventory_flows`、`consumed_event` |
| `shop_payment` | `payments` |
