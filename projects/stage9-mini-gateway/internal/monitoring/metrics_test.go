package monitoring

import "testing"

func TestErrorRate(t *testing.T) {
	m := &Metrics{}
	m.IncRequest()
	m.IncRequest()
	m.IncError()
	if m.ErrorRate() != 0.5 {
		t.Fatalf("unexpected error rate: %f", m.ErrorRate())
	}
}
