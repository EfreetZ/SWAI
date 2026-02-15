package raft

import "context"

// Cluster 简化 Raft 集群。
type Cluster struct {
	nodes map[string]*Node
}

// NewCluster 创建集群。
func NewCluster(ids []string) *Cluster {
	nodes := make(map[string]*Node, len(ids))
	for _, id := range ids {
		nodes[id] = NewNode(Config{ID: id, Peers: ids})
	}
	return &Cluster{nodes: nodes}
}

// ElectLeader 选主（简化：固定第一个节点）。
func (c *Cluster) ElectLeader() *Node {
	for _, n := range c.nodes {
		n.BecomeLeader()
		return n
	}
	return nil
}

// ReplicateKV 复制写（简化：写入所有节点）。
func (c *Cluster) ReplicateKV(ctx context.Context, key, value string) error {
	for _, n := range c.nodes {
		if err := n.ApplyKV(ctx, key, value); err != nil {
			return err
		}
	}
	return nil
}

// Node 获取节点。
func (c *Cluster) Node(id string) *Node {
	return c.nodes[id]
}
