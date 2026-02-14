package txn

import (
	"context"
	"errors"
	"sync"

	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"
	"github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/wal"
)

var ErrTxNotActive = errors.New("transaction is not active")

// Manager 事务管理器。
type Manager struct {
	mu       sync.Mutex
	nextTxID wal.TxID
	wal      *wal.Writer
	tree     *storage.BPlusTree
	lockMgr  *LockManager
}

// NewManager 创建事务管理器。
func NewManager(walWriter *wal.Writer, tree *storage.BPlusTree) *Manager {
	return &Manager{nextTxID: 1, wal: walWriter, tree: tree, lockMgr: NewLockManager()}
}

// Begin 开启事务。
func (m *Manager) Begin(ctx context.Context) (*Transaction, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	txID := m.nextTxID
	m.nextTxID++
	m.mu.Unlock()

	tx := &Transaction{TxID: txID, State: Active, WriteSets: make([]WriteRecord, 0)}
	_, err := m.wal.Append(ctx, &wal.LogRecord{TxID: txID, Type: wal.LogBegin})
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Put 在事务内写入 key。
func (m *Manager) Put(ctx context.Context, tx *Transaction, key, value []byte) error {
	if tx == nil || tx.State != Active {
		return ErrTxNotActive
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := m.lockMgr.LockKey(ctx, string(key)); err != nil {
		return err
	}
	defer m.lockMgr.UnlockKey(string(key))

	oldValue, _ := m.tree.Search(ctx, key)
	tx.WriteSets = append(tx.WriteSets, WriteRecord{Key: append([]byte(nil), key...), OldValue: oldValue, NewValue: append([]byte(nil), value...), Type: wal.LogInsert})
	return nil
}

// Delete 在事务内删除 key。
func (m *Manager) Delete(ctx context.Context, tx *Transaction, key []byte) error {
	if tx == nil || tx.State != Active {
		return ErrTxNotActive
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := m.lockMgr.LockKey(ctx, string(key)); err != nil {
		return err
	}
	defer m.lockMgr.UnlockKey(string(key))

	oldValue, _ := m.tree.Search(ctx, key)
	tx.WriteSets = append(tx.WriteSets, WriteRecord{Key: append([]byte(nil), key...), OldValue: oldValue, Type: wal.LogDelete})
	return nil
}

// Commit 提交事务。
func (m *Manager) Commit(ctx context.Context, tx *Transaction) error {
	if tx == nil || tx.State != Active {
		return ErrTxNotActive
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	for _, writeSet := range tx.WriteSets {
		if err := m.lockMgr.LockKey(ctx, string(writeSet.Key)); err != nil {
			return err
		}
		if writeSet.Type == wal.LogInsert {
			if err := m.tree.Insert(ctx, writeSet.Key, writeSet.NewValue); err != nil {
				m.lockMgr.UnlockKey(string(writeSet.Key))
				return err
			}
		} else if writeSet.Type == wal.LogDelete {
			_ = m.tree.Delete(ctx, writeSet.Key)
		}
		_, err := m.wal.Append(ctx, &wal.LogRecord{TxID: tx.TxID, Type: writeSet.Type, OldValue: writeSet.Key, NewValue: writeSet.NewValue})
		m.lockMgr.UnlockKey(string(writeSet.Key))
		if err != nil {
			return err
		}
	}
	if _, err := m.wal.Append(ctx, &wal.LogRecord{TxID: tx.TxID, Type: wal.LogCommit}); err != nil {
		return err
	}
	if err := m.wal.Flush(ctx); err != nil {
		return err
	}
	tx.State = Committed
	return nil
}

// Abort 回滚事务。
func (m *Manager) Abort(ctx context.Context, tx *Transaction) error {
	if tx == nil || tx.State != Active {
		return ErrTxNotActive
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := m.wal.Append(ctx, &wal.LogRecord{TxID: tx.TxID, Type: wal.LogAbort}); err != nil {
		return err
	}
	tx.State = Aborted
	tx.WriteSets = nil
	return nil
}
