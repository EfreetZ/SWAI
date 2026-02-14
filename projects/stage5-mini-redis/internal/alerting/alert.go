package alerting

import "github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/monitoring"

// Alert 告警对象。
type Alert struct {
	Name      string
	Message   string
	Current   float64
	Threshold float64
}

// Rule 告警规则。
type Rule struct {
	MaxQPS          uint64
	MaxConnected    int
	MaxAOFSizeBytes uint64
}

// Evaluate 评估监控指标并产生告警。
func Evaluate(metrics monitoring.Metrics, rule Rule) []Alert {
	alerts := make([]Alert, 0, 3)
	if rule.MaxQPS > 0 && metrics.QPS > rule.MaxQPS {
		alerts = append(alerts, Alert{Name: "qps_high", Message: "qps exceeds threshold", Current: float64(metrics.QPS), Threshold: float64(rule.MaxQPS)})
	}
	if rule.MaxConnected > 0 && metrics.ConnectedConns > rule.MaxConnected {
		alerts = append(alerts, Alert{Name: "conns_high", Message: "connected clients exceed threshold", Current: float64(metrics.ConnectedConns), Threshold: float64(rule.MaxConnected)})
	}
	if rule.MaxAOFSizeBytes > 0 && metrics.AOFSizeBytes > rule.MaxAOFSizeBytes {
		alerts = append(alerts, Alert{Name: "aof_size_high", Message: "aof size exceeds threshold", Current: float64(metrics.AOFSizeBytes), Threshold: float64(rule.MaxAOFSizeBytes)})
	}
	return alerts
}
