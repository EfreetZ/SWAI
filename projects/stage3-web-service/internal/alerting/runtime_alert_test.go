package alerting

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/monitoring"
)

func TestEvaluateRuntime(t *testing.T) {
	alerts := EvaluateRuntime(monitoring.RuntimeMetrics{Goroutines: 100, HeapAlloc: 2048}, Rule{MaxGoroutines: 10, MaxHeapAlloc: 1024})
	if len(alerts) != 2 {
		t.Fatalf("len(alerts) = %d, want 2", len(alerts))
	}
}
