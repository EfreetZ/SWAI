package alerting

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/monitoring"
)

func TestEvaluate(t *testing.T) {
	alerts := Evaluate(monitoring.BrokerMetrics{LagTotal: 100}, Rule{LagThreshold: 50})
	if len(alerts) != 1 {
		t.Fatalf("expected one alert, got %d", len(alerts))
	}

	alerts = Evaluate(monitoring.BrokerMetrics{LagTotal: 10}, Rule{LagThreshold: 50})
	if len(alerts) != 0 {
		t.Fatalf("expected no alert, got %d", len(alerts))
	}
}
