package balancer

import (
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
)

func TestRoundRobinAndRandom(t *testing.T) {
	instances := []*registry.ServiceInstance{{ID: "a", Weight: 1}, {ID: "b", Weight: 1}, {ID: "c", Weight: 1}}
	rr := &RoundRobinBalancer{}
	first, err := rr.Pick(instances)
	if err != nil {
		t.Fatalf("rr pick failed: %v", err)
	}
	if first == nil {
		t.Fatal("nil instance")
	}

	rnd := NewRandomBalancer()
	picked, err := rnd.Pick(instances)
	if err != nil {
		t.Fatalf("random pick failed: %v", err)
	}
	if picked == nil {
		t.Fatal("nil picked")
	}
}
