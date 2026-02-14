package monitoring

import "testing"

func TestBufferHitRatio(t *testing.T) {
	metrics := StorageMetrics{BufferHits: 90, BufferMisses: 10}
	if got := metrics.BufferHitRatio(); got != 0.9 {
		t.Fatalf("BufferHitRatio() = %f, want 0.9", got)
	}
}
