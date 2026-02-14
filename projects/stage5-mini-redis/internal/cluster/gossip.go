package cluster

import "time"

// GossipType gossip 消息类型。
type GossipType string

const (
	GossipPing GossipType = "PING"
	GossipPong GossipType = "PONG"
)

// GossipMessage gossip 消息。
type GossipMessage struct {
	SenderID string
	Type     GossipType
	Epoch    uint64
	SentAt   time.Time
}
