package transport

import (
	"net"
	"sync"
	"time"
)

// ConnWrapper 连接包装。
type ConnWrapper struct {
	Conn      net.Conn
	CreatedAt time.Time
	UsedAt    time.Time
}

// ConnManager 连接管理。
type ConnManager struct {
	mu    sync.RWMutex
	conns map[string]*ConnWrapper
}

// NewConnManager 创建连接管理。
func NewConnManager() *ConnManager {
	return &ConnManager{conns: make(map[string]*ConnWrapper)}
}

// Put 放入连接。
func (m *ConnManager) Put(addr string, conn net.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[addr] = &ConnWrapper{Conn: conn, CreatedAt: time.Now(), UsedAt: time.Now()}
}

// Get 获取连接。
func (m *ConnManager) Get(addr string) net.Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w := m.conns[addr]
	if w == nil {
		return nil
	}
	return w.Conn
}

// CloseAll 关闭全部连接。
func (m *ConnManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, wrapper := range m.conns {
		_ = wrapper.Conn.Close()
		delete(m.conns, key)
	}
}
