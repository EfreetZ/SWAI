package broker

import (
	"errors"
	"path/filepath"
	"sync"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/storage"
)

var ErrTopicNotFound = errors.New("topic not found")

// TopicConfig topic 配置。
type TopicConfig struct {
	NumPartitions int
	SegmentBytes  int64
}

// Topic topic 模型。
type Topic struct {
	Name       string
	Partitions []*storage.Partition
	Config     TopicConfig
}

// TopicManager topic 管理。
type TopicManager struct {
	mu      sync.RWMutex
	baseDir string
	topics  map[string]*Topic
}

// NewTopicManager 创建管理器。
func NewTopicManager(baseDir string) *TopicManager {
	return &TopicManager{baseDir: baseDir, topics: make(map[string]*Topic)}
}

// CreateTopic 创建 topic。
func (m *TopicManager) CreateTopic(name string, cfg TopicConfig) (*Topic, error) {
	if cfg.NumPartitions <= 0 {
		cfg.NumPartitions = 1
	}
	if cfg.SegmentBytes <= 0 {
		cfg.SegmentBytes = 1024 * 1024
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.topics[name]; ok {
		return existing, nil
	}

	partitions := make([]*storage.Partition, 0, cfg.NumPartitions)
	for i := 0; i < cfg.NumPartitions; i++ {
		partition, err := storage.NewPartition(name, i, filepath.Join(m.baseDir, "data"), cfg.SegmentBytes)
		if err != nil {
			return nil, err
		}
		partitions = append(partitions, partition)
	}
	topic := &Topic{Name: name, Partitions: partitions, Config: cfg}
	m.topics[name] = topic
	return topic, nil
}

// GetTopic 获取 topic。
func (m *TopicManager) GetTopic(name string) (*Topic, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	topic, ok := m.topics[name]
	if !ok {
		return nil, ErrTopicNotFound
	}
	return topic, nil
}

// ListTopics 列出 topic。
func (m *TopicManager) ListTopics() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]string, 0, len(m.topics))
	for name := range m.topics {
		res = append(res, name)
	}
	return res
}
