package db

import "context"

// SAdd 向集合添加元素。
func (d *DB) SAdd(ctx context.Context, key, member string) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	entry, ok := d.data[key]
	if !ok || d.isExpired(entry) {
		set := map[string]struct{}{member: {}}
		d.data[key] = &Entry{Type: TypeSet, Value: set}
		return true, nil
	}
	if entry.Type != TypeSet {
		return false, ErrWrongType
	}
	set := entry.Value.(map[string]struct{})
	_, existed := set[member]
	set[member] = struct{}{}
	return !existed, nil
}

// SMembers 获取集合全部成员。
func (d *DB) SMembers(ctx context.Context, key string) ([]string, error) {
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
	if entry.Type != TypeSet {
		return nil, ErrWrongType
	}
	set := entry.Value.(map[string]struct{})
	result := make([]string, 0, len(set))
	for member := range set {
		result = append(result, member)
	}
	return result, nil
}
