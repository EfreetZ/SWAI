package stress

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/db"
)

// TestStressConcurrentCommands 高并发压测基础命令。
func TestStressConcurrentCommands(t *testing.T) {
	d := db.New()
	ctx := context.Background()

	const workers = 64
	const loops = 500
	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < loops; i++ {
				key := "w" + strconv.Itoa(workerID) + "-" + strconv.Itoa(i)
				_, _ = d.ExecuteCommand(ctx, []string{"SET", key, "v"})
				_, _ = d.ExecuteCommand(ctx, []string{"GET", key})
			}
		}(w)
	}
	wg.Wait()
}
