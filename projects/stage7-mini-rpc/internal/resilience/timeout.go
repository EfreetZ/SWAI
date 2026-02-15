package resilience

import (
	"context"
	"errors"
	"time"
)

var ErrTimeout = errors.New("rpc call timeout")

// TimeoutConfig 超时配置。
type TimeoutConfig struct {
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	CallTimeout    time.Duration
}

// CallWithTimeout 包装调用超时。
func CallWithTimeout(ctx context.Context, fn func() error, timeout time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		return fn()
	}
	childCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()
	select {
	case err := <-done:
		return err
	case <-childCtx.Done():
		return ErrTimeout
	}
}
