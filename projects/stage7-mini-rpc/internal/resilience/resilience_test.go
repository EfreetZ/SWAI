package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTimeoutRetryCircuit(t *testing.T) {
	err := CallWithTimeout(context.Background(), func() error {
		time.Sleep(2 * time.Millisecond)
		return nil
	}, time.Nanosecond)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected timeout, got %v", err)
	}

	attempt := 0
	rErr := Retry(context.Background(), &RetryPolicy{MaxRetries: 2, InitialDelay: time.Nanosecond, MaxDelay: time.Microsecond, BackoffFactor: 2}, func() error {
		attempt++
		if attempt < 3 {
			return errors.New("fail")
		}
		return nil
	})
	if rErr != nil {
		t.Fatalf("retry should succeed: %v", rErr)
	}

	cb := NewCircuitBreaker(1, time.Millisecond)
	_ = cb.Execute(func() error { return errors.New("boom") })
	if cb.State() != StateOpen {
		t.Fatalf("expected open state")
	}
}
