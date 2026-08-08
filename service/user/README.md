# user 用户服务

用户、密码校验与收货地址管理的 gRPC 服务。

- 端口：`50051`
- 数据库：`shop_user`
- 服务名（Consul）：`shop.user.service`

## 本地运行

```bash
make build
./bin/user -conf configs
```

依赖 PostgreSQL、Redis、Consul、Jaeger，可通过根目录 `make infra-up` 启动基础设施。

## 配置

配置见 `configs/config.yaml`，敏感字段通过 `config.local.yaml` 覆盖，说明见 [docs/configuration.md](../../docs/configuration.md)。
