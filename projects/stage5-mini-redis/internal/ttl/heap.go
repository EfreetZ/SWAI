package ttl

import "container/heap"

type item struct {
	key      string
	expireAt int64
	index    int
}

type minHeap []*item

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return h[i].expireAt < h[j].expireAt }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }
func (h *minHeap) Push(x any)        { v := x.(*item); v.index = len(*h); *h = append(*h, v) }
func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	v := old[n-1]
	v.index = -1
	*h = old[:n-1]
	return v
}
func (h minHeap) Peek() *item {
	if len(h) == 0 {
		return nil
	}
	return h[0]
}

// Heap 过期最小堆。
type Heap struct {
	h minHeap
}

// NewHeap 创建最小堆。
func NewHeap() *Heap {
	h := minHeap{}
	heap.Init(&h)
	return &Heap{h: h}
}

// Push 插入元素。
func (h *Heap) Push(key string, expireAt int64) {
	heap.Push(&h.h, &item{key: key, expireAt: expireAt})
}

// Pop 弹出最早过期元素。
func (h *Heap) Pop() (string, int64, bool) {
	if len(h.h) == 0 {
		return "", 0, false
	}
	v := heap.Pop(&h.h).(*item)
	return v.key, v.expireAt, true
}

// Peek 查看堆顶元素。
func (h *Heap) Peek() (string, int64, bool) {
	v := h.h.Peek()
	if v == nil {
		return "", 0, false
	}
	return v.key, v.expireAt, true
}

// Len 返回元素数量。
func (h *Heap) Len() int { return len(h.h) }
