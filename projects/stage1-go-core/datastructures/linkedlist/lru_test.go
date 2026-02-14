package linkedlist

import (
	"strconv"
	"sync"
	"testing"
)

func TestNewLRUCacheInvalidCapacity(t *testing.T) {
	if _, err := NewLRUCache[string, int](0); err != ErrInvalidCapacity {
		t.Fatalf("NewLRUCache() error = %v, want %v", err, ErrInvalidCapacity)
	}
}

func TestLRUCachePutGetEvict(t *testing.T) {
	cache, err := NewLRUCache[string, int](2)
	if err != nil {
		t.Fatalf("NewLRUCache() error = %v", err)
	}

	cache.Put("a", 1)
	cache.Put("b", 2)

	if v, ok := cache.Get("a"); !ok || v != 1 {
		t.Fatalf("Get(a) = (%v, %v), want (1, true)", v, ok)
	}

	cache.Put("c", 3) // 应淘汰 b

	if _, ok := cache.Get("b"); ok {
		t.Fatal("Get(b) should be evicted")
	}
	if v, ok := cache.Get("c"); !ok || v != 3 {
		t.Fatalf("Get(c) = (%v, %v), want (3, true)", v, ok)
	}
}

func TestLRUCacheUpdateExisting(t *testing.T) {
	cache, err := NewLRUCache[string, int](2)
	if err != nil {
		t.Fatalf("NewLRUCache() error = %v", err)
	}

	cache.Put("a", 1)
	cache.Put("a", 10)

	if v, ok := cache.Get("a"); !ok || v != 10 {
		t.Fatalf("Get(a) = (%v, %v), want (10, true)", v, ok)
	}
}

func TestLRUCacheConcurrentAccess(t *testing.T) {
	cache, err := NewLRUCache[string, int](64)
	if err != nil {
		t.Fatalf("NewLRUCache() error = %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "k" + strconv.Itoa(i)
			cache.Put(key, i)
			_, _ = cache.Get(key)
		}(i)
	}
	wg.Wait()

	if got := cache.Len(); got > 64 {
		t.Fatalf("Len() = %d, want <= 64", got)
	}
}

func BenchmarkLRUCachePut(b *testing.B) {
	cache, err := NewLRUCache[string, int](1024)
	if err != nil {
		b.Fatalf("NewLRUCache() error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Put(strconv.Itoa(i), i)
	}
}

func BenchmarkLRUCacheGet(b *testing.B) {
	cache, err := NewLRUCache[int, int](1024)
	if err != nil {
		b.Fatalf("NewLRUCache() error = %v", err)
	}
	for i := 0; i < 1024; i++ {
		cache.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(i % 1024)
	}
}
