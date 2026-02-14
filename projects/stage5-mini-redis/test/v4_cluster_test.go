package test

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage5-mini-redis/internal/cluster"
)

func TestV4ClusterSlotRouting(t *testing.T) {
	state := cluster.NewState()
	state.AddNode(&cluster.Node{ID: "n1", Addr: "127.0.0.1:7001", Slots: []cluster.SlotRange{{Start: 0, End: 5460}}})
	state.AddNode(&cluster.Node{ID: "n2", Addr: "127.0.0.1:7002", Slots: []cluster.SlotRange{{Start: 5461, End: 10922}}})
	state.AddNode(&cluster.Node{ID: "n3", Addr: "127.0.0.1:7003", Slots: []cluster.SlotRange{{Start: 10923, End: 16383}}})

	node, slot, err := state.ResolveNode("order:1001")
	if err != nil {
		t.Fatalf("ResolveNode() error = %v", err)
	}
	if node == nil || node.Addr == "" {
		t.Fatalf("resolved node invalid for slot %d", slot)
	}
}
