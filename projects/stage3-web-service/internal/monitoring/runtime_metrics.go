package monitoring

import (
	"runtime"
	"time"
)

// RuntimeMetrics 运行时指标。
type RuntimeMetrics struct {
	CollectedAt time.Time
	Goroutines  int
	HeapAlloc   uint64
	HeapObjects uint64
}

// CollectRuntimeMetrics 采集运行时指标。
func CollectRuntimeMetrics() RuntimeMetrics {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return RuntimeMetrics{
		CollectedAt: time.Now(),
		Goroutines:  runtime.NumGoroutine(),
		HeapAlloc:   mem.HeapAlloc,
		HeapObjects: mem.HeapObjects,
	}
}
