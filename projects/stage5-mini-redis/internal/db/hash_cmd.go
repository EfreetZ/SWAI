package db

import "context"

// HSet 设置 hash 字段值。
func (d *DB) HSet(ctx context.Context, key, field, value string) error {
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
		h := map[string]string{field: value}
		d.data[key] = &Entry{Type: TypeHash, Value: h}
		return nil
	}
	if entry.Type != TypeHash {
		return ErrWrongType
	}
	h := entry.Value.(map[string]string)
	h[field] = value
	return nil
}

// HGet 获取 hash 字段值。
func (d *DB) HGet(ctx context.Context, key, field string) (string, bool, error) {
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
	if entry.Type != TypeHash {
		return "", false, ErrWrongType
	}
	h := entry.Value.(map[string]string)
	value, exists := h[field]
	return value, exists, nil
}

// HGetAll 获取 hash 全量字段。
func (d *DB) HGetAll(ctx context.Context, key string) (map[string]string, error) {
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
		return map[string]string{}, nil
	}
	if entry.Type != TypeHash {
		return nil, ErrWrongType
	}
	h := entry.Value.(map[string]string)
	copyMap := make(map[string]string, len(h))
	for field, value := range h {
		copyMap[field] = value
	}
	return copyMap, nil
}
