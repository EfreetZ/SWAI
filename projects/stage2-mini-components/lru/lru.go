package lru

import "errors"

var ErrInvalidCapacity = errors.New("capacity must be greater than 0")

type node[K comparable, V any] struct {
	key   K
	value V
	prev  *node[K, V]
	next  *node[K, V]
}

// Cache 是基础 LRU 缓存实现（非并发安全）。
type Cache[K comparable, V any] struct {
	capacity int
	items    map[K]*node[K, V]
	head     *node[K, V]
	tail     *node[K, V]
}

// New 创建基础 LRU 缓存。
func New[K comparable, V any](capacity int) (*Cache[K, V], error) {
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	return &Cache[K, V]{
		capacity: capacity,
		items:    make(map[K]*node[K, V], capacity),
	}, nil
}

// Get 读取缓存并将节点移动到头部。
func (c *Cache[K, V]) Get(key K) (V, bool) {
	n, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	c.moveToHead(n)
	return n.value, true
}

// Put 写入缓存，超出容量会淘汰最久未使用节点。
func (c *Cache[K, V]) Put(key K, value V) {
	if n, ok := c.items[key]; ok {
		n.value = value
		c.moveToHead(n)
		return
	}

	n := &node[K, V]{key: key, value: value}
	c.items[key] = n
	c.addToHead(n)

	if len(c.items) > c.capacity {
		c.removeTail()
	}
}

// Len 返回缓存长度。
func (c *Cache[K, V]) Len() int {
	return len(c.items)
}

func (c *Cache[K, V]) moveToHead(n *node[K, V]) {
	if c.head == n {
		return
	}
	c.removeNode(n)
	c.addToHead(n)
}

func (c *Cache[K, V]) removeNode(n *node[K, V]) {
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

func (c *Cache[K, V]) addToHead(n *node[K, V]) {
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

func (c *Cache[K, V]) removeTail() {
	if c.tail == nil {
		return
	}
	last := c.tail
	c.removeNode(last)
	delete(c.items, last.key)
}
