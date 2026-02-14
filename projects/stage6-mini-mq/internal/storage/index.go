package storage

import (
	"errors"
	"sort"
	"sync"
)

// IndexEntry 稀疏索引项。
type IndexEntry struct {
	Offset   int64
	Position int64
}

// SparseIndex 稀疏索引。
type SparseIndex struct {
	mu      sync.RWMutex
	entries []IndexEntry
}

// NewSparseIndex 创建索引。
func NewSparseIndex() *SparseIndex {
	return &SparseIndex{entries: make([]IndexEntry, 0)}
}

// Append 追加索引。
func (i *SparseIndex) Append(offset, position int64) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.entries = append(i.entries, IndexEntry{Offset: offset, Position: position})
}

// Lookup 查找最接近且不大于目标 offset 的位置。
func (i *SparseIndex) Lookup(offset int64) (int64, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if len(i.entries) == 0 {
		return 0, errors.New("empty index")
	}
	idx := sort.Search(len(i.entries), func(j int) bool {
		return i.entries[j].Offset > offset
	})
	if idx == 0 {
		return i.entries[0].Position, nil
	}
	return i.entries[idx-1].Position, nil
}

// Entries 返回索引副本。
func (i *SparseIndex) Entries() []IndexEntry {
	i.mu.RLock()
	defer i.mu.RUnlock()
	res := make([]IndexEntry, len(i.entries))
	copy(res, i.entries)
	return res
}
