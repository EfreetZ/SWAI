package replication

import "testing"

func TestBacklog(t *testing.T) {
	b := NewBacklog(2)
	o1 := b.Append("SET a 1")
	o2 := b.Append("SET b 2")
	_ = o1
	_ = b.Append("SET c 3")
	entries := b.EntriesSince(o2 - 1)
	if len(entries) == 0 {
		t.Fatal("EntriesSince() should return entries")
	}
}

func TestMasterSlave(t *testing.T) {
	m := NewMaster()
	m.RegisterSlave("s1")
	offset := m.Broadcast("SET a 1")
	s := NewSlave("127.0.0.1", 16379)
	s.Ack(offset)
	if s.Offset() != offset {
		t.Fatalf("Offset() = %d, want %d", s.Offset(), offset)
	}
}
