package lfu

import "testing"

func TestLFUPutGetEvict(t *testing.T) {
	cache, err := New[string, int](2)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	cache.Put("a", 1)
	cache.Put("b", 2)
	_, _ = cache.Get("a")
	_, _ = cache.Get("a")
	cache.Put("c", 3)

	if _, ok := cache.Get("b"); ok {
		t.Fatal("b should be evicted as low frequency key")
	}
	if v, ok := cache.Get("a"); !ok || v != 1 {
		t.Fatalf("Get(a) = (%v, %v), want (1, true)", v, ok)
	}
}

func TestLFUUpdate(t *testing.T) {
	cache, err := New[string, int](2)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	cache.Put("a", 1)
	cache.Put("a", 10)
	if v, ok := cache.Get("a"); !ok || v != 10 {
		t.Fatalf("Get(a) = (%v, %v), want (10, true)", v, ok)
	}
}

func BenchmarkLFUGet(b *testing.B) {
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
