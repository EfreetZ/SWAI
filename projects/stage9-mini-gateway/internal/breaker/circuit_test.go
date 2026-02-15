package breaker

import (
	"errors"
	"testing"
	"time"
)

func TestBreaker(t *testing.T) {
	cb := New(1, time.Millisecond)
	err := cb.Execute(func() error { return errors.New("fail") })
	if err == nil {
		t.Fatal("expected failure")
	}
	err = cb.Execute(func() error { return nil })
	if !errors.Is(err, ErrOpen) {
		t.Fatalf("expected open error, got %v", err)
	}
}
