package interfaces

import (
	"context"
	"sync"

	baseerrors "github.com/EfreetZ/SWAI/projects/stage1-go-core/basics/errors"
)

// KVStore 定义最小 key-value 存储接口。
type KVStore interface {
	Set(ctx context.Context, key, value string) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

// InMemoryStore 是 KVStore 的内存实现。
type InMemoryStore struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewInMemoryStore 创建内存存储。
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{data: make(map[string]string)}
}

// Set 写入 key-value。
func (s *InMemoryStore) Set(ctx context.Context, key, value string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
	return nil
}

// Get 读取 key 对应值，不存在时返回 ErrNotFound。
func (s *InMemoryStore) Get(ctx context.Context, key string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	s.mu.RLock()
	value, ok := s.data[key]
	s.mu.RUnlock()
	if !ok {
		return "", baseerrors.Wrap("store.get", baseerrors.ErrNotFound)
	}
	return value, nil
}

// Delete 删除 key。
func (s *InMemoryStore) Delete(ctx context.Context, key string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	delete(s.data, key)
	s.mu.Unlock()
	return nil
}
