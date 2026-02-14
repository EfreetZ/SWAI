package interfaces

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"

	baseerrors "github.com/EfreetZ/SWAI/projects/stage1-go-core/basics/errors"
)

func TestInMemoryStoreCRUD(t *testing.T) {
	t.Parallel()

	store := NewInMemoryStore()
	ctx := context.Background()

	if err := store.Set(ctx, "k1", "v1"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	value, err := store.Get(ctx, "k1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if value != "v1" {
		t.Fatalf("Get() = %q, want %q", value, "v1")
	}

	if err := store.Delete(ctx, "k1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = store.Get(ctx, "k1")
	if !baseerrors.IsNotFound(err) {
		t.Fatalf("Get() after delete error = %v, want not found", err)
	}
}

func TestInMemoryStoreContextCancel(t *testing.T) {
	t.Parallel()

	store := NewInMemoryStore()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := store.Set(ctx, "k", "v"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Set() error = %v, want context canceled", err)
	}
}

func TestInMemoryStoreConcurrent(t *testing.T) {
	t.Parallel()

	store := NewInMemoryStore()
	ctx := context.Background()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "k" + strconv.Itoa(i)
			_ = store.Set(ctx, key, "v")
			_, _ = store.Get(ctx, key)
		}(i)
	}
	wg.Wait()
}

func BenchmarkInMemoryStoreSet(b *testing.B) {
	store := NewInMemoryStore()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.Set(ctx, strconv.Itoa(i), "value")
	}
}
