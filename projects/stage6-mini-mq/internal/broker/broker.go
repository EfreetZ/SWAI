package broker

import (
	"context"
	"errors"
	"sync"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/storage"
)

var ErrPartitionOutOfRange = errors.New("partition out of range")

// Broker MQ Broker 核心。
type Broker struct {
	ID         int
	Addr       string
	Topics     *TopicManager
	OffsetMgr  *OffsetManager
	GroupCoord *GroupCoordinator
	mu         sync.RWMutex
}

// NewBroker 创建 Broker。
func NewBroker(id int, addr, baseDir string) *Broker {
	return &Broker{
		ID:         id,
		Addr:       addr,
		Topics:     NewTopicManager(baseDir),
		OffsetMgr:  NewOffsetManager(),
		GroupCoord: NewGroupCoordinator(),
	}
}

// CreateTopic 创建 topic。
func (b *Broker) CreateTopic(name string, cfg TopicConfig) error {
	_, err := b.Topics.CreateTopic(name, cfg)
	return err
}

// Produce 生产消息。
func (b *Broker) Produce(ctx context.Context, topic string, partition int, key, value []byte) (int64, error) {
	t, err := b.Topics.GetTopic(topic)
	if err != nil {
		return 0, err
	}
	if partition < 0 || partition >= len(t.Partitions) {
		return 0, ErrPartitionOutOfRange
	}
	return t.Partitions[partition].Append(ctx, key, value)
}

// Fetch 消费消息。
func (b *Broker) Fetch(ctx context.Context, topic string, partition int, offset int64) (*storage.Message, error) {
	t, err := b.Topics.GetTopic(topic)
	if err != nil {
		return nil, err
	}
	if partition < 0 || partition >= len(t.Partitions) {
		return nil, ErrPartitionOutOfRange
	}
	return t.Partitions[partition].Read(ctx, offset)
}

// ListTopics 列表 topics。
func (b *Broker) ListTopics() []string {
	return b.Topics.ListTopics()
}
