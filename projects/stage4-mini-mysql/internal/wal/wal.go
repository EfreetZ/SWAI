package wal

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
)

// Writer WAL 核心实现。
type Writer struct {
	mu      sync.Mutex
	file    *os.File
	writer  *bufio.Writer
	nextLSN LSN
	records []*LogRecord
}

// NewWriter 创建 WAL。
func NewWriter(path string) (*Writer, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	writer := &Writer{file: file, writer: bufio.NewWriter(file), nextLSN: 1, records: make([]*LogRecord, 0)}
	if err = writer.loadExisting(); err != nil {
		_ = file.Close()
		return nil, err
	}
	return writer, nil
}

// Close 关闭 WAL。
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.writer.Flush(); err != nil {
		return err
	}
	return w.file.Close()
}

// Append 追加日志。
func (w *Writer) Append(ctx context.Context, record *LogRecord) (LSN, error) {
	if record == nil {
		return 0, errors.New("record is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	recordCopy := *record
	recordCopy.LSN = w.nextLSN
	w.nextLSN++
	payload, err := json.Marshal(&recordCopy)
	if err != nil {
		return 0, err
	}
	if _, err = w.writer.Write(payload); err != nil {
		return 0, err
	}
	if err = w.writer.WriteByte('\n'); err != nil {
		return 0, err
	}
	w.records = append(w.records, &recordCopy)
	return recordCopy.LSN, nil
}

// Flush 刷盘。
func (w *Writer) Flush(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.writer.Flush(); err != nil {
		return err
	}
	return w.file.Sync()
}

// ReadFrom 从指定 LSN 读取日志。
func (w *Writer) ReadFrom(ctx context.Context, lsn LSN) ([]*LogRecord, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	result := make([]*LogRecord, 0)
	for _, record := range w.records {
		if record.LSN >= lsn {
			copyRecord := *record
			result = append(result, &copyRecord)
		}
	}
	return result, nil
}

// Checkpoint 生成检查点（简化：即刷盘）。
func (w *Writer) Checkpoint(ctx context.Context) error {
	return w.Flush(ctx)
}

func (w *Writer) loadExisting() error {
	if _, err := w.file.Seek(0, 0); err != nil {
		return err
	}
	scanner := bufio.NewScanner(w.file)
	for scanner.Scan() {
		record := &LogRecord{}
		if err := json.Unmarshal(scanner.Bytes(), record); err != nil {
			continue
		}
		w.records = append(w.records, record)
		if record.LSN >= w.nextLSN {
			w.nextLSN = record.LSN + 1
		}
	}
	if _, err := w.file.Seek(0, 2); err != nil {
		return err
	}
	return scanner.Err()
}
