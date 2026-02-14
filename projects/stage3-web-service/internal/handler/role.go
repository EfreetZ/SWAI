package handler

import (
	"net/http"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/service"
)

// RoleHandler 角色处理器。
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler 创建角色处理器。
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// ListRoles 返回角色列表。
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		Error(w, http.StatusMethodNotAllowed, ErrCodeBadRequest, "method not allowed")
		return
	}
	roles, err := h.roleService.List(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, ErrCodeInternal, "internal server error")
		return
	}
	Success(w, roles)
}
