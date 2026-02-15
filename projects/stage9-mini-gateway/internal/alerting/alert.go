package alerting

import (
	"fmt"

	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/monitoring"
)

// Rule 告警规则。
type Rule struct {
	MaxErrorRate float64
}

// Evaluate 评估告警。
func Evaluate(m *monitoring.Metrics, r Rule) []string {
	alerts := make([]string, 0)
	if m.ErrorRate() > r.MaxErrorRate {
		alerts = append(alerts, fmt.Sprintf("gateway error rate high: %.4f", m.ErrorRate()))
	}
	return alerts
}
