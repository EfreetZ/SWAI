package monitoring

import "testing"

func TestCollectRuntimeMetrics(t *testing.T) {
	metrics := CollectRuntimeMetrics()
	if metrics.CollectedAt.IsZero() {
		t.Fatal("CollectedAt is zero")
	}
	if metrics.Goroutines <= 0 {
		t.Fatalf("Goroutines = %d, want > 0", metrics.Goroutines)
	}
}
