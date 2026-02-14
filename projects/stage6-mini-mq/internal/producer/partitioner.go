package producer

import "hash/crc32"

// Partitioner 分区策略接口。
type Partitioner interface {
	Partition(key []byte, numPartitions int) int
}

// RoundRobinPartitioner 轮询分区器。
type RoundRobinPartitioner struct {
	counter int64
}

// Partition 分配分区。
func (p *RoundRobinPartitioner) Partition(key []byte, numPartitions int) int {
	if numPartitions <= 0 {
		return 0
	}
	p.counter++
	return int(p.counter % int64(numPartitions))
}

// KeyHashPartitioner key 哈希分区器。
type KeyHashPartitioner struct{}

// Partition 分配分区。
func (p *KeyHashPartitioner) Partition(key []byte, numPartitions int) int {
	if numPartitions <= 0 {
		return 0
	}
	if len(key) == 0 {
		return 0
	}
	return int(crc32.ChecksumIEEE(key) % uint32(numPartitions))
}
