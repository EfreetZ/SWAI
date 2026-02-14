package replication

import "sync"

// Master 主节点复制管理。
type Master struct {
	mu      sync.Mutex
	backlog *Backlog
	slaves  map[string]int64
}

// NewMaster 创建主节点复制管理器。
func NewMaster() *Master {
	return &Master{backlog: NewBacklog(4096), slaves: make(map[string]int64)}
}

// RegisterSlave 注册从节点。
func (m *Master) RegisterSlave(slaveID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.slaves[slaveID] = 0
}

// Broadcast 传播写命令。
func (m *Master) Broadcast(command string) int64 {
	return m.backlog.Append(command)
}

// PullSince 从指定 offset 拉取增量命令。
func (m *Master) PullSince(offset int64) []string {
	return m.backlog.EntriesSince(offset)
}
