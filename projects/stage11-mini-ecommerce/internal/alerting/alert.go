package alerting

import (
	"fmt"
	"sync/atomic"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/monitoring"
)

// Rule 告警规则。
type Rule struct {
	MaxFailedOrders int64
}

// Evaluate 评估告警。
func Evaluate(m *monitoring.Metrics, r Rule) []string {
	alerts := make([]string, 0)
	failed := atomic.LoadInt64(&m.OrdersFailed)
	if failed > r.MaxFailedOrders {
		alerts = append(alerts, fmt.Sprintf("failed orders too high: %d", failed))
	}
	return alerts
}
