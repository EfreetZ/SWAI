package storage

import (
	"bytes"
	"context"
	"errors"
	"sort"
	"sync"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

// BPlusTree 简化 B+Tree（内存索引 + 有序键集合）。
type BPlusTree struct {
	mu    sync.RWMutex
	order int
	pager PageManager
	store map[string][]byte
	keys  []string
}

// NewBPlusTree 创建简化 B+Tree。
func NewBPlusTree(order int, pager PageManager) *BPlusTree {
	if order <= 0 {
		order = 16
	}
	return &BPlusTree{order: order, pager: pager, store: make(map[string][]byte), keys: make([]string, 0)}
}

// Insert 插入键值。
func (t *BPlusTree) Insert(ctx context.Context, key, value []byte) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	sKey := string(key)
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.store[sKey]; !ok {
		t.keys = append(t.keys, sKey)
		sort.Strings(t.keys)
	}
	t.store[sKey] = append([]byte(nil), value...)
	return nil
}

// Delete 删除键。
func (t *BPlusTree) Delete(ctx context.Context, key []byte) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	sKey := string(key)
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.store[sKey]; !ok {
		return ErrKeyNotFound
	}
	delete(t.store, sKey)
	idx := sort.SearchStrings(t.keys, sKey)
	if idx < len(t.keys) && t.keys[idx] == sKey {
		t.keys = append(t.keys[:idx], t.keys[idx+1:]...)
	}
	return nil
}

// Search 点查。
func (t *BPlusTree) Search(ctx context.Context, key []byte) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	t.mu.RLock()
	defer t.mu.RUnlock()
	value, ok := t.store[string(key)]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return append([]byte(nil), value...), nil
}

// RangeScan 范围扫描。
func (t *BPlusTree) RangeScan(ctx context.Context, startKey, endKey []byte) (*Iterator, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	start := string(startKey)
	end := string(endKey)
	t.mu.RLock()
	defer t.mu.RUnlock()
	pairs := make([]KeyValue, 0)
	for _, key := range t.keys {
		if bytes.Compare([]byte(key), []byte(start)) < 0 {
			continue
		}
		if end != "" && bytes.Compare([]byte(key), []byte(end)) > 0 {
			break
		}
		pairs = append(pairs, KeyValue{Key: []byte(key), Value: append([]byte(nil), t.store[key]...)})
	}
	return NewIterator(pairs), nil
}

// Len 当前键数量。
func (t *BPlusTree) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.keys)
}
