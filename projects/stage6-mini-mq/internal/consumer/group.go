package consumer

import (
	"sync"
	"time"
)

// Member 消费组成员。
type Member struct {
	ID            string
	LastHeartbeat time.Time
	Partitions    []TopicPartition
}

// Group 消费组。
type Group struct {
	ID         string
	Generation int
	Members    map[string]*Member
}

// GroupCoordinator 消费组协调器。
type GroupCoordinator struct {
	mu     sync.Mutex
	groups map[string]*Group
}

// NewGroupCoordinator 创建协调器。
func NewGroupCoordinator() *GroupCoordinator {
	return &GroupCoordinator{groups: make(map[string]*Group)}
}

// Join 加入消费组。
func (c *GroupCoordinator) Join(groupID, memberID string) *Group {
	c.mu.Lock()
	defer c.mu.Unlock()
	group, ok := c.groups[groupID]
	if !ok {
		group = &Group{ID: groupID, Generation: 1, Members: make(map[string]*Member)}
		c.groups[groupID] = group
	}
	if _, exists := group.Members[memberID]; !exists {
		group.Generation++
	}
	group.Members[memberID] = &Member{ID: memberID, LastHeartbeat: time.Now()}
	return group
}

// Leave 离开消费组。
func (c *GroupCoordinator) Leave(groupID, memberID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	group, ok := c.groups[groupID]
	if !ok {
		return
	}
	delete(group.Members, memberID)
	group.Generation++
}

// Heartbeat 更新心跳。
func (c *GroupCoordinator) Heartbeat(groupID, memberID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	group, ok := c.groups[groupID]
	if !ok {
		return
	}
	member, exists := group.Members[memberID]
	if !exists {
		return
	}
	member.LastHeartbeat = time.Now()
}

// Rebalance 执行分区重平衡。
func (c *GroupCoordinator) Rebalance(groupID string, partitions []TopicPartition, strategy AssignmentStrategy) map[string][]TopicPartition {
	c.mu.Lock()
	defer c.mu.Unlock()
	group, ok := c.groups[groupID]
	if !ok {
		return map[string][]TopicPartition{}
	}
	members := make([]string, 0, len(group.Members))
	for memberID := range group.Members {
		members = append(members, memberID)
	}
	assigned := strategy.Assign(members, partitions)
	for memberID, tps := range assigned {
		if member := group.Members[memberID]; member != nil {
			member.Partitions = tps
		}
	}
	group.Generation++
	return assigned
}
