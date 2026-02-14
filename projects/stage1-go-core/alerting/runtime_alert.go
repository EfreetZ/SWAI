package alerting

import "github.com/EfreetZ/SWAI/projects/stage1-go-core/monitoring"

// Severity 表示告警等级。
type Severity string

const (
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// RuntimeRule 定义运行时指标阈值。
type RuntimeRule struct {
	MaxGoroutines int
	MaxHeapAlloc  uint64
	MaxGCCount    uint32
}

// Alert 表示一条告警记录。
type Alert struct {
	Name      string
	Severity  Severity
	Message   string
	Current   float64
	Threshold float64
}

// EvaluateRuntimeAlerts 根据阈值规则评估运行时指标。
func EvaluateRuntimeAlerts(metrics monitoring.RuntimeMetrics, rule RuntimeRule) []Alert {
	alerts := make([]Alert, 0, 3)

	if rule.MaxGoroutines > 0 && metrics.Goroutines > rule.MaxGoroutines {
		alerts = append(alerts, Alert{
			Name:      "goroutines_high",
			Severity:  SeverityWarning,
			Message:   "goroutine count is above threshold",
			Current:   float64(metrics.Goroutines),
			Threshold: float64(rule.MaxGoroutines),
		})
	}

	if rule.MaxHeapAlloc > 0 && metrics.HeapAlloc > rule.MaxHeapAlloc {
		alerts = append(alerts, Alert{
			Name:      "heap_alloc_high",
			Severity:  SeverityCritical,
			Message:   "heap allocation is above threshold",
			Current:   float64(metrics.HeapAlloc),
			Threshold: float64(rule.MaxHeapAlloc),
		})
	}

	if rule.MaxGCCount > 0 && metrics.GCCount > rule.MaxGCCount {
		alerts = append(alerts, Alert{
			Name:      "gc_count_high",
			Severity:  SeverityWarning,
			Message:   "gc count is above threshold",
			Current:   float64(metrics.GCCount),
			Threshold: float64(rule.MaxGCCount),
		})
	}

	return alerts
}
