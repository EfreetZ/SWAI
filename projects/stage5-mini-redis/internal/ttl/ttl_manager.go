package ttl

import (
	"context"
	"sync"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/db"
)

// Manager TTL 管理器。
type Manager struct {
	mu   sync.Mutex
	heap *Heap
	db   *db.DB
}

// NewManager 创建 TTL 管理器。
func NewManager(database *db.DB) *Manager {
	return &Manager{heap: NewHeap(), db: database}
}

// Add 添加过期 key。
func (m *Manager) Add(key string, expireAt time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.heap.Push(key, expireAt.UnixNano())
}

// ActiveExpire 主动过期。
func (m *Manager) ActiveExpire(ctx context.Context, sampleLimit int) {
	if ctx == nil {
		ctx = context.Background()
	}
	if sampleLimit <= 0 {
		sampleLimit = 20
	}
	m.db.ActiveExpire(ctx, sampleLimit)
}

// Start 启动定期过期任务。
func (m *Manager) Start(ctx context.Context, interval time.Duration) {
	if ctx == nil {
		ctx = context.Background()
	}
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.ActiveExpire(ctx, 20)
		}
	}
}
