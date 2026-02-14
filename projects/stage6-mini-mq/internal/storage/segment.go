package storage

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ErrOffsetNotFound = errors.New("offset not found")

// Message 消息结构。
type Message struct {
	Offset    int64
	Key       []byte
	Value     []byte
	Timestamp int64
}

// Segment 日志段。
type Segment struct {
	BaseOffset int64
	NextOffset int64
	LogFile    *os.File
	LogSize    int64
	MaxBytes   int64
	mu         sync.Mutex
}

// NewSegment 创建 segment。
func NewSegment(dir string, baseOffset, maxBytes int64) (*Segment, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	logPath := filepath.Join(dir, formatSegmentName(baseOffset)+".log")
	file, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return &Segment{BaseOffset: baseOffset, NextOffset: baseOffset, LogFile: file, LogSize: stat.Size(), MaxBytes: maxBytes}, nil
}

// Append 追加消息。
func (s *Segment) Append(ctx context.Context, msg *Message) (int64, int64, error) {
	if msg == nil {
		return 0, 0, errors.New("message is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return 0, 0, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	msg.Offset = s.NextOffset
	msg.Timestamp = time.Now().UnixMilli()
	encoded, err := encodeMessage(msg)
	if err != nil {
		return 0, 0, err
	}
	position := s.LogSize
	if _, err = s.LogFile.Write(encoded); err != nil {
		return 0, 0, err
	}
	s.LogSize += int64(len(encoded))
	s.NextOffset++
	return msg.Offset, position, nil
}

// Read 按偏移读取消息（线性扫描）。
func (s *Segment) Read(ctx context.Context, offset int64) (*Message, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.LogFile.Seek(0, 0); err != nil {
		return nil, err
	}
	reader := bufio.NewReader(s.LogFile)
	for {
		msg, err := decodeMessage(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, ErrOffsetNotFound
			}
			return nil, err
		}
		if msg.Offset == offset {
			return msg, nil
		}
	}
}

// IsFull 是否达到滚动阈值。
func (s *Segment) IsFull() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.LogSize >= s.MaxBytes
}

// Close 关闭段文件。
func (s *Segment) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.LogFile.Close()
}

func formatSegmentName(baseOffset int64) string {
	name := make([]byte, 20)
	for i := len(name) - 1; i >= 0; i-- {
		name[i] = byte('0' + (baseOffset % 10))
		baseOffset /= 10
	}
	return string(name)
}

func encodeMessage(msg *Message) ([]byte, error) {
	keyLen := uint32(len(msg.Key))
	valueLen := uint32(len(msg.Value))
	totalLen := 8 + 8 + 4 + keyLen + 4 + valueLen
	buf := make([]byte, 4+4+totalLen)
	binary.BigEndian.PutUint32(buf[0:4], totalLen)
	binary.BigEndian.PutUint64(buf[8:16], uint64(msg.Offset))
	binary.BigEndian.PutUint64(buf[16:24], uint64(msg.Timestamp))
	binary.BigEndian.PutUint32(buf[24:28], keyLen)
	copy(buf[28:28+keyLen], msg.Key)
	valueLenPos := 28 + keyLen
	binary.BigEndian.PutUint32(buf[valueLenPos:valueLenPos+4], valueLen)
	copy(buf[valueLenPos+4:], msg.Value)
	crc := crc32.ChecksumIEEE(buf[8:])
	binary.BigEndian.PutUint32(buf[4:8], crc)
	return buf, nil
}

func decodeMessage(reader *bufio.Reader) (*Message, error) {
	head := make([]byte, 8)
	if _, err := io.ReadFull(reader, head); err != nil {
		return nil, err
	}
	totalLen := binary.BigEndian.Uint32(head[0:4])
	crc := binary.BigEndian.Uint32(head[4:8])
	payload := make([]byte, totalLen)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	if crc32.ChecksumIEEE(payload) != crc {
		return nil, errors.New("crc mismatch")
	}

	msg := &Message{}
	msg.Offset = int64(binary.BigEndian.Uint64(payload[0:8]))
	msg.Timestamp = int64(binary.BigEndian.Uint64(payload[8:16]))
	keyLen := binary.BigEndian.Uint32(payload[16:20])
	msg.Key = append([]byte(nil), payload[20:20+keyLen]...)
	valueLenPos := 20 + keyLen
	valueLen := binary.BigEndian.Uint32(payload[valueLenPos : valueLenPos+4])
	msg.Value = append([]byte(nil), payload[valueLenPos+4:valueLenPos+4+valueLen]...)
	return msg, nil
}
