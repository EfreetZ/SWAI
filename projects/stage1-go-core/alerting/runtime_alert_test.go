package alerting

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage1-go-core/monitoring"
)

func TestEvaluateRuntimeAlerts(t *testing.T) {
	metrics := monitoring.RuntimeMetrics{
		Goroutines: 100,
		HeapAlloc:  512,
		GCCount:    20,
	}
	rule := RuntimeRule{
		MaxGoroutines: 50,
		MaxHeapAlloc:  256,
		MaxGCCount:    10,
	}

	alerts := EvaluateRuntimeAlerts(metrics, rule)
	if len(alerts) != 3 {
		t.Fatalf("len(alerts) = %d, want 3", len(alerts))
	}
}

func TestEvaluateRuntimeAlertsNoAlert(t *testing.T) {
	metrics := monitoring.RuntimeMetrics{Goroutines: 10, HeapAlloc: 64, GCCount: 1}
	rule := RuntimeRule{MaxGoroutines: 100, MaxHeapAlloc: 1024, MaxGCCount: 50}

	alerts := EvaluateRuntimeAlerts(metrics, rule)
	if len(alerts) != 0 {
		t.Fatalf("len(alerts) = %d, want 0", len(alerts))
	}
}
