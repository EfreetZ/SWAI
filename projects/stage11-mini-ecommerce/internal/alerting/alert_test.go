package alerting

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/monitoring"
)

func TestEvaluate(t *testing.T) {
	m := &monitoring.Metrics{}
	m.IncFailed()
	alerts := Evaluate(m, Rule{MaxFailedOrders: 0})
	if len(alerts) != 1 {
		t.Fatalf("expected alert")
	}
}
