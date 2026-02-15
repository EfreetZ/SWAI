package monitoring

import "sync/atomic"

// Metrics 电商核心指标。
type Metrics struct {
	OrdersCreated int64
	OrdersPaid    int64
	OrdersFailed  int64
}

func (m *Metrics) IncCreated() { atomic.AddInt64(&m.OrdersCreated, 1) }
func (m *Metrics) IncPaid()    { atomic.AddInt64(&m.OrdersPaid, 1) }
func (m *Metrics) IncFailed()  { atomic.AddInt64(&m.OrdersFailed, 1) }
