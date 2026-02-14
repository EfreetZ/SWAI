package chaos

import (
	"context"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/db"
)

// TestChaosTTLAndWrongType 模拟过期与错误类型访问场景。
func TestChaosTTLAndWrongType(t *testing.T) {
	d := db.New()
	ctx := context.Background()

	_, _ = d.ExecuteCommand(ctx, []string{"SET", "k", "v", "EX", "1"})
	time.Sleep(1100 * time.Millisecond)
	result, err := d.ExecuteCommand(ctx, []string{"GET", "k"})
	if err != nil {
		t.Fatalf("GET expired key error = %v", err)
	}
	if result != "(nil)" {
		t.Fatalf("expired key result = %q, want (nil)", result)
	}

	_, _ = d.ExecuteCommand(ctx, []string{"SET", "a", "1"})
	if _, err = d.ExecuteCommand(ctx, []string{"LPUSH", "a", "x"}); err == nil {
		t.Fatal("LPUSH on string key should fail")
	}
}
