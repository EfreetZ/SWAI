package monitoring

import "testing"

func TestErrorRate(t *testing.T) {
	m := RPCMetrics{Requests: 10, Failures: 2}
	if m.ErrorRate() != 0.2 {
		t.Fatalf("unexpected error rate: %f", m.ErrorRate())
	}
}
