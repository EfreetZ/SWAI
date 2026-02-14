package circuitbreaker

import (
	"errors"
	"testing"
	"time"
)

func TestBreakerStateTransition(t *testing.T) {
	cb := New(2, 1, 20*time.Millisecond, 1)

	fnErr := errors.New("downstream failed")
	_ = cb.Execute(func() error { return fnErr })
	_ = cb.Execute(func() error { return fnErr })

	if cb.State() != Open {
		t.Fatalf("state = %v, want Open", cb.State())
	}

	if err := cb.Execute(func() error { return nil }); err != ErrOpenState {
		t.Fatalf("Execute() in open state = %v, want %v", err, ErrOpenState)
	}

	time.Sleep(25 * time.Millisecond)
	if err := cb.Execute(func() error { return nil }); err != nil {
		t.Fatalf("half-open request error = %v", err)
	}

	if cb.State() != Closed {
		t.Fatalf("state = %v, want Closed", cb.State())
	}
}

func BenchmarkBreakerExecute(b *testing.B) {
	cb := New(1000, 1, time.Second, 1)
	for i := 0; i < b.N; i++ {
		if err := cb.Execute(func() error { return nil }); err != nil {
			b.Fatalf("Execute() error = %v", err)
		}
	}
}
