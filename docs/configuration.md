# 配置说明

本项目使用 Kratos 原生 `file.NewSource` 读取配置：启动时通过 `-conf` 指定配置目录（例如 `./bin/user -conf configs`），Kratos 会读取该目录下所有非隐藏的 YAML 文件并自动合并。

## Demo 配置与本地覆盖

仓库内提交的 `config.yaml` 是开箱即用的 Demo 配置（数据库、Redis、RabbitMQ 等均为本地默认凭据），clone 下来不需要任何额外配置即可启动。

真实环境（服务器、生产）不要直接修改并提交 `config.yaml`，而是使用本地覆盖文件：

1. 在对应配置目录下复制模板：

```bash
cp service/user/configs/config.local.example.yaml service/user/configs/config.local.yaml
```

2. 按需修改 `config.local.yaml` 中的数据库、Redis、JWT、Consul 等配置。

3. `config.local.yaml` 已加入 `.gitignore`，不会提交到仓库。Kratos 按文件名排序合并配置，`config.local.yaml` 排在 `config.yaml` / `registry.yaml` 之后，因此其中的同名配置会覆盖默认值。

## 可覆盖的敏感字段

每个服务的 `configs/` 目录都提供了 `config.local.example.yaml` 模板，常见字段如下：

| 字段 | 说明 |
| --- | --- |
| `data.database.source` | PostgreSQL DSN，包含用户名密码 |
| `data.redis.password` | Redis 密码 |
| `mq.addr` | RabbitMQ AMQP 地址 |
| `auth.jwt_key` | JWT 签名密钥（order/shop/admin） |
| `trace.endpoint` | Jaeger 上报地址 |
| `consul.address` / `consul.scheme` | Consul 服务发现地址 |

## Docker Compose

Compose 会把 `deploy/configs/<service>/` 整个目录挂载到容器 `/data/conf`。因此在本地 `deploy/configs/<service>/` 下创建 `config.local.yaml`，容器内的 Kratos 同样会自动合并，无需修改 compose 文件或重新构建镜像：

```bash
cp deploy/configs/user/config.local.example.yaml deploy/configs/user/config.local.yaml
# 修改 deploy/configs/user/config.local.yaml 后重启对应容器
docker compose -f deploy/docker-compose.yml restart user
```

> 注意：Compose 内部服务使用容器名互访（如 `postgres`、`redis`、`consul`、`jaeger`、`rabbitmq`），覆盖配置时请保持这些主机名，或改为实际可达的外部地址。

## Kubernetes

K8s 部署时把基础配置放入 ConfigMap，把真实凭据放入 Secret，两者都挂载到同一配置目录（例如 `/data/conf`），文件命名为 `config.local.yaml` 即可被 Kratos 自动合并。具体示例见 [K8s 部署说明](deploy/k8s/README.md)。

## 安全提醒

- 仓库内默认凭据仅用于本地演示，公网部署前必须通过 `config.local.yaml`、ConfigMap/Secret 或部署时生成的配置文件替换。
- 不要提交 `config.local.yaml`、`.env` 等包含真实凭据的文件。
- 生产环境建议使用密钥管理工具生成配置，而不是手工维护明文文件。
