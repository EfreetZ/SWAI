package broker

import "sync"

// GroupCoordinator broker 侧 group 协调器（简化）。
type GroupCoordinator struct {
	mu     sync.Mutex
	groups map[string]map[string]struct{}
}

// NewGroupCoordinator 创建 group 协调器。
func NewGroupCoordinator() *GroupCoordinator {
	return &GroupCoordinator{groups: make(map[string]map[string]struct{})}
}

// Join 会员加入组。
func (c *GroupCoordinator) Join(groupID, memberID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	group := c.groups[groupID]
	if group == nil {
		group = make(map[string]struct{})
		c.groups[groupID] = group
	}
	group[memberID] = struct{}{}
}

// Leave 会员离开组。
func (c *GroupCoordinator) Leave(groupID, memberID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	group := c.groups[groupID]
	if group == nil {
		return
	}
	delete(group, memberID)
}

// Members 返回组内成员。
func (c *GroupCoordinator) Members(groupID string) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	group := c.groups[groupID]
	members := make([]string, 0, len(group))
	for memberID := range group {
		members = append(members, memberID)
	}
	return members
}
