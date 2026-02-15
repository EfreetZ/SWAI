package balancer

import (
	"sync/atomic"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
)

// RoundRobinBalancer 轮询。
type RoundRobinBalancer struct {
	counter uint64
}

func (b *RoundRobinBalancer) Pick(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoInstance
	}
	idx := atomic.AddUint64(&b.counter, 1)
	return instances[idx%uint64(len(instances))], nil
}
