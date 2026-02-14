package delayqueue

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestHeapQueueRunAndCancel(t *testing.T) {
	queue := NewHeapQueue()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go queue.Start(ctx)

	var ran atomic.Int32
	err := queue.Add(&Task{
		ID:        "task-1",
		ExecuteAt: time.Now().Add(20 * time.Millisecond),
		Callback: func() {
			ran.Add(1)
		},
	})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	err = queue.Add(&Task{ID: "task-2", ExecuteAt: time.Now().Add(30 * time.Millisecond), Callback: func() { ran.Add(1) }})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if err = queue.Cancel("task-2"); err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}

	time.Sleep(80 * time.Millisecond)
	if got := ran.Load(); got != 1 {
		t.Fatalf("ran = %d, want 1", got)
	}
}

func TestTimingWheelAdd(t *testing.T) {
	wheel := NewTimingWheel(10*time.Millisecond, 16)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go wheel.Start(ctx)

	var fired atomic.Int32
	wheel.Add(25*time.Millisecond, func() {
		fired.Add(1)
	})

	time.Sleep(60 * time.Millisecond)
	if fired.Load() != 1 {
		t.Fatalf("fired = %d, want 1", fired.Load())
	}
}

func BenchmarkHeapQueueAdd(b *testing.B) {
	queue := NewHeapQueue()
	for i := 0; i < b.N; i++ {
		_ = queue.Add(&Task{
			ID:        time.Now().String() + "-" + string(rune(i%26+'a')),
			ExecuteAt: time.Now().Add(time.Second),
			Callback:  func() {},
		})
	}
}
