//go:build wireinject
// +build wireinject

package main

import (
	"payment/internal/biz"
	"payment/internal/conf"
	"payment/internal/data"
	"payment/internal/server"
	"payment/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func initApp(*conf.Server, *conf.Registry, *conf.Data, *conf.Service, *conf.Mq, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
