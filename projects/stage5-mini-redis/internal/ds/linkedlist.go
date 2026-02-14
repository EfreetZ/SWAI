package ds

type listNode struct {
	value string
	prev  *listNode
	next  *listNode
}

// LinkedList 双向链表。
type LinkedList struct {
	head *listNode
	tail *listNode
	len  int
}

// LPush 头插。
func (l *LinkedList) LPush(value string) {
	n := &listNode{value: value, next: l.head}
	if l.head != nil {
		l.head.prev = n
	}
	l.head = n
	if l.tail == nil {
		l.tail = n
	}
	l.len++
}

// RPush 尾插。
func (l *LinkedList) RPush(value string) {
	n := &listNode{value: value, prev: l.tail}
	if l.tail != nil {
		l.tail.next = n
	}
	l.tail = n
	if l.head == nil {
		l.head = n
	}
	l.len++
}

// LPop 头删。
func (l *LinkedList) LPop() (string, bool) {
	if l.head == nil {
		return "", false
	}
	n := l.head
	l.head = n.next
	if l.head != nil {
		l.head.prev = nil
	} else {
		l.tail = nil
	}
	l.len--
	return n.value, true
}

// Range 获取区间元素。
func (l *LinkedList) Range(start, stop int) []string {
	if start < 0 {
		start = 0
	}
	if stop < start {
		return nil
	}
	result := make([]string, 0)
	idx := 0
	for node := l.head; node != nil; node = node.next {
		if idx >= start && idx <= stop {
			result = append(result, node.value)
		}
		if idx > stop {
			break
		}
		idx++
	}
	return result
}
