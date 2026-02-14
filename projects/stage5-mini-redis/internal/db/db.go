package db

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/ds"
)

var ErrWrongType = errors.New("wrong type operation")

type ValueType uint8

const (
	TypeString ValueType = iota
	TypeList
	TypeSet
	TypeZSet
	TypeHash
)

// Entry 键值条目。
type Entry struct {
	Type     ValueType
	Value    any
	ExpireAt time.Time
}

// DB 核心内存数据库。
type DB struct {
	mu   sync.RWMutex
	data map[string]*Entry
}

// New 创建 DB。
func New() *DB {
	return &DB{data: make(map[string]*Entry)}
}

// SetString 设置字符串。
func (d *DB) SetString(ctx context.Context, key, value string, ttl time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	entry := &Entry{Type: TypeString, Value: value}
	if ttl > 0 {
		entry.ExpireAt = time.Now().Add(ttl)
	}
	d.data[key] = entry
	return nil
}

// GetString 获取字符串。
func (d *DB) GetString(ctx context.Context, key string) (string, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", false, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	entry, ok := d.data[key]
	if !ok {
		return "", false, nil
	}
	if d.isExpired(entry) {
		delete(d.data, key)
		return "", false, nil
	}
	if entry.Type != TypeString {
		return "", false, ErrWrongType
	}
	value, _ := entry.Value.(string)
	return value, true, nil
}

// Del 删除 key。
func (d *DB) Del(ctx context.Context, key string) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.data[key]; !ok {
		return false, nil
	}
	delete(d.data, key)
	return true, nil
}

// Exists 判断 key 是否存在。
func (d *DB) Exists(ctx context.Context, key string) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	entry, ok := d.data[key]
	if !ok {
		return false, nil
	}
	if d.isExpired(entry) {
		delete(d.data, key)
		return false, nil
	}
	return true, nil
}

// Keys 返回匹配 pattern 的 key（支持 * 前后缀模糊）。
func (d *DB) Keys(ctx context.Context, pattern string) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	result := make([]string, 0)
	for key, entry := range d.data {
		if d.isExpired(entry) {
			delete(d.data, key)
			continue
		}
		if matchPattern(key, pattern) {
			result = append(result, key)
		}
	}
	sort.Strings(result)
	return result, nil
}

// Expire 设置过期时间。
func (d *DB) Expire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	entry, ok := d.data[key]
	if !ok {
		return false, nil
	}
	entry.ExpireAt = time.Now().Add(ttl)
	return true, nil
}

// TTL 返回剩余生存时间秒数。
func (d *DB) TTL(ctx context.Context, key string) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return -1, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	entry, ok := d.data[key]
	if !ok {
		return -2, nil
	}
	if entry.ExpireAt.IsZero() {
		return -1, nil
	}
	if d.isExpired(entry) {
		delete(d.data, key)
		return -2, nil
	}
	return int64(time.Until(entry.ExpireAt).Seconds()), nil
}

// Snapshot 返回全量快照副本。
func (d *DB) Snapshot() map[string]*Entry {
	d.mu.RLock()
	defer d.mu.RUnlock()
	copyData := make(map[string]*Entry, len(d.data))
	for key, entry := range d.data {
		entryCopy := *entry
		switch entry.Type {
		case TypeSet:
			s := entry.Value.(map[string]struct{})
			sCopy := make(map[string]struct{}, len(s))
			for m := range s {
				sCopy[m] = struct{}{}
			}
			entryCopy.Value = sCopy
		case TypeHash:
			h := entry.Value.(map[string]string)
			hCopy := make(map[string]string, len(h))
			for f, v := range h {
				hCopy[f] = v
			}
			entryCopy.Value = hCopy
		case TypeList:
			list := entry.Value.(*ds.LinkedList)
			lCopy := &ds.LinkedList{}
			for _, item := range list.Range(0, 1<<20) {
				lCopy.RPush(item)
			}
			entryCopy.Value = lCopy
		case TypeZSet:
			z := entry.Value.(*ds.SkipList)
			zCopy := ds.NewSkipList()
			for _, item := range z.RangeByScore(-1<<20, 1<<20) {
				zCopy.Insert(item.Member, item.Score)
			}
			entryCopy.Value = zCopy
		}
		copyData[key] = &entryCopy
	}
	return copyData
}

// LoadSnapshot 加载快照。
func (d *DB) LoadSnapshot(snapshot map[string]*Entry) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.data = snapshot
}

// ActiveExpire 定期清理过期键。
func (d *DB) ActiveExpire(ctx context.Context, sampleLimit int) {
	if sampleLimit <= 0 {
		sampleLimit = 20
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	count := 0
	for key, entry := range d.data {
		if d.isExpired(entry) {
			delete(d.data, key)
		}
		count++
		if count >= sampleLimit {
			break
		}
	}
}

func (d *DB) isExpired(entry *Entry) bool {
	return entry != nil && !entry.ExpireAt.IsZero() && time.Now().After(entry.ExpireAt)
}

func matchPattern(key, pattern string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		inner := strings.Trim(pattern, "*")
		return strings.Contains(key, inner)
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(key, strings.TrimPrefix(pattern, "*"))
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(key, strings.TrimSuffix(pattern, "*"))
	}
	return key == pattern
}
