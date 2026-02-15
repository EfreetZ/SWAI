package monitoring

import "time"

// Snapshot 可观测性聚合快照。
type Snapshot struct {
	Requests  int64
	Errors    int64
	AvgMicros float64
	At        time.Time
}
