package monitoring

import "time"

// Metrics Redis 关键指标。
type Metrics struct {
	QPS            uint64
	ConnectedConns int
	HotKeyOps      uint64
	AOFSizeBytes   uint64
	CollectedAt    time.Time
}

// Snapshot 创建指标快照。
func Snapshot(qps uint64, conns int, hotKeyOps uint64, aofSize uint64) Metrics {
	return Metrics{QPS: qps, ConnectedConns: conns, HotKeyOps: hotKeyOps, AOFSizeBytes: aofSize, CollectedAt: time.Now()}
}
