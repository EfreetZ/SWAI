package persist

import (
	"context"
	"encoding/json"
	"os"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/db"
)

// SaveSnapshot 保存快照。
func SaveSnapshot(ctx context.Context, path string, snapshot map[string]*db.Entry) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

// LoadSnapshot 加载快照。
func LoadSnapshot(ctx context.Context, path string) (map[string]*db.Entry, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	snapshot := make(map[string]*db.Entry)
	if err = json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}
