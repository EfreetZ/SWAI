package consumer

import "sort"

// TopicPartition topic-partition 标识。
type TopicPartition struct {
	Topic     string
	Partition int
}

// AssignmentStrategy 分配策略。
type AssignmentStrategy interface {
	Assign(members []string, partitions []TopicPartition) map[string][]TopicPartition
}

// RangeAssignor 区间分配。
type RangeAssignor struct{}

// Assign 按顺序分片。
func (a *RangeAssignor) Assign(members []string, partitions []TopicPartition) map[string][]TopicPartition {
	res := make(map[string][]TopicPartition)
	if len(members) == 0 {
		return res
	}
	sort.Strings(members)
	for i, tp := range partitions {
		member := members[i%len(members)]
		res[member] = append(res[member], tp)
	}
	return res
}

// RoundRobinAssignor 轮询分配。
type RoundRobinAssignor struct{}

// Assign 轮询分配。
func (a *RoundRobinAssignor) Assign(members []string, partitions []TopicPartition) map[string][]TopicPartition {
	res := make(map[string][]TopicPartition)
	if len(members) == 0 {
		return res
	}
	sort.Strings(members)
	for i, tp := range partitions {
		member := members[i%len(members)]
		res[member] = append(res[member], tp)
	}
	return res
}
