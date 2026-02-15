package monitoring

import "sync/atomic"

// Metrics 网关指标。
type Metrics struct {
	TotalRequests int64
	TotalErrors   int64
}

// IncRequest 增加请求。
func (m *Metrics) IncRequest() {
	atomic.AddInt64(&m.TotalRequests, 1)
}

// IncError 增加错误。
func (m *Metrics) IncError() {
	atomic.AddInt64(&m.TotalErrors, 1)
}

// ErrorRate 返回错误率。
func (m *Metrics) ErrorRate() float64 {
	total := atomic.LoadInt64(&m.TotalRequests)
	if total == 0 {
		return 0
	}
	errCnt := atomic.LoadInt64(&m.TotalErrors)
	return float64(errCnt) / float64(total)
}
