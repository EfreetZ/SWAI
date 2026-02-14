package lru

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestLRUCachePutGetEvict(t *testing.T) {
	cache, err := New[string, int](2)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	cache.Put("a", 1)
	cache.Put("b", 2)
	if _, ok := cache.Get("a"); !ok {
		t.Fatal("Get(a) should hit")
	}

	cache.Put("c", 3)
	if _, ok := cache.Get("b"); ok {
		t.Fatal("Get(b) should be evicted")
	}
}

func TestConcurrentCache(t *testing.T) {
	cache, err := NewConcurrent[string, int](64)
	if err != nil {
		t.Fatalf("NewConcurrent() error = %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := strconv.Itoa(i)
			cache.Put(key, i)
			_, _ = cache.Get(key)
		}(i)
	}
	wg.Wait()
}

func TestTTLCacheExpire(t *testing.T) {
	cache, err := NewTTL[string, int](2)
	if err != nil {
		t.Fatalf("NewTTL() error = %v", err)
	}

	cache.Put("a", 1, 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	if _, ok := cache.Get("a"); ok {
		t.Fatal("Get(a) should expire")
	}
}

func BenchmarkLRUGet(b *testing.B) {
	cache, err := New[int, int](1024)
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}
	for i := 0; i < 1024; i++ {
		cache.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(i % 1024)
	}
}
