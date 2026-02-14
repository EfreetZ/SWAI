package txn

import "github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/wal"

type State uint8

const (
	Active State = iota
	Committed
	Aborted
)

// WriteRecord 事务写集。
type WriteRecord struct {
	Key      []byte
	OldValue []byte
	NewValue []byte
	Type     wal.LogType
}

// Transaction 事务对象。
type Transaction struct {
	TxID      wal.TxID
	State     State
	WriteSets []WriteRecord
}
