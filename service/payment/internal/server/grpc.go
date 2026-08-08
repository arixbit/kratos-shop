package server

import (
	v1 "payment/api/payment/v1"
	"payment/internal/conf"
	"payment/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	kmetrics "github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel"
)

func NewGRPCServer(c *conf.Server, s *service.PaymentService, logger log.Logger) *grpc.Server {
	meter := otel.Meter("kratos")
	requests, _ := kmetrics.DefaultRequestsCounter(meter, "server_requests_code_total")
	seconds, _ := kmetrics.DefaultSecondsHistogram(meter, "server_requests_seconds_bucket")
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
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
	v1.RegisterPaymentServer(srv, s)
	return srv
}
