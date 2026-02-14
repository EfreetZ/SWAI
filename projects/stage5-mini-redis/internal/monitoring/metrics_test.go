package monitoring

import "testing"

func TestSnapshot(t *testing.T) {
	m := Snapshot(1000, 200, 50, 4096)
	if m.QPS != 1000 || m.ConnectedConns != 200 {
		t.Fatalf("snapshot mismatch: %+v", m)
	}
	if m.CollectedAt.IsZero() {
		t.Fatal("CollectedAt is zero")
	}
}
