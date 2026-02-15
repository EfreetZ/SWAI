package metrics

import "testing"

func TestSnapshot(t *testing.T) {
	m := &REDMetrics{}
	m.IncRequest(100, false)
	m.IncRequest(200, true)
	req, errCnt, avg := m.Snapshot()
	if req != 2 || errCnt != 1 {
		t.Fatalf("unexpected snapshot: %d %d", req, errCnt)
	}
	if avg != 150 {
		t.Fatalf("unexpected avg: %f", avg)
	}
}
