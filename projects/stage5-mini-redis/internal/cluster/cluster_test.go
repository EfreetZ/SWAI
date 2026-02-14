package cluster

import "testing"

func TestClusterResolve(t *testing.T) {
	state := NewState()
	state.AddNode(&Node{ID: "n1", Addr: "127.0.0.1:7001", Slots: []SlotRange{{Start: 0, End: 8191}}})
	state.AddNode(&Node{ID: "n2", Addr: "127.0.0.1:7002", Slots: []SlotRange{{Start: 8192, End: 16383}}})

	node, slot, err := state.ResolveNode("user:1001")
	if err != nil {
		t.Fatalf("ResolveNode() error = %v", err)
	}
	if node == nil {
		t.Fatalf("node is nil for slot %d", slot)
	}
}
