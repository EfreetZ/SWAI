package ttl

import (
	"context"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/db"
)

func TestHeap(t *testing.T) {
	h := NewHeap()
	h.Push("a", 1)
	h.Push("b", 2)
	key, _, ok := h.Pop()
	if !ok || key != "a" {
		t.Fatalf("Pop() = (%q, %v), want (a, true)", key, ok)
	}
}

func TestTTLManagerActiveExpire(t *testing.T) {
	database := db.New()
	manager := NewManager(database)
	ctx := context.Background()
	_ = database.SetString(ctx, "k", "v", 10*time.Millisecond)
	time.Sleep(15 * time.Millisecond)
	manager.ActiveExpire(ctx, 20)
	_, ok, _ := database.GetString(ctx, "k")
	if ok {
		t.Fatal("key should be expired")
	}
}
