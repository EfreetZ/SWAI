package middleware

import (
	"net/http"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/handler"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/service"
)

// RBAC 校验角色权限。
func RBAC(checker *service.RBACChecker, resource, action string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := RoleFromContext(r.Context())
		allowed, err := checker.HasPermission(r.Context(), role, resource, action)
		if err != nil {
			handler.Error(w, http.StatusInternalServerError, handler.ErrCodeInternal, "internal server error")
			return
		}
		if !allowed {
			handler.Error(w, http.StatusForbidden, handler.ErrCodeForbidden, "forbidden")
			return
		}
		next.ServeHTTP(w, r)
	})
}
