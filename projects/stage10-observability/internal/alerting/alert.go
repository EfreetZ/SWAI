package alerting

import (
	"fmt"

	"github.com/EfreetZ/SWAI/projects/stage10-observability/internal/metrics"
)

// Rule 告警规则。
type Rule struct {
	MaxErrorRate float64
	MaxAvgMicros float64
}

// Evaluate 评估告警。
func Evaluate(red *metrics.REDMetrics, rule Rule) []string {
	alerts := make([]string, 0)
	_, _, avg := red.Snapshot()
	errRate := red.ErrorRate()
	if errRate > rule.MaxErrorRate {
		alerts = append(alerts, fmt.Sprintf("high error rate: %.4f", errRate))
	}
	if avg > rule.MaxAvgMicros {
		alerts = append(alerts, fmt.Sprintf("high avg latency: %.2f", avg))
	}
	return alerts
}
