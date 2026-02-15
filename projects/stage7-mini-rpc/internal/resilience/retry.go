package resilience

import (
	"context"
	"time"
)

// RetryPolicy 重试策略。
type RetryPolicy struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	RetryOn       func(err error) bool
}

// Retry 按策略执行重试。
func Retry(ctx context.Context, policy *RetryPolicy, fn func() error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if policy == nil {
		return fn()
	}
	if policy.InitialDelay <= 0 {
		policy.InitialDelay = 10 * time.Millisecond
	}
	if policy.MaxDelay <= 0 {
		policy.MaxDelay = time.Second
	}
	if policy.BackoffFactor <= 1 {
		policy.BackoffFactor = 2
	}

	delay := policy.InitialDelay
	var lastErr error
	for i := 0; i <= policy.MaxRetries; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		if policy.RetryOn != nil && !policy.RetryOn(err) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		next := time.Duration(float64(delay) * policy.BackoffFactor)
		if next > policy.MaxDelay {
			next = policy.MaxDelay
		}
		delay = next
	}
	return lastErr
}
