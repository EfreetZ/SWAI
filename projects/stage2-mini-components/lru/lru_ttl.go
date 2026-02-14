package lru

import (
	"time"
)

type ttlValue[V any] struct {
	value     V
	expiresAt time.Time
}

// TTLCache 是带 TTL 的 LRU 缓存。
type TTLCache[K comparable, V any] struct {
	cache *Cache[K, ttlValue[V]]
}

// NewTTL 创建带 TTL 的缓存。
func NewTTL[K comparable, V any](capacity int) (*TTLCache[K, V], error) {
	cache, err := New[K, ttlValue[V]](capacity)
	if err != nil {
		return nil, err
	}
	return &TTLCache[K, V]{cache: cache}, nil
}

// Put 写入带过期时间的值。
func (c *TTLCache[K, V]) Put(key K, value V, ttl time.Duration) {
	expiresAt := time.Now().Add(ttl)
	if ttl <= 0 {
		expiresAt = time.Now()
	}
	c.cache.Put(key, ttlValue[V]{value: value, expiresAt: expiresAt})
}

// Get 读取并惰性淘汰过期值。
func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	n, ok := c.cache.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	entry := n.value
	if time.Now().After(entry.expiresAt) {
		c.cache.removeNode(n)
		delete(c.cache.items, key)
		var zero V
		return zero, false
	}
	c.cache.moveToHead(n)
	return entry.value, true
}

// Len 返回缓存长度。
func (c *TTLCache[K, V]) Len() int {
	return c.cache.Len()
}
