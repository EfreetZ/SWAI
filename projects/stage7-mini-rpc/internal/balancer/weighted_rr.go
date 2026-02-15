package balancer

import "github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"

// WeightedRRBalancer 加权轮询（简化实现）。
type WeightedRRBalancer struct {
	idx int
}

func (b *WeightedRRBalancer) Pick(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoInstance
	}
	expanded := make([]*registry.ServiceInstance, 0)
	for _, ins := range instances {
		w := ins.Weight
		if w <= 0 {
			w = 1
		}
		for i := 0; i < w; i++ {
			expanded = append(expanded, ins)
		}
	}
	if len(expanded) == 0 {
		return nil, ErrNoInstance
	}
	b.idx = (b.idx + 1) % len(expanded)
	return expanded[b.idx], nil
}
