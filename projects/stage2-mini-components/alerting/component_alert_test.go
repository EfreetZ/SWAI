package alerting

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage2-mini-components/monitoring"
)

func TestEvaluate(t *testing.T) {
	metrics := monitoring.ComponentMetrics{
		Cache: monitoring.CacheMetrics{Hits: 1, Misses: 9},
		Queue: monitoring.QueueMetrics{PendingTasks: 120, LagMillis: 300},
	}
	rule := Rule{MinHitRatio: 0.5, MaxPendingTask: 100, MaxQueueLagMS: 200}

	alerts := Evaluate(metrics, rule)
	if len(alerts) != 3 {
		t.Fatalf("len(alerts) = %d, want 3", len(alerts))
	}
}

func TestEvaluateNoAlert(t *testing.T) {
	metrics := monitoring.ComponentMetrics{
		Cache: monitoring.CacheMetrics{Hits: 9, Misses: 1},
		Queue: monitoring.QueueMetrics{PendingTasks: 20, LagMillis: 30},
	}
	rule := Rule{MinHitRatio: 0.5, MaxPendingTask: 100, MaxQueueLagMS: 200}

	alerts := Evaluate(metrics, rule)
	if len(alerts) != 0 {
		t.Fatalf("len(alerts) = %d, want 0", len(alerts))
	}
}
