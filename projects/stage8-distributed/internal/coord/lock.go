package coord

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrLockTaken = errors.New("lock already taken")
var ErrLockOwner = errors.New("lock owner mismatch")

type lockEntry struct {
	owner     string
	expiresAt time.Time
}

// LockService 分布式锁（内存模拟）。
type LockService struct {
	mu    sync.Mutex
	locks map[string]lockEntry
}

// NewLockService 创建锁服务。
func NewLockService() *LockService {
	return &LockService{locks: make(map[string]lockEntry)}
}

// Acquire 获取锁。
func (s *LockService) Acquire(ctx context.Context, key, owner string, ttl time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = time.Second
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.locks[key]
	if ok && entry.expiresAt.After(now) {
		return ErrLockTaken
	}
	s.locks[key] = lockEntry{owner: owner, expiresAt: now.Add(ttl)}
	return nil
}

// Release 释放锁。
func (s *LockService) Release(ctx context.Context, key, owner string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.locks[key]
	if !ok {
		return nil
	}
	if entry.owner != owner {
		return ErrLockOwner
	}
	delete(s.locks, key)
	return nil
}
