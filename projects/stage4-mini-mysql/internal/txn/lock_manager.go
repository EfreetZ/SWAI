package txn

import (
	"context"
	"sync"
)

// LockManager 简化页级锁管理。
type LockManager struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

// NewLockManager 创建锁管理器。
func NewLockManager() *LockManager {
	return &LockManager{locks: make(map[string]*sync.Mutex)}
}

// LockKey 加锁。
func (m *LockManager) LockKey(ctx context.Context, key string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	lock, ok := m.locks[key]
	if !ok {
		lock = &sync.Mutex{}
		m.locks[key] = lock
	}
	m.mu.Unlock()

	lock.Lock()
	return nil
}

// UnlockKey 解锁。
func (m *LockManager) UnlockKey(key string) {
	m.mu.Lock()
	lock := m.locks[key]
	m.mu.Unlock()
	if lock != nil {
		lock.Unlock()
	}
}
