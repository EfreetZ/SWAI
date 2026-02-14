package replication

import "sync"

// Backlog 复制积压缓冲区（环形）。
type Backlog struct {
	mu       sync.Mutex
	capacity int
	offset   int64
	entries  []string
}

// NewBacklog 创建 backlog。
func NewBacklog(capacity int) *Backlog {
	if capacity <= 0 {
		capacity = 1024
	}
	return &Backlog{capacity: capacity, entries: make([]string, 0, capacity)}
}

// Append 追加命令。
func (b *Backlog) Append(command string) int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.offset++
	if len(b.entries) >= b.capacity {
		b.entries = b.entries[1:]
	}
	b.entries = append(b.entries, command)
	return b.offset
}

// EntriesSince 返回 offset 之后的命令。
func (b *Backlog) EntriesSince(offset int64) []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if offset >= b.offset {
		return nil
	}
	result := make([]string, 0)
	start := int(offset - (b.offset - int64(len(b.entries))))
	if start < 0 {
		start = 0
	}
	for i := start; i < len(b.entries); i++ {
		result = append(result, b.entries[i])
	}
	return result
}
