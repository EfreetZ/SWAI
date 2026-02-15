package monitoring

import "time"

// RPCMetrics RPC 指标。
type RPCMetrics struct {
	Requests    int64
	Failures    int64
	CollectedAt time.Time
}

// ErrorRate 计算错误率。
func (m RPCMetrics) ErrorRate() float64 {
	if m.Requests == 0 {
		return 0
	}
	return float64(m.Failures) / float64(m.Requests)
}
