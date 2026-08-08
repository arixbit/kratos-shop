# 贡献指南

感谢你愿意参与 kratos-shop。本指南帮助你了解如何提交 Issue、发起 Pull Request 以及通过 CI 检查。

## 报告问题

- 使用 GitHub Issues 提交 bug 或功能建议，模板会自动给出填写项。
- Bug 报告请包含：复现步骤、期望行为、实际行为、运行环境（Go 版本、Docker 版本、操作系统）。
- 涉及敏感信息（数据库密码、JWT 密钥等）请勿写入 Issue，走 [SECURITY.md](SECURITY.md) 的流程。

## 开发环境

- Go 1.26.5+
- Docker Desktop / OrbStack
- 本地启动方式见 [README](README.md)「服务启动」与 [docs/development.md](docs/development.md)

## 分支与提交

- 功能开发请基于 `main` 创建分支，分支名建议：`feat/xxx`、`fix/xxx`、`docs/xxx`。
- 提交信息使用简洁的祈使句，例如 `feat: add xxx`、`fix: correct xxx`、`docs: update xxx`。
- 每个提交尽量只做一件事，方便 review 和回滚。

## 代码规范

- 所有 Go 代码必须通过 `make vet`。
- 修改前先跑 `make test`（需要本地或 CI 的 PostgreSQL）。
- 涉及 proto/wire 的改动，提交生成后的代码（`.pb.go`、`wire_gen.go`），并保持 `make api` / `make wire` 可重复生成。
- 涉及配置的改动，同步更新对应的 `config.local.example.yaml` 与文档。

## 提交 Pull Request

1. 确保 CI（build + vet + test）通过。
2. PR 标题概括改动，描述中说明动机、改动点和验证方式。
3. 如果改动影响部署，补充说明迁移或配置步骤。
4. 维护者 review 后会合并或提出修改意见。

## 新增测试

- 核心业务逻辑（订单、支付、库存、用户密码等）必须有单元测试。
- 依赖数据库的测试沿用现有 ginkgo 套件，CI 会启动 PostgreSQL 执行。
