package wal

import (
	"context"
	"path/filepath"
	"testing"
)

func TestWALAppendReadFlush(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.wal")
	writer, err := NewWriter(path)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	defer func() {
		_ = writer.Close()
	}()

	lsn, err := writer.Append(context.Background(), &LogRecord{TxID: 1, Type: LogInsert, OldValue: []byte("k"), NewValue: []byte("v")})
	if err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if lsn == 0 {
		t.Fatal("LSN should not be zero")
	}
	if err = writer.Flush(context.Background()); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	records, err := writer.ReadFrom(context.Background(), 1)
	if err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}
}

func BenchmarkWALAppend(b *testing.B) {
	path := filepath.Join(b.TempDir(), "bench.wal")
	writer, err := NewWriter(path)
	if err != nil {
		b.Fatalf("NewWriter() error = %v", err)
	}
	defer func() {
		_ = writer.Close()
	}()

	for i := 0; i < b.N; i++ {
		_, appendErr := writer.Append(context.Background(), &LogRecord{TxID: 1, Type: LogInsert, OldValue: []byte("k"), NewValue: []byte("v")})
		if appendErr != nil {
			b.Fatalf("Append() error = %v", appendErr)
		}
	}
}
