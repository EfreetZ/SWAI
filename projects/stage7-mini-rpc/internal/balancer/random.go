package balancer

import (
	"math/rand"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
)

// RandomBalancer 随机。
type RandomBalancer struct {
	rnd *rand.Rand
}

func NewRandomBalancer() *RandomBalancer {
	return &RandomBalancer{rnd: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (b *RandomBalancer) Pick(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoInstance
	}
	idx := b.rnd.Intn(len(instances))
	return instances[idx], nil
}
