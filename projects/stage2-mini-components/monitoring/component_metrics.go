package monitoring

import "time"

// CacheMetrics 缓存指标。
type CacheMetrics struct {
	Hits      uint64
	Misses    uint64
	Collected time.Time
}

// HitRatio 计算命中率。
func (m CacheMetrics) HitRatio() float64 {
	total := m.Hits + m.Misses
	if total == 0 {
		return 0
	}
	return float64(m.Hits) / float64(total)
}

// QueueMetrics 队列指标。
type QueueMetrics struct {
	PendingTasks uint64
	LagMillis    int64
	Collected    time.Time
}

// ComponentMetrics 聚合组件指标。
type ComponentMetrics struct {
	Cache CacheMetrics
	Queue QueueMetrics
}
