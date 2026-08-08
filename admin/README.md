# admin 后台管理服务

后台管理 HTTP 服务，提供登录、注册、用户与地址管理接口，管理员权限通过 `AuthorityId=2` 控制。

- HTTP 端口：`9099`
- 数据库：`shop_user`（复用用户库）
- 服务名（Consul）：`admin.api`

## 本地运行

```bash
make build
./bin/admin -conf configs
```

依赖 PostgreSQL、Redis、Consul、Jaeger，可通过根目录 `make infra-up` 启动基础设施。

## 配置

配置见 `configs/config.yaml`，敏感字段（含 `auth.jwt_key`）通过 `config.local.yaml` 覆盖，说明见 [docs/configuration.md](../docs/configuration.md)。
