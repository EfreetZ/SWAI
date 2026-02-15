package alerting

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/monitoring"
)

func TestEvaluate(t *testing.T) {
	m := &monitoring.Metrics{}
	m.IncRequest()
	m.IncError()
	alerts := Evaluate(m, Rule{MaxErrorRate: 0.2})
	if len(alerts) != 1 {
		t.Fatalf("expected alert")
	}
}
