package coord

import (
	"context"
	"testing"
	"time"
)

func TestLockAndSnowflake(t *testing.T) {
	locks := NewLockService()
	if err := locks.Acquire(context.Background(), "k", "o1", 10*time.Millisecond); err != nil {
		t.Fatalf("acquire failed: %v", err)
	}
	if err := locks.Acquire(context.Background(), "k", "o2", 10*time.Millisecond); err == nil {
		t.Fatal("expected contention error")
	}
	if err := locks.Release(context.Background(), "k", "o1"); err != nil {
		t.Fatalf("release failed: %v", err)
	}

	sf, err := NewSnowflake(1)
	if err != nil {
		t.Fatalf("new snowflake failed: %v", err)
	}
	id1, _ := sf.NextID()
	id2, _ := sf.NextID()
	if id2 <= id1 {
		t.Fatal("ids must increase")
	}
}
