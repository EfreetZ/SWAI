package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// UserRepository 用户仓储接口。
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	List(ctx context.Context) ([]model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id int64) error
}

// InMemoryUserRepo 用户仓储内存实现。
type InMemoryUserRepo struct {
	mu        sync.RWMutex
	nextID    int64
	usersByID map[int64]*model.User
	idByName  map[string]int64
	idByEmail map[string]int64
}

// NewInMemoryUserRepo 创建用户仓储。
func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{
		nextID:    1,
		usersByID: make(map[int64]*model.User),
		idByName:  make(map[string]int64),
		idByEmail: make(map[string]int64),
	}
}

// Create 创建用户。
func (r *InMemoryUserRepo) Create(ctx context.Context, user *model.User) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.idByName[user.Username]; ok {
		return model.ErrUserExists
	}
	if _, ok := r.idByEmail[user.Email]; ok {
		return model.ErrUserExists
	}

	id := r.nextID
	r.nextID++
	now := time.Now()
	copyUser := *user
	copyUser.ID = id
	copyUser.CreatedAt = now
	copyUser.UpdatedAt = now
	r.usersByID[id] = &copyUser
	r.idByName[user.Username] = id
	r.idByEmail[user.Email] = id
	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

// GetByUsername 按用户名查询。
func (r *InMemoryUserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.idByName[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	u := *r.usersByID[id]
	return &u, nil
}

// GetByID 按 ID 查询。
func (r *InMemoryUserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.usersByID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	copyUser := *u
	return &copyUser, nil
}

// List 返回用户列表。
func (r *InMemoryUserRepo) List(ctx context.Context) ([]model.User, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]model.User, 0, len(r.usersByID))
	for _, u := range r.usersByID {
		users = append(users, *u)
	}
	return users, nil
}

// Update 更新用户。
func (r *InMemoryUserRepo) Update(ctx context.Context, user *model.User) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.usersByID[user.ID]
	if !ok {
		return ErrUserNotFound
	}
	stored.Email = user.Email
	stored.Role = user.Role
	stored.UpdatedAt = time.Now()
	return nil
}

// Delete 删除用户。
func (r *InMemoryUserRepo) Delete(ctx context.Context, id int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.usersByID[id]
	if !ok {
		return ErrUserNotFound
	}
	delete(r.idByName, u.Username)
	delete(r.idByEmail, u.Email)
	delete(r.usersByID, id)
	return nil
}
