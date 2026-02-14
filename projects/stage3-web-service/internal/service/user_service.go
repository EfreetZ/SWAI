package service

import (
	"context"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/model"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/repository"
)

// UserService 用户服务。
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建用户服务。
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// List 列出用户。
func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	return s.userRepo.List(ctx)
}

// GetByID 获取用户详情。
func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// Update 更新用户。
func (s *UserService) Update(ctx context.Context, user *model.User) error {
	if user == nil || user.ID <= 0 {
		return model.ErrBadRequest
	}
	return s.userRepo.Update(ctx, user)
}

// Delete 删除用户。
func (s *UserService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return model.ErrBadRequest
	}
	return s.userRepo.Delete(ctx, id)
}
