.PHONY: help build test tidy db-init up down ps logs e2e infra-up infra-down

MODULES := admin service/user service/goods service/cart service/order service/inventory service/payment shop

help:
	@echo "Usage:"
	@echo "  make build    构建全部服务"
	@echo "  make test     运行全部服务的 Go 测试"
	@echo "  make tidy     整理全部模块依赖"
	@echo "  make db-init  初始化 PostgreSQL 数据库与 Mock 数据"
	@echo "  make up       一键启动整套环境（deploy/docker-compose.yml）"
	@echo "  make down     停止整套环境"
	@echo "  make ps       查看整套环境容器状态"
	@echo "  make logs     查看整套环境日志"
	@echo "  make e2e      运行端到端测试（需服务已启动）"
	@echo "  make infra-up   只启动基础设施（Postgres/Redis/RabbitMQ/Consul/ES/Jaeger）"
	@echo "  make infra-down 停止基础设施容器"

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

tidy:
	@for d in $(MODULES); do \
		echo "==> tidy $$d"; \
		(cd $$d && GOPROXY=https://proxy.golang.org,direct go mod tidy) || exit 1; \
	done

db-init:
	./scripts/init-db.sh

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
