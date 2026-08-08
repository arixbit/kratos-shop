package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	kmetrics "github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel"
	v1 "goods/api/goods/v1"
	"goods/internal/conf"
	"goods/internal/service"
)

// NewGRPCServer new a gRPC s.
func NewGRPCServer(c *conf.Server, greeter *service.GoodsService, logger log.Logger) *grpc.Server {
	meter := otel.Meter("kratos")
	requests, _ := kmetrics.DefaultRequestsCounter(meter, "server_requests_code_total")
	seconds, _ := kmetrics.DefaultSecondsHistogram(meter, "server_requests_seconds_bucket")
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			validate.Validator(),
			logging.Server(logger),
			kmetrics.Server(
				kmetrics.WithRequests(requests),
				kmetrics.WithSeconds(seconds),
			),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterGoodsServer(srv, greeter)
	return srv
}
