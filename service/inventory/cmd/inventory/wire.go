//go:build wireinject
// +build wireinject

package main

import (
	"inventory/internal/biz"
	"inventory/internal/conf"
	"inventory/internal/data"
	"inventory/internal/server"
	"inventory/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func initApp(*conf.Server, *conf.Registry, *conf.Data, *conf.Mq, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
