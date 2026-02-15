package alerting

import (
	"fmt"

	"github.com/EfreetZ/SWAI/projects/stage8-distributed/internal/monitoring"
)

// Rule 告警规则。
type Rule struct {
	MaxLockContention int64
}

// Evaluate 评估告警。
func Evaluate(m monitoring.Metrics, rule Rule) []string {
	alerts := make([]string, 0)
	if m.LockContention > rule.MaxLockContention {
		alerts = append(alerts, fmt.Sprintf("lock contention high: %d", m.LockContention))
	}
	return alerts
}
