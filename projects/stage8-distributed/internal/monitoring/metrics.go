package monitoring

import "time"

// Metrics 分布式系统指标。
type Metrics struct {
	LeaderChanges  int64
	ApplyOps       int64
	LockContention int64
	CollectedAt    time.Time
}

// Healthy 判断是否健康。
func (m Metrics) Healthy(maxContention int64) bool {
	return m.LockContention <= maxContention
}
