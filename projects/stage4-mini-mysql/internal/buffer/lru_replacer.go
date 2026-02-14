package buffer

import "sync"

type FrameID int

// Replacer 页面置换接口。
type Replacer interface {
	Victim() (FrameID, bool)
	Pin(frameID FrameID)
	Unpin(frameID FrameID)
	Size() int
}

type lruNode struct {
	frame FrameID
	prev  *lruNode
	next  *lruNode
}

// LRUReplacer LRU 置换器。
type LRUReplacer struct {
	mu    sync.Mutex
	nodes map[FrameID]*lruNode
	head  *lruNode
	tail  *lruNode
}

// NewLRUReplacer 创建 LRU 置换器。
func NewLRUReplacer() *LRUReplacer {
	return &LRUReplacer{nodes: make(map[FrameID]*lruNode)}
}

// Victim 选择可淘汰页。
func (r *LRUReplacer) Victim() (FrameID, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.tail == nil {
		return 0, false
	}
	victim := r.tail
	r.remove(victim)
	delete(r.nodes, victim.frame)
	return victim.frame, true
}

// Pin 将页标记为不可淘汰。
func (r *LRUReplacer) Pin(frameID FrameID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.nodes[frameID]
	if !ok {
		return
	}
	r.remove(n)
	delete(r.nodes, frameID)
}

// Unpin 将页加入可淘汰集合。
func (r *LRUReplacer) Unpin(frameID FrameID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.nodes[frameID]; exists {
		return
	}
	n := &lruNode{frame: frameID}
	r.nodes[frameID] = n
	r.addToHead(n)
}

// Size 返回可淘汰数量。
func (r *LRUReplacer) Size() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.nodes)
}

func (r *LRUReplacer) addToHead(n *lruNode) {
	n.prev = nil
	n.next = r.head
	if r.head != nil {
		r.head.prev = n
	}
	r.head = n
	if r.tail == nil {
		r.tail = n
	}
}

func (r *LRUReplacer) remove(n *lruNode) {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		r.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		r.tail = n.prev
	}
	n.prev = nil
	n.next = nil
}
