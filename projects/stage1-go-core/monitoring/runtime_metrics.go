package monitoring

import (
	"runtime"
	"time"
)

// RuntimeMetrics 描述运行时关键指标快照。
type RuntimeMetrics struct {
	CollectedAt time.Time
	Goroutines  int
	HeapAlloc   uint64
	HeapObjects uint64
	GCCount     uint32
	CgoCalls    int64
}

// CollectRuntimeMetrics 采集当前进程运行时指标。
func CollectRuntimeMetrics() RuntimeMetrics {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return RuntimeMetrics{
		CollectedAt: time.Now(),
		Goroutines:  runtime.NumGoroutine(),
		HeapAlloc:   mem.HeapAlloc,
		HeapObjects: mem.HeapObjects,
		GCCount:     mem.NumGC,
		CgoCalls:    runtime.NumCgoCall(),
	}
}
