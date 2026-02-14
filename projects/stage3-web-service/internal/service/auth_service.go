package service

import (
	"context"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/model"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/pkg"
	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/repository"
)

// AuthService 认证服务。
type AuthService struct {
	userRepo repository.UserRepository
	jwtMgr   *pkg.JWTManager
}

// NewAuthService 创建认证服务。
func NewAuthService(userRepo repository.UserRepository, jwtMgr *pkg.JWTManager) *AuthService {
	return &AuthService{userRepo: userRepo, jwtMgr: jwtMgr}
}

// Register 注册用户。
func (s *AuthService) Register(ctx context.Context, username, email, password string) (*model.User, error) {
	if err := pkg.ValidateRegisterParams(username, email, password); err != nil {
		return nil, err
	}

	hashed, err := pkg.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashed,
		Role:         model.RoleViewer,
	}
	if err = s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login 登录并签发 token。
func (s *AuthService) Login(ctx context.Context, username, password string) (string, string, *model.User, error) {
	if err := pkg.ValidateLoginParams(username, password); err != nil {
		return "", "", nil, err
	}

	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", "", nil, model.ErrUnauthorized
	}

	if err = pkg.VerifyPassword(password, user.PasswordHash); err != nil {
		return "", "", nil, model.ErrWrongPassword
	}

	accessToken, refreshToken, err := s.jwtMgr.GenerateTokenPair(user.ID, user.Username, user.Role)
	if err != nil {
		return "", "", nil, err
	}
	return accessToken, refreshToken, user, nil
}

// Refresh 刷新 access token。
func (s *AuthService) Refresh(refreshToken string) (string, error) {
	return s.jwtMgr.RefreshToken(refreshToken)
}
