package data

import (
	goodsV1 "admin/api/service/goods/v1"
	userV1 "admin/api/service/user/v1"
	"admin/internal/conf"
	"context"
	consul "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"
	consulAPI "github.com/hashicorp/consul/api"
	grpcx "google.golang.org/grpc"
	"time"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewUserRepo, NewAddressRepo, NewUserServiceClient, NewGoodsServiceClient, NewRegistrar, NewDiscovery)

// Data .
type Data struct {
	log *log.Helper
	uc  userV1.UserClient
	gc  goodsV1.GoodsClient
}

// NewData .
func NewData(c *conf.Data, uc userV1.UserClient, gc goodsV1.GoodsClient, logger log.Logger) (*Data, error) {
	l := log.NewHelper(log.With(logger, "module", "data"))
	return &Data{log: l, uc: uc, gc: gc}, nil
}

// NewUserServiceClient 链接用户服务 grpc
func NewUserServiceClient(ac *conf.Auth, sr *conf.Service, rr registry.Discovery) userV1.UserClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(sr.User.Endpoint),
		grpc.WithDiscovery(rr),
		grpc.WithMiddleware(
			tracing.Client(),
			recovery.Recovery(),
		),
		grpc.WithTimeout(2*time.Second),
		grpc.WithOptions(grpcx.WithStatsHandler(&tracing.ClientHandler{})),
	)
	if err != nil {
		panic(err)
	}
	c := userV1.NewUserClient(conn)
	return c
}

// NewGoodsServiceClient 链接商品服务 grpc
func NewGoodsServiceClient(ac *conf.Auth, sr *conf.Service, rr registry.Discovery) goodsV1.GoodsClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(sr.Goods.Endpoint),
		grpc.WithDiscovery(rr),
		grpc.WithMiddleware(
			tracing.Client(),
			recovery.Recovery(),
		),
		grpc.WithTimeout(2*time.Second),
		grpc.WithOptions(grpcx.WithStatsHandler(&tracing.ClientHandler{})),
	)
	if err != nil {
		panic(err)
	}
	return goodsV1.NewGoodsClient(conn)
}

// NewRegistrar add consul
func NewRegistrar(conf *conf.Registry) registry.Registrar {
	c := consulAPI.DefaultConfig()
	c.Address = conf.Consul.Address
	c.Scheme = conf.Consul.Scheme
	cli, err := consulAPI.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(true))
	return r
}

func NewDiscovery(conf *conf.Registry) registry.Discovery {
	c := consulAPI.DefaultConfig()
	c.Address = conf.Consul.Address
	c.Scheme = conf.Consul.Scheme
	cli, err := consulAPI.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(false))
	return r
}
