package test

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/db"
)

func TestV0BasicCommands(t *testing.T) {
	d := db.New()
	ctx := context.Background()

	if result, err := d.ExecuteCommand(ctx, []string{"PING"}); err != nil || result != "PONG" {
		t.Fatalf("PING = (%q, %v)", result, err)
	}
	if result, err := d.ExecuteCommand(ctx, []string{"SET", "k", "v"}); err != nil || result != "OK" {
		t.Fatalf("SET = (%q, %v)", result, err)
	}
	if result, err := d.ExecuteCommand(ctx, []string{"GET", "k"}); err != nil || result != "v" {
		t.Fatalf("GET = (%q, %v)", result, err)
	}
}
