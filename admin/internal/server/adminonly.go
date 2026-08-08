package server

import (
	"context"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwt2 "github.com/golang-jwt/jwt/v5"
)

func adminOnlyMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if claims, ok := jwt.FromContext(ctx); ok {
				if m, ok := claims.(jwt2.MapClaims); ok {
					if role, ok := m["AuthorityId"].(float64); ok && int(role) == 2 {
						return handler(ctx, req)
					}
				}
			}
			return nil, kerrors.New(403, "ADMIN_ONLY", "需要管理员权限")
		}
	}
}
