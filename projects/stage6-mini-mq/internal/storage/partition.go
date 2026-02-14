package storage

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
)

// Partition 分区。
type Partition struct {
	TopicName string
	ID        int
	Dir       string

	ActiveSeg *Segment
	Segments  []*Segment
	Index     *SparseIndex

	segmentMaxBytes int64
	mu              sync.RWMutex
}

// NewPartition 创建分区。
func NewPartition(topic string, id int, dir string, segmentMaxBytes int64) (*Partition, error) {
	seg, err := NewSegment(filepath.Join(dir, topic, partitionDirName(id)), 0, segmentMaxBytes)
	if err != nil {
		return nil, err
	}
	return &Partition{TopicName: topic, ID: id, Dir: dir, ActiveSeg: seg, Segments: []*Segment{seg}, Index: NewSparseIndex(), segmentMaxBytes: segmentMaxBytes}, nil
}

// Append 追加消息。
func (p *Partition) Append(ctx context.Context, key, value []byte) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ActiveSeg.IsFull() {
		if err := p.rollSegment(); err != nil {
			return 0, err
		}
	}
	offset, position, err := p.ActiveSeg.Append(ctx, &Message{Key: key, Value: value})
	if err != nil {
		return 0, err
	}
	if offset%10 == 0 {
		p.Index.Append(offset, position)
	}
	return offset, nil
}

// Read 读取指定 offset。
func (p *Partition) Read(ctx context.Context, offset int64) (*Message, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	p.mu.RLock()
	segments := append([]*Segment(nil), p.Segments...)
	p.mu.RUnlock()
	for i := len(segments) - 1; i >= 0; i-- {
		msg, err := segments[i].Read(ctx, offset)
		if err == nil {
			return msg, nil
		}
		if errors.Is(err, ErrOffsetNotFound) {
			continue
		}
		return nil, err
	}
	return nil, ErrOffsetNotFound
}

// Close 关闭分区资源。
func (p *Partition) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, seg := range p.Segments {
		_ = seg.Close()
	}
	return nil
}

func (p *Partition) rollSegment() error {
	baseOffset := p.ActiveSeg.NextOffset
	newSeg, err := NewSegment(filepath.Join(p.Dir, p.TopicName, partitionDirName(p.ID)), baseOffset, p.segmentMaxBytes)
	if err != nil {
		return err
	}
	p.ActiveSeg = newSeg
	p.Segments = append(p.Segments, newSeg)
	return nil
}

func partitionDirName(id int) string {
	return "partition-" + strconvItoa(id)
}

func strconvItoa(v int) string {
	if v == 0 {
		return "0"
	}
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	buf := make([]byte, 0, 12)
	for v > 0 {
		buf = append([]byte{byte('0' + v%10)}, buf...)
		v /= 10
	}
	return sign + string(buf)
}
