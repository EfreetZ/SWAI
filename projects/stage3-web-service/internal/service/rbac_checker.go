package service

import (
	"context"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/repository"
)

// RBACChecker RBAC 权限检查器。
type RBACChecker struct {
	roleRepo repository.RoleRepository
}

// NewRBACChecker 创建检查器。
func NewRBACChecker(roleRepo repository.RoleRepository) *RBACChecker {
	return &RBACChecker{roleRepo: roleRepo}
}

// HasPermission 判断 role 对指定资源动作是否有权限。
func (c *RBACChecker) HasPermission(ctx context.Context, roleName, resource, action string) (bool, error) {
	role, err := c.roleRepo.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}
	for _, permission := range role.Permissions {
		if permission.Resource == resource && permission.Action == action {
			return true, nil
		}
	}
	return false, nil
}
