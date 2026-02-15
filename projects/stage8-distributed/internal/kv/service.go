package kv

import (
	"context"
	"errors"
	"sync"

	"github.com/EfreetZ/SWAI/projects/stage8-distributed/internal/raft"
)

var ErrNotLeader = errors.New("not leader")

// Service 上层 KV 服务。
type Service struct {
	cluster *raft.Cluster
	leader  *raft.Node
	mu      sync.RWMutex
}

// NewService 创建 KV 服务。
func NewService(cluster *raft.Cluster) *Service {
	leader := cluster.ElectLeader()
	return &Service{cluster: cluster, leader: leader}
}

// Put 写入 KV。
func (s *Service) Put(ctx context.Context, key, value string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.RLock()
	leader := s.leader
	s.mu.RUnlock()
	if leader == nil || leader.State() != raft.Leader {
		return ErrNotLeader
	}
	return s.cluster.ReplicateKV(ctx, key, value)
}

// Get 读取 KV。
func (s *Service) Get(key string) (string, bool) {
	s.mu.RLock()
	leader := s.leader
	s.mu.RUnlock()
	if leader == nil {
		return "", false
	}
	return leader.GetKV(key)
}

// Delete 删除 KV。
func (s *Service) Delete(ctx context.Context, key string) error {
	return s.Put(ctx, key, "")
}
