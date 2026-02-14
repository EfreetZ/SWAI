package test

import (
	"context"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/db"
)

func TestV1PipelineAndTTL(t *testing.T) {
	d := db.New()
	ctx := context.Background()

	commands := [][]string{{"SET", "a", "1"}, {"SET", "b", "2"}, {"GET", "a"}}
	for _, command := range commands {
		if _, err := d.ExecuteCommand(ctx, command); err != nil {
			t.Fatalf("pipeline command %v error = %v", command, err)
		}
	}

	_, _ = d.ExecuteCommand(ctx, []string{"SET", "ttl-key", "v", "EX", "1"})
	time.Sleep(1100 * time.Millisecond)
	result, err := d.ExecuteCommand(ctx, []string{"GET", "ttl-key"})
	if err != nil {
		t.Fatalf("GET ttl-key error = %v", err)
	}
	if result != "(nil)" {
		t.Fatalf("ttl key result = %q, want (nil)", result)
	}
}
