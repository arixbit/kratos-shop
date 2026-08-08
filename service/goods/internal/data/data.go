package data

import (
	"context"
	inventoryV1 "goods/api/service/inventory/v1"
	"goods/internal/biz"
	"goods/internal/conf"
	slog "log"
	"os"
	"time"

	consul "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-redis/redis/extra/redisotel"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	consulAPI "github.com/hashicorp/consul/api"
	"github.com/olivere/elastic/v7"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	grpcx "google.golang.org/grpc"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewDB, NewTransaction, NewRedis, NewElasticsearch,
	NewBrandRepo,
	NewCategoryRepo,
	NewGoodsTypeRepo,
	NewSpecificationRepo,
	NewGoodsAttrRepo,
	NewGoodsRepo,
	NewGoodsSkuRepoRepo,
	NewInventoryRepo,
	NewEsGoodsRepo,
	NewInventoryServiceClient,
	NewDiscovery,
)

type Data struct {
	db       *gorm.DB
	rdb      *redis.Client
	EsClient *elastic.Client
}

// 用来承载事务的上下文
type contextTxKey struct{}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client, es *elastic.Client) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		db:       db,
		rdb:      rdb,
		EsClient: es,
	}, cleanup, nil
}

// NewTransaction .
func NewTransaction(d *Data) biz.Transaction {
	return d
}

// ExecTx gorm Transaction
func (d *Data) ExecTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, contextTxKey{}, tx)
		return fn(ctx)
	})
}

// DB 根据此方法来判断当前的 db 是不是使用 事务的 DB
func (d *Data) DB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if ok {
		return tx
	}
	return d.db
}

// NewDB .
func NewDB(c *conf.Data) *gorm.DB {
	// 终端打印输入 sql 执行记录
	newLogger := logger.New(
		slog.New(os.Stdout, "\r\n", slog.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢查询 SQL 阈值
			Colorful:      true,        // 禁用彩色打印
			//IgnoreRecordNotFoundError: false,
			LogLevel: logger.Info, // Log lever
		},
	)

	db, err := gorm.Open(postgres.Open(c.Database.Source), &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy:                           schema.NamingStrategy{
			//SingularTable: true, // 表名是否加 s
		},
	})

	if err != nil {
		log.Errorf("failed opening connection to sqlite: %v", err)
		panic("failed to connect database")
	}

	return db
}

func NewRedis(c *conf.Data) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		DialTimeout:  c.Redis.DialTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
	})
	rdb.AddHook(redisotel.TracingHook{})
	return rdb
}

func NewElasticsearch(c *conf.Data) *elastic.Client {
	es, err := elastic.NewClient(elastic.SetURL(c.Elastic.Addr), elastic.SetSniff(false),
		elastic.SetTraceLog(slog.New(os.Stdout, "shop", slog.LstdFlags)))
	if err != nil {
		panic(err)
	}

	return es
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

func NewInventoryServiceClient(sr *conf.Service, rr registry.Discovery) inventoryV1.InventoryClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(sr.Inventory.Endpoint),
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
	return inventoryV1.NewInventoryClient(conn)
}
