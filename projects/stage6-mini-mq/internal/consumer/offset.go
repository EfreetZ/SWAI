package consumer

import "sync"

// OffsetManager offset 管理。
type OffsetManager struct {
	mu      sync.RWMutex
	offsets map[string]int64
}

// NewOffsetManager 创建 offset 管理器。
func NewOffsetManager() *OffsetManager {
	return &OffsetManager{offsets: make(map[string]int64)}
}

func offsetKey(group, topic string, partition int) string {
	return group + ":" + topic + ":" + strconvItoa(partition)
}

// Commit 提交 offset。
func (m *OffsetManager) Commit(group, topic string, partition int, offset int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.offsets[offsetKey(group, topic, partition)] = offset
}

// Get 获取 offset。
func (m *OffsetManager) Get(group, topic string, partition int) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.offsets[offsetKey(group, topic, partition)]
}

func strconvItoa(v int) string {
	if v == 0 {
		return "0"
	}
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	buf := make([]byte, 0, 12)
	for v > 0 {
		buf = append([]byte{byte('0' + v%10)}, buf...)
		v /= 10
	}
	return sign + string(buf)
}
