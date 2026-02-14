package wal

import (
	"context"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
)

// Recovery 简化恢复器。
type Recovery struct {
	wal  *Writer
	tree *storage.BPlusTree
}

// NewRecovery 创建恢复器。
func NewRecovery(walWriter *Writer, tree *storage.BPlusTree) *Recovery {
	return &Recovery{wal: walWriter, tree: tree}
}

// Replay 从指定 LSN 开始回放日志。
func (r *Recovery) Replay(ctx context.Context, from LSN) error {
	records, err := r.wal.ReadFrom(ctx, from)
	if err != nil {
		return err
	}
	for _, record := range records {
		switch record.Type {
		case LogInsert:
			if err = r.tree.Insert(ctx, record.OldValue, record.NewValue); err != nil {
				return err
			}
		case LogDelete:
			_ = r.tree.Delete(ctx, record.OldValue)
		}
	}
	return nil
}
