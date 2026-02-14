package alerting

import "github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/monitoring"

// Alert 告警定义。
type Alert struct {
	Name      string
	Message   string
	Current   float64
	Threshold float64
}

// Rule 告警规则。
type Rule struct {
	MinBufferHitRatio float64
	MaxWALSizeBytes   uint64
}

// EvaluateStorage 评估存储告警。
func EvaluateStorage(metrics monitoring.StorageMetrics, rule Rule) []Alert {
	alerts := make([]Alert, 0, 2)
	if rule.MinBufferHitRatio > 0 && metrics.BufferHitRatio() < rule.MinBufferHitRatio {
		alerts = append(alerts, Alert{
			Name:      "buffer_hit_ratio_low",
			Message:   "buffer hit ratio is below threshold",
			Current:   metrics.BufferHitRatio(),
			Threshold: rule.MinBufferHitRatio,
		})
	}
	if rule.MaxWALSizeBytes > 0 && metrics.WALSizeBytes > rule.MaxWALSizeBytes {
		alerts = append(alerts, Alert{
			Name:      "wal_size_high",
			Message:   "wal size exceeds threshold",
			Current:   float64(metrics.WALSizeBytes),
			Threshold: float64(rule.MaxWALSizeBytes),
		})
	}
	return alerts
}
