package alerting

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage10-observability/internal/metrics"
)

func TestEvaluate(t *testing.T) {
	m := &metrics.REDMetrics{}
	m.IncRequest(1000, true)
	alerts := Evaluate(m, Rule{MaxErrorRate: 0.1, MaxAvgMicros: 500})
	if len(alerts) == 0 {
		t.Fatal("expected alerts")
	}
}
