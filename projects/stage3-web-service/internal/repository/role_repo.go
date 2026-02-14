package repository

import (
	"context"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/model"
)

// RoleRepository 角色仓储接口。
type RoleRepository interface {
	List(ctx context.Context) ([]model.Role, error)
	GetByName(ctx context.Context, name string) (*model.Role, error)
}

// InMemoryRoleRepo 角色仓储内存实现。
type InMemoryRoleRepo struct {
	roles map[string]model.Role
}

// NewInMemoryRoleRepo 创建角色仓储。
func NewInMemoryRoleRepo() *InMemoryRoleRepo {
	return &InMemoryRoleRepo{
		roles: map[string]model.Role{
			model.RoleAdmin: {
				Name: model.RoleAdmin,
				Permissions: []model.Permission{
					{Resource: "user", Action: "create"},
					{Resource: "user", Action: "read"},
					{Resource: "user", Action: "update"},
					{Resource: "user", Action: "delete"},
					{Resource: "role", Action: "read"},
					{Resource: "role", Action: "update"},
				},
			},
			model.RoleEditor: {
				Name: model.RoleEditor,
				Permissions: []model.Permission{
					{Resource: "user", Action: "read"},
					{Resource: "user", Action: "update"},
				},
			},
			model.RoleViewer: {
				Name: model.RoleViewer,
				Permissions: []model.Permission{
					{Resource: "user", Action: "read"},
				},
			},
		},
	}
}

// List 返回角色列表。
func (r *InMemoryRoleRepo) List(ctx context.Context) ([]model.Role, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	roles := make([]model.Role, 0, len(r.roles))
	for _, role := range r.roles {
		roles = append(roles, role)
	}
	return roles, nil
}

// GetByName 按名称获取角色。
func (r *InMemoryRoleRepo) GetByName(ctx context.Context, name string) (*model.Role, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	role, ok := r.roles[name]
	if !ok {
		return nil, model.ErrNotFound
	}
	copyRole := role
	return &copyRole, nil
}
