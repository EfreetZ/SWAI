package lru

import "sync"

// ConcurrentCache 是并发安全 LRU 封装。
type ConcurrentCache[K comparable, V any] struct {
	mu    sync.RWMutex
	cache *Cache[K, V]
}

// NewConcurrent 创建并发安全缓存。
func NewConcurrent[K comparable, V any](capacity int) (*ConcurrentCache[K, V], error) {
	cache, err := New[K, V](capacity)
	if err != nil {
		return nil, err
	}
	return &ConcurrentCache[K, V]{cache: cache}, nil
}

// Get 线程安全读取。
func (c *ConcurrentCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cache.Get(key)
}

// Put 线程安全写入。
func (c *ConcurrentCache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache.Put(key, value)
}

// Len 返回缓存大小。
func (c *ConcurrentCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache.Len()
}
