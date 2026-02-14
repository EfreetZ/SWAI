package consumer

import "testing"

func TestGroupCoordinatorRebalance(t *testing.T) {
	coord := NewGroupCoordinator()
	coord.Join("g1", "m1")
	coord.Join("g1", "m2")

	partitions := BuildTopicPartitions("events", 4)
	assigned := coord.Rebalance("g1", partitions, &RoundRobinAssignor{})
	if len(assigned) != 2 {
		t.Fatalf("unexpected assigned members: %d", len(assigned))
	}
	if len(assigned["m1"])+len(assigned["m2"]) != 4 {
		t.Fatal("assignment count mismatch")
	}
}
