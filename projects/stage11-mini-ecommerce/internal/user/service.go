package user

import (
	"context"
	"errors"
	"sync"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
)

var ErrUserExists = errors.New("user exists")
var ErrInvalidCredential = errors.New("invalid credential")

// Service 用户服务。
type Service struct {
	mu     sync.RWMutex
	users  map[string]model.User
	nextID int64
}

func NewService() *Service {
	return &Service{users: make(map[string]model.User), nextID: 1}
}

func (s *Service) Register(ctx context.Context, username, password string) (model.User, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return model.User{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[username]; ok {
		return model.User{}, ErrUserExists
	}
	u := model.User{ID: s.nextID, Username: username, Password: password}
	s.nextID++
	s.users[username] = u
	return u, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (model.User, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return model.User{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[username]
	if !ok || u.Password != password {
		return model.User{}, ErrInvalidCredential
	}
	return u, nil
}
