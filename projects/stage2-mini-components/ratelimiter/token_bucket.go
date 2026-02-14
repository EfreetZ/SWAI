package ratelimiter

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrInvalidRate  = errors.New("rate must be greater than 0")
	ErrInvalidBurst = errors.New("burst must be greater than 0")
)

// TokenBucket 令牌桶限流器。
type TokenBucket struct {
	rate     float64
	burst    int
	tokens   float64
	lastTime time.Time
	mu       sync.Mutex
}

// NewTokenBucket 创建令牌桶。
func NewTokenBucket(rate float64, burst int) (*TokenBucket, error) {
	if rate <= 0 {
		return nil, ErrInvalidRate
	}
	if burst <= 0 {
		return nil, ErrInvalidBurst
	}
	return &TokenBucket{rate: rate, burst: burst, tokens: float64(burst), lastTime: time.Now()}, nil
}

// Allow 检查是否允许 1 个请求。
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN 检查是否允许 n 个请求。
func (tb *TokenBucket) AllowN(n int) bool {
	if n <= 0 {
		return true
	}

	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	if tb.tokens < float64(n) {
		return false
	}
	tb.tokens -= float64(n)
	return true
}

// Wait 阻塞等待令牌。
func (tb *TokenBucket) Wait(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	for {
		if tb.Allow() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastTime).Seconds()
	tb.lastTime = now
	tb.tokens += elapsed * tb.rate
	if tb.tokens > float64(tb.burst) {
		tb.tokens = float64(tb.burst)
	}
}
