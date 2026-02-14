package wal

import "github.com/EfreetZ/SWAI/projects/stage4-mini-mysql/internal/storage"

type LSN uint64
type TxID uint64

type LogType uint8

const (
	LogBegin LogType = iota
	LogInsert
	LogDelete
	LogCommit
	LogAbort
)

// LogRecord WAL 日志记录。
type LogRecord struct {
	LSN      LSN            `json:"lsn"`
	TxID     TxID           `json:"tx_id"`
	Type     LogType        `json:"type"`
	PageID   storage.PageID `json:"page_id"`
	Offset   uint16         `json:"offset"`
	OldValue []byte         `json:"old_value,omitempty"`
	NewValue []byte         `json:"new_value,omitempty"`
}
