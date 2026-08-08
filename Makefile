.PHONY: help build test vet test-cover tidy db-init migrate-up migrate-down up down ps logs e2e infra-up infra-down build-image site-openapi

MODULES := admin service/user service/goods service/cart service/order service/inventory service/payment shop
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo unknown)
GOPROXY ?= https://proxy.golang.org,direct
IMAGE_REGISTRY ?= ghcr.io/arixbit/kratos-shop

help:
	@echo "Usage:"
	@echo "  make build    构建全部服务"
	@echo "  make test     运行全部服务的 Go 测试"
	@echo "  make vet      运行 go vet 静态检查"
	@echo "  make test-cover 运行全部测试并输出覆盖率"
	@echo "  make tidy     整理全部模块依赖"
	@echo "  make db-init  初始化 PostgreSQL 数据库与 Mock 数据"
	@echo "  make migrate-up  执行全部数据库迁移（golang-migrate）"
	@echo "  make migrate-down 回滚全部数据库迁移"
	@echo "  make up       一键启动整套环境（deploy/docker-compose.yml）"
	@echo "  make down     停止整套环境"
	@echo "  make ps       查看整套环境容器状态"
	@echo "  make logs     查看整套环境日志"
	@echo "  make e2e      运行端到端测试（需服务已启动）"
	@echo "  make infra-up   只启动基础设施（Postgres/Redis/RabbitMQ/Consul/ES/Jaeger）"
	@echo "  make infra-down 停止基础设施容器"
	@echo "  make build-image 构建全部服务的 Docker 镜像（GHCR 命名）"
	@echo "  make site-openapi 同步最新 OpenAPI 到 site/openapi（GitHub Pages 使用）"

build:
	@for d in $(MODULES); do \
		echo "==> build $$d"; \
		$(MAKE) -C $$d build || exit 1; \
	done

test:
	@for d in $(MODULES); do \
		echo "==> test $$d"; \
		(cd $$d && go test ./... ) || exit 1; \
	done

vet:
	@for d in $(MODULES); do \
		echo "==> vet $$d"; \
		(cd $$d && go vet ./...) || exit 1; \
	done

test-cover:
	@for d in $(MODULES); do \
		echo "==> test-cover $$d"; \
		(cd $$d && go test -cover ./...) || exit 1; \
	done

tidy:
	@for d in $(MODULES); do \
		echo "==> tidy $$d"; \
		(cd $$d && GOPROXY=https://proxy.golang.org,direct go mod tidy) || exit 1; \
	done

build-image:
	@for d in $(MODULES); do \
		name=$$(basename $$d); \
		echo "==> build image $$name"; \
		docker build -t $(IMAGE_REGISTRY)/$$name:$(VERSION) \
			--build-arg VERSION=$(VERSION) \
			--build-arg GOPROXY=$(GOPROXY) \
			$$d || exit 1; \
	done

site-openapi:
	mkdir -p site/openapi
	cp shop/openapi.yaml site/openapi/shop.yaml
	cp admin/openapi.yaml site/openapi/admin.yaml

db-init:
	./scripts/init-db.sh

migrate-up:
	./scripts/migrate.sh up

migrate-down:
	./scripts/migrate.sh down

up:
	docker compose -f deploy/docker-compose.yml up -d --build

down:
	docker compose -f deploy/docker-compose.yml down

ps:
	docker compose -f deploy/docker-compose.yml ps

logs:
	docker compose -f deploy/docker-compose.yml logs -f

infra-up:
	docker compose -f deploy/docker-compose.yml up -d postgres redis rabbitmq consul elasticsearch jaeger

infra-down:
	docker compose -f deploy/docker-compose.yml stop postgres redis rabbitmq consul elasticsearch jaeger

e2e:
	./scripts/e2e.sh
