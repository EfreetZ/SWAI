package storage

import "testing"

func TestSparseIndexLookup(t *testing.T) {
	idx := NewSparseIndex()
	idx.Append(0, 0)
	idx.Append(10, 100)
	idx.Append(20, 200)

	pos, err := idx.Lookup(15)
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}
	if pos != 100 {
		t.Fatalf("unexpected position: %d", pos)
	}
}
