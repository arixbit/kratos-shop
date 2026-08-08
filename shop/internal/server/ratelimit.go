package server

import (
	"context"
	"sync"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"golang.org/x/time/rate"
)

const (
	rateLimitPerSecond = 20
	rateLimitBurst     = 40
)

var (
	rateLimitMu sync.Mutex
	rateLimits  = map[string]*rate.Limiter{}
)

func rateLimitMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			limiter := getLimiter(clientIP(ctx))
			if !limiter.Allow() {
				return nil, kerrors.New(429, "RATE_LIMITED", "请求过于频繁，请稍后重试")
			}
			return handler(ctx, req)
		}
	}
}

func clientIP(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if v := tr.RequestHeader().Get("X-Forwarded-For"); v != "" {
			return v
		}
		if v := tr.RequestHeader().Get("X-Real-Ip"); v != "" {
			return v
		}
		if v := tr.RequestHeader().Get("RemoteAddr"); v != "" {
			return v
		}
	}
	return "local"
}

func getLimiter(ip string) *rate.Limiter {
	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()
	if l, ok := rateLimits[ip]; ok {
		return l
	}
	l := rate.NewLimiter(rateLimitPerSecond, rateLimitBurst)
	rateLimits[ip] = l
	return l
}
