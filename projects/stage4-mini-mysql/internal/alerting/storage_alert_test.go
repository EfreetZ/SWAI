package alerting

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/monitoring"
)

func TestEvaluateStorage(t *testing.T) {
	metrics := monitoring.StorageMetrics{BufferHits: 1, BufferMisses: 9, WALSizeBytes: 2048}
	rule := Rule{MinBufferHitRatio: 0.5, MaxWALSizeBytes: 1024}
	alerts := EvaluateStorage(metrics, rule)
	if len(alerts) != 2 {
		t.Fatalf("len(alerts) = %d, want 2", len(alerts))
	}
}
