package infra

import (
	"errors"
	"sync"
	"time"
)

const (
	epoch     int64 = 1700000000000
	nodeBits        = 10
	seqBits         = 12
	maxNodeID int64 = -1 ^ (-1 << nodeBits)
	maxSeq    int64 = -1 ^ (-1 << seqBits)
)

// Snowflake 分布式 ID。
type Snowflake struct {
	mu       sync.Mutex
	nodeID   int64
	lastTime int64
	seq      int64
}

func NewSnowflake(nodeID int64) (*Snowflake, error) {
	if nodeID < 0 || nodeID > maxNodeID {
		return nil, errors.New("invalid node id")
	}
	return &Snowflake{nodeID: nodeID}, nil
}

func (s *Snowflake) NextID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UnixMilli()
	if now < s.lastTime {
		return 0, errors.New("clock moved backwards")
	}
	if now == s.lastTime {
		s.seq = (s.seq + 1) & maxSeq
		if s.seq == 0 {
			for now <= s.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.seq = 0
	}
	s.lastTime = now
	id := ((now - epoch) << (nodeBits + seqBits)) | (s.nodeID << seqBits) | s.seq
	return id, nil
}
