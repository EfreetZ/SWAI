package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/handler"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/pkg"
)

type userContextKey string

const (
	userIDKey   userContextKey = "user_id"
	roleNameKey userContextKey = "role"
)

// Auth 校验 access token。
func Auth(jwtMgr *pkg.JWTManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			handler.Error(w, http.StatusUnauthorized, handler.ErrCodeUnauthorized, "missing authorization header")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			handler.Error(w, http.StatusUnauthorized, handler.ErrCodeUnauthorized, "invalid authorization header")
			return
		}

		claims, err := jwtMgr.ParseToken(parts[1])
		if err != nil || claims.Type != "access" {
			handler.Error(w, http.StatusUnauthorized, handler.ErrCodeUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, roleNameKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFromContext 获取用户 ID。
func UserIDFromContext(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	if value, ok := ctx.Value(userIDKey).(int64); ok {
		return value
	}
	return 0
}

// RoleFromContext 获取角色。
func RoleFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if value, ok := ctx.Value(roleNameKey).(string); ok {
		return value
	}
	return ""
}
