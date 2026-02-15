package balancer

import (
	"sync"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
)

// LeastConnBalancer 最少连接。
type LeastConnBalancer struct {
	mu    sync.RWMutex
	conns map[string]int64
}

func NewLeastConnBalancer() *LeastConnBalancer {
	return &LeastConnBalancer{conns: make(map[string]int64)}
}

func (b *LeastConnBalancer) Pick(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoInstance
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	var selected *registry.ServiceInstance
	var min int64
	for i, ins := range instances {
		count := b.conns[ins.ID]
		if i == 0 || count < min {
			selected = ins
			min = count
		}
	}
	if selected == nil {
		return nil, ErrNoInstance
	}
	b.conns[selected.ID]++
	return selected, nil
}

func (b *LeastConnBalancer) Done(instanceID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.conns[instanceID] > 0 {
		b.conns[instanceID]--
	}
}
