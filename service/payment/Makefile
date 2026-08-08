GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

.PHONY: init api openapi config build test all
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: api
api:
	protoc --proto_path=api \
	       --go_out=paths=source_relative:api \
	       --go-grpc_out=paths=source_relative:api \
	       $(shell find api -name '*.proto')

.PHONY: openapi
openapi:
	protoc --proto_path=. \
	       --proto_path=./third_party \
	       --openapi_out==paths=source_relative:. \
	       $(shell find api -name '*.proto')

.PHONY: config
config:
	protoc --proto_path=internal \
	       --go_out=paths=source_relative:internal \
	       $(shell find internal -name '*.proto')

.PHONY: build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/payment ./cmd/payment

.PHONY: test
test:
	go test ./...

.PHONY: all
all: api config build

.DEFAULT_GOAL := help

help:
	@echo 'make build / test / api / config'
