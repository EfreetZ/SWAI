package performance

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/db"
)

// TestSetGetPerformance 验证基础 SET/GET 吞吐表现。
func TestSetGetPerformance(t *testing.T) {
	d := db.New()
	ctx := context.Background()
	start := time.Now()

	for i := 0; i < 10000; i++ {
		key := "k" + strconv.Itoa(i)
		if _, err := d.ExecuteCommand(ctx, []string{"SET", key, "v"}); err != nil {
			t.Fatalf("SET error = %v", err)
		}
	}
	for i := 0; i < 10000; i++ {
		key := "k" + strconv.Itoa(i)
		if _, err := d.ExecuteCommand(ctx, []string{"GET", key}); err != nil {
			t.Fatalf("GET error = %v", err)
		}
	}
	t.Logf("set/get 20000 ops elapsed=%s", time.Since(start))
}
