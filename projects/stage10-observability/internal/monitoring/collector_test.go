package monitoring

import (
	"testing"
	"time"
)

func TestSnapshot(t *testing.T) {
	s := Snapshot{Requests: 1, Errors: 0, AvgMicros: 10, At: time.Now()}
	if s.Requests != 1 {
		t.Fatal("invalid snapshot")
	}
}
