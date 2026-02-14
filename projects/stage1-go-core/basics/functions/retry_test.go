package functions

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetrySuccess(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	err := Retry(context.Background(), 3, nil, func(ctx context.Context) error {
		if count.Add(1) < 3 {
			return errors.New("temporary")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Retry() error = %v, want nil", err)
	}
	if got := count.Load(); got != 3 {
		t.Fatalf("attempts = %d, want 3", got)
	}
}

func TestRetryFail(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("permanent")
	err := Retry(context.Background(), 2, nil, func(ctx context.Context) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Retry() error = %v, want %v", err, wantErr)
	}
}

func TestRetryContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()

	err := Retry(ctx, 3, func(i int) time.Duration { return 20 * time.Millisecond }, func(ctx context.Context) error {
		return errors.New("retry")
	})
	if err == nil {
		t.Fatal("Retry() error = nil, want context deadline exceeded")
	}
}

func TestRetryInvalidArgs(t *testing.T) {
	t.Parallel()

	if err := Retry(context.Background(), 0, nil, func(ctx context.Context) error { return nil }); err != ErrInvalidAttempts {
		t.Fatalf("attempts error = %v, want %v", err, ErrInvalidAttempts)
	}
	if err := Retry(context.Background(), 1, nil, nil); err != ErrNilFunc {
		t.Fatalf("fn error = %v, want %v", err, ErrNilFunc)
	}
}

func BenchmarkRetryNoBackoff(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Retry(context.Background(), 1, nil, func(ctx context.Context) error {
			return nil
		})
	}
}
