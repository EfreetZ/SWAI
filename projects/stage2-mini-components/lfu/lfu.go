package lfu

import (
	"errors"
)

var ErrInvalidCapacity = errors.New("capacity must be greater than 0")

type node[K comparable, V any] struct {
	key   K
	value V
	freq  int
	prev  *node[K, V]
	next  *node[K, V]
}

type doublyList[K comparable, V any] struct {
	head *node[K, V]
	tail *node[K, V]
	len  int
}

func (l *doublyList[K, V]) pushFront(n *node[K, V]) {
	n.prev = nil
	n.next = l.head
	if l.head != nil {
		l.head.prev = n
	}
	l.head = n
	if l.tail == nil {
		l.tail = n
	}
	l.len++
}

func (l *doublyList[K, V]) remove(n *node[K, V]) {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		l.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		l.tail = n.prev
	}
	n.prev = nil
	n.next = nil
	l.len--
}

func (l *doublyList[K, V]) popTail() *node[K, V] {
	if l.tail == nil {
		return nil
	}
	n := l.tail
	l.remove(n)
	return n
}

// Cache 是 LFU 缓存实现（非并发安全）。
type Cache[K comparable, V any] struct {
	capacity int
	minFreq  int
	items    map[K]*node[K, V]
	freqMap  map[int]*doublyList[K, V]
}

// New 创建 LFU 缓存。
func New[K comparable, V any](capacity int) (*Cache[K, V], error) {
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}
	return &Cache[K, V]{
		capacity: capacity,
		items:    make(map[K]*node[K, V], capacity),
		freqMap:  make(map[int]*doublyList[K, V]),
	}, nil
}

// Get 获取值并增加频次。
func (c *Cache[K, V]) Get(key K) (V, bool) {
	n, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	c.increaseFreq(n)
	return n.value, true
}

// Put 放入缓存并在容量满时淘汰最低频节点。
func (c *Cache[K, V]) Put(key K, value V) {
	if n, ok := c.items[key]; ok {
		n.value = value
		c.increaseFreq(n)
		return
	}

	if len(c.items) >= c.capacity {
		bucket := c.freqMap[c.minFreq]
		evicted := bucket.popTail()
		delete(c.items, evicted.key)
		if bucket.len == 0 {
			delete(c.freqMap, c.minFreq)
		}
	}

	n := &node[K, V]{key: key, value: value, freq: 1}
	if c.freqMap[1] == nil {
		c.freqMap[1] = &doublyList[K, V]{}
	}
	c.freqMap[1].pushFront(n)
	c.items[key] = n
	c.minFreq = 1
}

// Len 返回缓存长度。
func (c *Cache[K, V]) Len() int {
	return len(c.items)
}

func (c *Cache[K, V]) increaseFreq(n *node[K, V]) {
	oldFreq := n.freq
	bucket := c.freqMap[oldFreq]
	bucket.remove(n)
	if bucket.len == 0 {
		delete(c.freqMap, oldFreq)
		if c.minFreq == oldFreq {
			c.minFreq++
		}
	}

	n.freq++
	if c.freqMap[n.freq] == nil {
		c.freqMap[n.freq] = &doublyList[K, V]{}
	}
	c.freqMap[n.freq].pushFront(n)
}
