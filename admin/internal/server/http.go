package server

import (
	v1 "admin/api/admin/v1"
	"admin/internal/conf"
	"admin/internal/service"
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	kmetrics "github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwt2 "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
	"go.opentelemetry.io/otel"
)

// NewHTTPServer new an HTTP s.
func NewHTTPServer(c *conf.Server, ac *conf.Auth, s *service.AdminService, logger log.Logger) *http.Server {
	meter := otel.Meter("kratos")
	requests, _ := kmetrics.DefaultRequestsCounter(meter, "server_requests_code_total")
	seconds, _ := kmetrics.DefaultSecondsHistogram(meter, "server_requests_seconds_bucket")
	var opts = []http.ServerOption{
		http.Middleware(
			rateLimitMiddleware(),
			recovery.Recovery(),
			validate.Validator(),
			tracing.Server(),
			kmetrics.Server(
				kmetrics.WithRequests(requests),
				kmetrics.WithSeconds(seconds),
			),
			selector.Server(
				jwt.Server(func(token *jwt2.Token) (interface{}, error) {
					return []byte(ac.JwtKey), nil
				}, jwt.WithSigningMethod(jwt2.SigningMethodHS256)),
				adminOnlyMiddleware(),
			).Match(NewWhiteListMatcher()).Build(),
			logging.Server(logger),
		),
		http.Filter(handlers.CORS(
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
			handlers.AllowedOrigins([]string{"*"}),
		)),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterAdminHTTPServer(srv, s)
	return srv
}

// NewWhiteListMatcher 白名单不需要token验证的接口
func NewWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	whiteList["/admin.admin.v1.admin/Captcha"] = struct{}{}
	whiteList["/admin.admin.v1.admin/Login"] = struct{}{}
	whiteList["/admin.admin.v1.admin/Register"] = struct{}{}
	whiteList["/admin.admin.v1.admin/RefreshToken"] = struct{}{}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}
