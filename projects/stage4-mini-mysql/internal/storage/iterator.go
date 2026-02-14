package storage

// KeyValue 键值对。
type KeyValue struct {
	Key   []byte
	Value []byte
}

// Iterator 范围扫描迭代器。
type Iterator struct {
	items []KeyValue
	index int
}

// NewIterator 创建迭代器。
func NewIterator(items []KeyValue) *Iterator {
	return &Iterator{items: items, index: -1}
}

// Next 移动到下一个元素。
func (it *Iterator) Next() bool {
	it.index++
	return it.index >= 0 && it.index < len(it.items)
}

// Item 返回当前元素。
func (it *Iterator) Item() KeyValue {
	if it.index < 0 || it.index >= len(it.items) {
		return KeyValue{}
	}
	return it.items[it.index]
}
