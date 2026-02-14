package bench

import (
	"context"
	"strconv"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/db"
)

func BenchmarkSetGet(b *testing.B) {
	d := db.New()
	ctx := context.Background()

	b.Run("set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := d.ExecuteCommand(ctx, []string{"SET", "k" + strconv.Itoa(i), "v"}); err != nil {
				b.Fatalf("SET error = %v", err)
			}
		}
	})

	for i := 0; i < 10000; i++ {
		_, _ = d.ExecuteCommand(ctx, []string{"SET", "k" + strconv.Itoa(i), "v"})
	}

	b.Run("get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := d.ExecuteCommand(ctx, []string{"GET", "k" + strconv.Itoa(i%10000)}); err != nil {
				b.Fatalf("GET error = %v", err)
			}
		}
	})
}
