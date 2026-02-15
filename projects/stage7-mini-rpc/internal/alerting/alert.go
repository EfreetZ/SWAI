package alerting

import (
	"fmt"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/monitoring"
)

// Rule 告警规则。
type Rule struct {
	ErrorRateThreshold float64
}

// Evaluate 评估告警。
func Evaluate(metrics monitoring.RPCMetrics, rule Rule) []string {
	alerts := make([]string, 0)
	if metrics.ErrorRate() > rule.ErrorRateThreshold {
		alerts = append(alerts, fmt.Sprintf("rpc error rate too high: %.4f", metrics.ErrorRate()))
	}
	return alerts
}
