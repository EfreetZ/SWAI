package metrics

import "sync/atomic"

// REDMetrics RED 指标。
type REDMetrics struct {
	requestTotal int64
	errorTotal   int64
	durationMic  int64
}

// IncRequest 记录请求与耗时。
func (m *REDMetrics) IncRequest(durationMicros int64, hasErr bool) {
	atomic.AddInt64(&m.requestTotal, 1)
	atomic.AddInt64(&m.durationMic, durationMicros)
	if hasErr {
		atomic.AddInt64(&m.errorTotal, 1)
	}
}

// Snapshot 读取快照。
func (m *REDMetrics) Snapshot() (requests int64, errors int64, avgDurationMicros float64) {
	requests = atomic.LoadInt64(&m.requestTotal)
	errors = atomic.LoadInt64(&m.errorTotal)
	totalDuration := atomic.LoadInt64(&m.durationMic)
	if requests == 0 {
		return requests, errors, 0
	}
	return requests, errors, float64(totalDuration) / float64(requests)
}

// ErrorRate 错误率。
func (m *REDMetrics) ErrorRate() float64 {
	req, errCnt, _ := m.Snapshot()
	if req == 0 {
		return 0
	}
	return float64(errCnt) / float64(req)
}
