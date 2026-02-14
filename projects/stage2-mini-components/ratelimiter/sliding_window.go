package ratelimiter

import (
	"sync"
	"time"
)

// SlidingWindowLimiter 是滑动窗口限流器。
type SlidingWindowLimiter struct {
	limit  int
	window time.Duration

	mu      sync.Mutex
	records []time.Time
}

// NewSlidingWindowLimiter 创建滑动窗口限流器。
func NewSlidingWindowLimiter(limit int, window time.Duration) *SlidingWindowLimiter {
	if limit <= 0 {
		limit = 1
	}
	if window <= 0 {
		window = time.Second
	}
	return &SlidingWindowLimiter{limit: limit, window: window, records: make([]time.Time, 0, limit)}
}

// Allow 判断请求是否通过。
func (l *SlidingWindowLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	boundary := now.Add(-l.window)
	idx := 0
	for idx < len(l.records) && l.records[idx].Before(boundary) {
		idx++
	}
	if idx > 0 {
		l.records = append([]time.Time(nil), l.records[idx:]...)
	}

	if len(l.records) >= l.limit {
		return false
	}
	l.records = append(l.records, now)
	return true
}
