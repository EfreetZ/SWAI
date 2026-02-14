package bench

import (
	"strconv"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage2-mini-components/lfu"
	"github.com/EfreetZ/SWAI/projects/stage2-mini-components/lru"
)

func BenchmarkLRUvsLFUGet(b *testing.B) {
	lruCache, err := lru.New[string, int](2048)
	if err != nil {
		b.Fatalf("lru.New() error = %v", err)
	}
	lfuCache, err := lfu.New[string, int](2048)
	if err != nil {
		b.Fatalf("lfu.New() error = %v", err)
	}

	for i := 0; i < 2048; i++ {
		key := strconv.Itoa(i)
		lruCache.Put(key, i)
		lfuCache.Put(key, i)
	}

	b.Run("lru_get", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = lruCache.Get(strconv.Itoa(i % 2048))
		}
	})

	b.Run("lfu_get", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = lfuCache.Get(strconv.Itoa(i % 2048))
		}
	})
}

func BenchmarkLRUTTLGet(b *testing.B) {
	ttlCache, err := lru.NewTTL[string, int](2048)
	if err != nil {
		b.Fatalf("lru.NewTTL() error = %v", err)
	}
	for i := 0; i < 2048; i++ {
		ttlCache.Put(strconv.Itoa(i), i, time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ttlCache.Get(strconv.Itoa(i % 2048))
	}
}
