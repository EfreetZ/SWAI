package alerting

import (
	"fmt"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/monitoring"
)

// Rule 告警阈值规则。
type Rule struct {
	LagThreshold int64
}

// Evaluate 评估告警。
func Evaluate(metrics monitoring.BrokerMetrics, rule Rule) []string {
	alerts := make([]string, 0)
	if metrics.LagTotal > rule.LagThreshold {
		alerts = append(alerts, fmt.Sprintf("consumer lag too high: %d", metrics.LagTotal))
	}
	return alerts
}
