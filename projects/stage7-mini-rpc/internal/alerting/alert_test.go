package alerting

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/monitoring"
)

func TestEvaluate(t *testing.T) {
	alerts := Evaluate(monitoring.RPCMetrics{Requests: 10, Failures: 5}, Rule{ErrorRateThreshold: 0.3})
	if len(alerts) != 1 {
		t.Fatalf("expected alert")
	}
}
