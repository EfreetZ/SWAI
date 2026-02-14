package consumer

// BuildTopicPartitions 构建 topic 分区列表。
func BuildTopicPartitions(topic string, numPartitions int) []TopicPartition {
	res := make([]TopicPartition, 0, numPartitions)
	for i := 0; i < numPartitions; i++ {
		res = append(res, TopicPartition{Topic: topic, Partition: i})
	}
	return res
}
