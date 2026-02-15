package registry

import "testing"

func TestMemoryRegistry(t *testing.T) {
	r := NewMemoryRegistry()
	ins := &ServiceInstance{ID: "1", Name: "svc", Addr: "127.0.0.1:1", Weight: 1}
	if err := r.Register(ins); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	list, err := r.Discover("svc")
	if err != nil {
		t.Fatalf("discover failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("unexpected list size: %d", len(list))
	}
	if err := r.Deregister(ins); err != nil {
		t.Fatalf("deregister failed: %v", err)
	}
}
