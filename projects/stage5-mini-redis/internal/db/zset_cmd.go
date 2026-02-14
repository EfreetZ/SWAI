package db

import (
	"context"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/ds"
)

// ZAdd 写入有序集合。
func (d *DB) ZAdd(ctx context.Context, key, member string, score float64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	entry, ok := d.data[key]
	if !ok || d.isExpired(entry) {
		z := ds.NewSkipList()
		z.Insert(member, score)
		d.data[key] = &Entry{Type: TypeZSet, Value: z}
		return nil
	}
	if entry.Type != TypeZSet {
		return ErrWrongType
	}
	z := entry.Value.(*ds.SkipList)
	z.Insert(member, score)
	return nil
}

// ZRangeByScore 范围查询有序集合。
func (d *DB) ZRangeByScore(ctx context.Context, key string, min, max float64) ([]ds.ZItem, error) {
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
	if entry.Type != TypeZSet {
		return nil, ErrWrongType
	}
	z := entry.Value.(*ds.SkipList)
	return z.RangeByScore(min, max), nil
}
