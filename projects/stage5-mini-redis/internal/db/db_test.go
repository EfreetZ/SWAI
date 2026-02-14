package db

import (
	"context"
	"testing"
	"time"
)

func TestStringCommands(t *testing.T) {
	d := New()
	ctx := context.Background()
	if err := d.SetString(ctx, "k", "v", 0); err != nil {
		t.Fatalf("SetString() error = %v", err)
	}
	value, ok, err := d.GetString(ctx, "k")
	if err != nil || !ok || value != "v" {
		t.Fatalf("GetString() = (%q, %v, %v), want (v, true, nil)", value, ok, err)
	}
}

func TestTTLExpire(t *testing.T) {
	d := New()
	ctx := context.Background()
	_ = d.SetString(ctx, "k", "v", 20*time.Millisecond)
	time.Sleep(30 * time.Millisecond)
	_, ok, err := d.GetString(ctx, "k")
	if err != nil {
		t.Fatalf("GetString() error = %v", err)
	}
	if ok {
		t.Fatal("key should be expired")
	}
}

func TestExecuteCommand(t *testing.T) {
	d := New()
	ctx := context.Background()
	if result, err := d.ExecuteCommand(ctx, []string{"SET", "k", "v"}); err != nil || result != "OK" {
		t.Fatalf("SET result = (%q, %v)", result, err)
	}
	if result, err := d.ExecuteCommand(ctx, []string{"GET", "k"}); err != nil || result != "v" {
		t.Fatalf("GET result = (%q, %v)", result, err)
	}
}

func BenchmarkSetGet(b *testing.B) {
	d := New()
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_ = d.SetString(ctx, "k", "v", 0)
		_, _, _ = d.GetString(ctx, "k")
	}
}
