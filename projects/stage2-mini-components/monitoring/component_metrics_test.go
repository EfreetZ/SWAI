package monitoring

import "testing"

func TestCacheMetricsHitRatio(t *testing.T) {
	m := CacheMetrics{Hits: 8, Misses: 2}
	if got := m.HitRatio(); got != 0.8 {
		t.Fatalf("HitRatio() = %f, want 0.8", got)
	}

	empty := CacheMetrics{}
	if got := empty.HitRatio(); got != 0 {
		t.Fatalf("HitRatio() = %f, want 0", got)
	}
}
