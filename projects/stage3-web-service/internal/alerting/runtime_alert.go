package alerting

import "github.com/EfreetZ/SWAI/projects/stage3-web-service/internal/monitoring"

// Alert 告警定义。
type Alert struct {
	Name      string
	Message   string
	Current   float64
	Threshold float64
}

// Rule 告警规则。
type Rule struct {
	MaxGoroutines int
	MaxHeapAlloc  uint64
}

// EvaluateRuntime 评估运行时指标。
func EvaluateRuntime(metrics monitoring.RuntimeMetrics, rule Rule) []Alert {
	alerts := make([]Alert, 0, 2)
	if rule.MaxGoroutines > 0 && metrics.Goroutines > rule.MaxGoroutines {
		alerts = append(alerts, Alert{
			Name:      "goroutines_high",
			Message:   "goroutine count is too high",
			Current:   float64(metrics.Goroutines),
			Threshold: float64(rule.MaxGoroutines),
		})
	}
	if rule.MaxHeapAlloc > 0 && metrics.HeapAlloc > rule.MaxHeapAlloc {
		alerts = append(alerts, Alert{
			Name:      "heap_alloc_high",
			Message:   "heap allocation is too high",
			Current:   float64(metrics.HeapAlloc),
			Threshold: float64(rule.MaxHeapAlloc),
		})
	}
	return alerts
}
