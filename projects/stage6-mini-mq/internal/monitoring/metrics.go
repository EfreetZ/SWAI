package monitoring

import "time"

// BrokerMetrics broker 指标快照。
type BrokerMetrics struct {
	ProducedTotal int64
	ConsumedTotal int64
	LagTotal      int64
	CollectedAt   time.Time
}

// Throughput 计算总吞吐（消息数）。
func (m BrokerMetrics) Throughput() int64 {
	return m.ProducedTotal + m.ConsumedTotal
}

// LagHealthy 判断积压是否健康。
func (m BrokerMetrics) LagHealthy(threshold int64) bool {
	return m.LagTotal <= threshold
}
