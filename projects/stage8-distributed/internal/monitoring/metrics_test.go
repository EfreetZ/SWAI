package monitoring

import "testing"

func TestHealthy(t *testing.T) {
	m := Metrics{LockContention: 1}
	if !m.Healthy(2) {
		t.Fatal("should be healthy")
	}
}
