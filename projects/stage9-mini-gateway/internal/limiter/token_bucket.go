package limiter

import (
	"sync"
	"time"
)

// TokenBucket 令牌桶。
type TokenBucket struct {
	mu       sync.Mutex
	capacity float64
	tokens   float64
	rate     float64
	last     time.Time
}

// NewTokenBucket 创建令牌桶。
func NewTokenBucket(qps, burst int) *TokenBucket {
	if qps <= 0 {
		qps = 1000
	}
	if burst <= 0 {
		burst = qps
	}
	now := time.Now()
	return &TokenBucket{capacity: float64(burst), tokens: float64(burst), rate: float64(qps), last: now}
}

// Allow 判断是否放行。
func (b *TokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * b.rate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens -= 1
	return true
}
