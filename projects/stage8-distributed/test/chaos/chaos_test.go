package chaos

import (
	"context"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage8-distributed/internal/coord"
)

func TestCanceledLockAcquire(t *testing.T) {
	ls := coord.NewLockService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := ls.Acquire(ctx, "k", "o", 0); err == nil {
		t.Fatal("expected canceled error")
	}
}
