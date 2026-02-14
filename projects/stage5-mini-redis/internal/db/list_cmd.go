package db

import (
	"context"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/ds"
)

// LPush 头插列表。
func (d *DB) LPush(ctx context.Context, key, value string) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	entry, ok := d.data[key]
	if !ok || d.isExpired(entry) {
		list := &ds.LinkedList{}
		list.LPush(value)
		d.data[key] = &Entry{Type: TypeList, Value: list}
		return 1, nil
	}
	if entry.Type != TypeList {
		return 0, ErrWrongType
	}
	list := entry.Value.(*ds.LinkedList)
	list.LPush(value)
	return len(list.Range(0, 1<<20)), nil
}

// RPush 尾插列表。
func (d *DB) RPush(ctx context.Context, key, value string) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	entry, ok := d.data[key]
	if !ok || d.isExpired(entry) {
		list := &ds.LinkedList{}
		list.RPush(value)
		d.data[key] = &Entry{Type: TypeList, Value: list}
		return 1, nil
	}
	if entry.Type != TypeList {
		return 0, ErrWrongType
	}
	list := entry.Value.(*ds.LinkedList)
	list.RPush(value)
	return len(list.Range(0, 1<<20)), nil
}

// LPop 头删列表。
func (d *DB) LPop(ctx context.Context, key string) (string, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", false, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	entry, ok := d.data[key]
	if !ok || d.isExpired(entry) {
		delete(d.data, key)
		return "", false, nil
	}
	if entry.Type != TypeList {
		return "", false, ErrWrongType
	}
	list := entry.Value.(*ds.LinkedList)
	value, exists := list.LPop()
	return value, exists, nil
}

// LRange 返回区间列表。
func (d *DB) LRange(ctx context.Context, key string, start, stop int) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	entry, ok := d.data[key]
	if !ok || d.isExpired(entry) {
		delete(d.data, key)
		return nil, nil
	}
	if entry.Type != TypeList {
		return nil, ErrWrongType
	}
	list := entry.Value.(*ds.LinkedList)
	return list.Range(start, stop), nil
}
