package monitoring

import "testing"

func TestMetrics(t *testing.T) {
	m := &Metrics{}
	m.IncCreated()
	m.IncPaid()
	m.IncFailed()
}
