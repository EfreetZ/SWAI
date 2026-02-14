package producer

import (
	"context"
	"errors"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
)

var ErrTopicNotFound = errors.New("topic not found")

// Client Producer 客户端。
type Client struct {
	broker      *broker.Broker
	batchBuffer *BatchBuffer
	partitioner Partitioner
}

// NewClient 创建 Producer。
func NewClient(b *broker.Broker, batchSize int, partitioner Partitioner) *Client {
	if partitioner == nil {
		partitioner = &RoundRobinPartitioner{}
	}
	return &Client{broker: b, batchBuffer: NewBatchBuffer(batchSize, 10), partitioner: partitioner}
}

// Send 发送消息。
func (p *Client) Send(ctx context.Context, topic string, key, value []byte) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	t, err := p.broker.Topics.GetTopic(topic)
	if err != nil {
		return ErrTopicNotFound
	}
	partition := p.partitioner.Partition(key, len(t.Partitions))
	_, err = p.broker.Produce(ctx, topic, partition, key, value)
	if err != nil {
		return err
	}
	flush := p.batchBuffer.Add(&Message{Topic: topic, Key: key, Value: value})
	if flush {
		p.batchBuffer.Drain()
	}
	return nil
}
