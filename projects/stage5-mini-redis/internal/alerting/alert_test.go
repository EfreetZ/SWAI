package alerting

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/monitoring"
)

func TestEvaluate(t *testing.T) {
	metrics := monitoring.Snapshot(1000, 200, 10, 4096)
	rule := Rule{MaxQPS: 100, MaxConnected: 50, MaxAOFSizeBytes: 1024}
	alerts := Evaluate(metrics, rule)
	if len(alerts) != 3 {
		t.Fatalf("len(alerts) = %d, want 3", len(alerts))
	}
}
