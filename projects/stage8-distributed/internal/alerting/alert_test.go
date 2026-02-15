package alerting

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage8-distributed/internal/monitoring"
)

func TestEvaluate(t *testing.T) {
	alerts := Evaluate(monitoring.Metrics{LockContention: 10}, Rule{MaxLockContention: 3})
	if len(alerts) != 1 {
		t.Fatalf("expected alert, got %d", len(alerts))
	}
}
