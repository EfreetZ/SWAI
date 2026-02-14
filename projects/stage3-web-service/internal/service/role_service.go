package service

import (
	"context"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/model"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/repository"
)

// RoleService 角色服务。
type RoleService struct {
	roleRepo repository.RoleRepository
}

// NewRoleService 创建角色服务。
func NewRoleService(roleRepo repository.RoleRepository) *RoleService {
	return &RoleService{roleRepo: roleRepo}
}

// List 列出角色。
func (s *RoleService) List(ctx context.Context) ([]model.Role, error) {
	return s.roleRepo.List(ctx)
}
