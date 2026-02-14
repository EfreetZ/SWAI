package functions

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidAttempts = errors.New("attempts must be greater than 0")
	ErrNilFunc         = errors.New("fn must not be nil")
)

// Retry 在给定次数内重试函数执行，支持 context 取消。
func Retry(ctx context.Context, attempts int, backoff func(int) time.Duration, fn func(context.Context) error) error {
	if attempts <= 0 {
		return ErrInvalidAttempts
	}
	if fn == nil {
		return ErrNilFunc
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var lastErr error
	for i := 1; i <= attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}
		lastErr = err

		if i == attempts {
			break
		}

		wait := time.Duration(0)
		if backoff != nil {
			wait = backoff(i)
		}
		if wait <= 0 {
			continue
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}

	return lastErr
}
