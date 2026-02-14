package consumer

import (
	"context"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/storage"
)

// Client Consumer 客户端。
type Client struct {
	broker   *broker.Broker
	groupID  string
	memberID string
	offsets  map[string]int64
}

// NewClient 创建 Consumer。
func NewClient(b *broker.Broker, groupID, memberID string) *Client {
	return &Client{broker: b, groupID: groupID, memberID: memberID, offsets: make(map[string]int64)}
}

// Poll 拉取消息。
func (c *Client) Poll(ctx context.Context, topic string, partition int) (*storage.Message, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	key := topic + ":" + strconvItoa(partition)
	offset := c.offsets[key]
	msg, err := c.broker.Fetch(ctx, topic, partition, offset)
	if err != nil {
		return nil, err
	}
	c.offsets[key] = offset + 1
	return msg, nil
}

// Commit 提交 offset。
func (c *Client) Commit(topic string, partition int) {
	key := topic + ":" + strconvItoa(partition)
	offset := c.offsets[key]
	c.broker.OffsetMgr.Commit(c.groupID, topic, partition, offset)
}
