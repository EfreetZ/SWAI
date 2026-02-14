package replication

import "sync"

// Slave 从节点复制状态。
type Slave struct {
	mu         sync.Mutex
	masterHost string
	masterPort int
	offset     int64
}

// NewSlave 创建从节点状态。
func NewSlave(masterHost string, masterPort int) *Slave {
	return &Slave{masterHost: masterHost, masterPort: masterPort}
}

// Offset 返回当前复制位点。
func (s *Slave) Offset() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.offset
}

// Ack 更新复制位点。
func (s *Slave) Ack(offset int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if offset > s.offset {
		s.offset = offset
	}
}
