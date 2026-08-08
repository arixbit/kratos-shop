package data

import (
	"context"
	slog "log"
	"os"
	"time"

	"payment/internal/conf"
	orderV1 "payment/api/service/order/v1"

	consul "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	consulAPI "github.com/hashicorp/consul/api"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	grpcx "google.golang.org/grpc"
)

var ProviderSet = wire.NewSet(NewData, NewDB, NewPaymentRepo, NewOrderServiceClient, NewDiscovery)

type Data struct {
	db          *gorm.DB
	orderClient orderV1.OrderClient
}

func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, oc orderV1.OrderClient) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{db: db, orderClient: oc}, cleanup, nil
}

func NewDB(c *conf.Data) *gorm.DB {
	newLogger := logger.New(
		slog.New(os.Stdout, "\r\n", slog.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			Colorful:      true,
			LogLevel:      logger.Info,
		},
	)
	db, err := gorm.Open(postgres.Open(c.Database.Source), &gorm.Config{Logger: newLogger})
	if err != nil {
		log.Errorf("failed opening connection: %v", err)
		panic("failed to connect database")
	}
	return db
}

func NewDiscovery(conf *conf.Registry) registry.Discovery {
	c := consulAPI.DefaultConfig()
	c.Address = conf.Consul.Address
	c.Scheme = conf.Consul.Scheme
	cli, err := consulAPI.NewClient(c)
	if err != nil {
		panic(err)
	}
	return consul.New(cli, consul.WithHealthCheck(false))
}

func NewOrderServiceClient(sr *conf.Service, rr registry.Discovery) orderV1.OrderClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(sr.Order.Endpoint),
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
	return orderV1.NewOrderClient(conn)
}
