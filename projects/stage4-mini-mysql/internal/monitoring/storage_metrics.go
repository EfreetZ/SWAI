package monitoring

import "time"

// StorageMetrics 存储相关指标。
type StorageMetrics struct {
	BufferHits   uint64
	BufferMisses uint64
	WALSizeBytes uint64
	CollectedAt  time.Time
}

// BufferHitRatio 返回 Buffer Pool 命中率。
func (m StorageMetrics) BufferHitRatio() float64 {
	total := m.BufferHits + m.BufferMisses
	if total == 0 {
		return 0
	}
	return float64(m.BufferHits) / float64(total)
}
