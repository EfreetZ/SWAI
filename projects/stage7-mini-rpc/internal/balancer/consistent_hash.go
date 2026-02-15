package balancer

import (
	"hash/crc32"
	"sort"
	"strconv"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
)

// ConsistentHashBalancer 一致性哈希（简化）。
type ConsistentHashBalancer struct {
	replicas int
	keys     []uint32
	ring     map[uint32]*registry.ServiceInstance
}

func NewConsistentHashBalancer(replicas int) *ConsistentHashBalancer {
	if replicas <= 0 {
		replicas = 10
	}
	return &ConsistentHashBalancer{replicas: replicas, ring: make(map[uint32]*registry.ServiceInstance)}
}

func (b *ConsistentHashBalancer) rebuild(instances []*registry.ServiceInstance) {
	b.keys = b.keys[:0]
	b.ring = make(map[uint32]*registry.ServiceInstance)
	for _, ins := range instances {
		for i := 0; i < b.replicas; i++ {
			key := crc32.ChecksumIEEE([]byte(ins.ID + "#" + strconv.Itoa(i)))
			b.keys = append(b.keys, key)
			b.ring[key] = ins
		}
	}
	sort.Slice(b.keys, func(i, j int) bool { return b.keys[i] < b.keys[j] })
}

func (b *ConsistentHashBalancer) Pick(instances []*registry.ServiceInstance) (*registry.ServiceInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoInstance
	}
	b.rebuild(instances)
	return b.ring[b.keys[0]], nil
}
