package cluster

import (
	"errors"
	"hash/crc32"
)

var ErrMoved = errors.New("moved")

// Node 集群节点。
type Node struct {
	ID    string
	Addr  string
	Slots []SlotRange
}

// State 集群状态。
type State struct {
	Nodes   map[string]*Node
	SlotMap [TotalSlots]*Node
}

// NewState 创建集群状态。
func NewState() *State {
	return &State{Nodes: make(map[string]*Node)}
}

// AddNode 添加节点并建立 slot 映射。
func (s *State) AddNode(node *Node) {
	if node == nil {
		return
	}
	s.Nodes[node.ID] = node
	for _, r := range node.Slots {
		for slot := r.Start; slot <= r.End && slot < TotalSlots; slot++ {
			s.SlotMap[slot] = node
		}
	}
}

// ResolveNode 根据 key 定位节点。
func (s *State) ResolveNode(key string) (*Node, uint16, error) {
	slot := KeyToSlot(key)
	node := s.SlotMap[slot]
	if node == nil {
		return nil, slot, ErrMoved
	}
	return node, slot, nil
}

// KeyToSlot 计算 key slot。
func KeyToSlot(key string) uint16 {
	return uint16(crc32.ChecksumIEEE([]byte(key)) % TotalSlots)
}
