package alerting

import "github.com/EfreetZ/SWAI/projects/stage2-mini-components/monitoring"

// Severity 告警等级。
type Severity string

const (
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Alert 告警对象。
type Alert struct {
	Name      string
	Severity  Severity
	Message   string
	Current   float64
	Threshold float64
}

// Rule 告警阈值。
type Rule struct {
	MinHitRatio    float64
	MaxQueueLagMS  int64
	MaxPendingTask uint64
}

// Evaluate 评估组件指标并生成告警。
func Evaluate(metrics monitoring.ComponentMetrics, rule Rule) []Alert {
	alerts := make([]Alert, 0, 3)

	hitRatio := metrics.Cache.HitRatio()
	if rule.MinHitRatio > 0 && hitRatio < rule.MinHitRatio {
		alerts = append(alerts, Alert{
			Name:      "cache_hit_ratio_low",
			Severity:  SeverityWarning,
			Message:   "cache hit ratio below threshold",
			Current:   hitRatio,
			Threshold: rule.MinHitRatio,
		})
	}

	if rule.MaxQueueLagMS > 0 && metrics.Queue.LagMillis > rule.MaxQueueLagMS {
		alerts = append(alerts, Alert{
			Name:      "queue_lag_high",
			Severity:  SeverityCritical,
			Message:   "queue lag exceeds threshold",
			Current:   float64(metrics.Queue.LagMillis),
			Threshold: float64(rule.MaxQueueLagMS),
		})
	}

	if rule.MaxPendingTask > 0 && metrics.Queue.PendingTasks > rule.MaxPendingTask {
		alerts = append(alerts, Alert{
			Name:      "queue_pending_high",
			Severity:  SeverityWarning,
			Message:   "pending tasks exceed threshold",
			Current:   float64(metrics.Queue.PendingTasks),
			Threshold: float64(rule.MaxPendingTask),
		})
	}

	return alerts
}
