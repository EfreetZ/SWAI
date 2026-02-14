package producer

import (
	"sync"
	"time"
)

// Message 待发送消息。
type Message struct {
	Topic string
	Key   []byte
	Value []byte
}

// BatchBuffer 批量缓冲。
type BatchBuffer struct {
	mu        sync.Mutex
	batchSize int
	linger    time.Duration
	messages  []*Message
	createdAt time.Time
}

// NewBatchBuffer 创建批量缓冲。
func NewBatchBuffer(batchSize int, linger time.Duration) *BatchBuffer {
	if batchSize <= 0 {
		batchSize = 100
	}
	if linger <= 0 {
		linger = 10 * time.Millisecond
	}
	return &BatchBuffer{batchSize: batchSize, linger: linger, messages: make([]*Message, 0, batchSize), createdAt: time.Now()}
}

// Add 添加消息并判断是否应 flush。
func (b *BatchBuffer) Add(msg *Message) (flush bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.messages) == 0 {
		b.createdAt = time.Now()
	}
	b.messages = append(b.messages, msg)
	if len(b.messages) >= b.batchSize {
		return true
	}
	if time.Since(b.createdAt) >= b.linger {
		return true
	}
	return false
}

// Drain 清空并返回当前批次。
func (b *BatchBuffer) Drain() []*Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.messages) == 0 {
		return nil
	}
	batch := make([]*Message, len(b.messages))
	copy(batch, b.messages)
	b.messages = b.messages[:0]
	b.createdAt = time.Now()
	return batch
}
