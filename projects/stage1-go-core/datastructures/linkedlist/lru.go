package linkedlist

import (
	"errors"
	"sync"
)

var ErrInvalidCapacity = errors.New("capacity must be greater than 0")

type node[K comparable, V any] struct {
	key   K
	value V
	prev  *node[K, V]
	next  *node[K, V]
}

// LRUCache 基于双向链表 + 哈希表实现固定容量缓存。
type LRUCache[K comparable, V any] struct {
	capacity int
	items    map[K]*node[K, V]
	head     *node[K, V]
	tail     *node[K, V]
	mu       sync.Mutex
}

// NewLRUCache 创建一个固定容量的 LRU 缓存。
func NewLRUCache[K comparable, V any](capacity int) (*LRUCache[K, V], error) {
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	return &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*node[K, V], capacity),
	}, nil
}

// Put 写入缓存，若 key 已存在则更新并提升为最近使用。
func (c *LRUCache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n, ok := c.items[key]; ok {
		n.value = value
		c.moveToFront(n)
		return
	}

	n := &node[K, V]{key: key, value: value}
	c.items[key] = n
	c.addToFront(n)

	if len(c.items) > c.capacity {
		c.evictTail()
	}
}

// Get 读取缓存，命中时会提升为最近使用。
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	n, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}

	c.moveToFront(n)
	return n.value, true
}

// Len 返回当前缓存条目数量。
func (c *LRUCache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

func (c *LRUCache[K, V]) addToFront(n *node[K, V]) {
	n.prev = nil
	n.next = c.head
	if c.head != nil {
		c.head.prev = n
	}
	c.head = n
	if c.tail == nil {
		c.tail = n
	}
}

func (c *LRUCache[K, V]) remove(n *node[K, V]) {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		c.head = n.next
	}

	if n.next != nil {
		n.next.prev = n.prev
	} else {
		c.tail = n.prev
	}

	n.prev = nil
	n.next = nil
}

func (c *LRUCache[K, V]) moveToFront(n *node[K, V]) {
	if c.head == n {
		return
	}
	c.remove(n)
	c.addToFront(n)
}

func (c *LRUCache[K, V]) evictTail() {
	if c.tail == nil {
		return
	}
	evict := c.tail
	c.remove(evict)
	delete(c.items, evict.key)
}
