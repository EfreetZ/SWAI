package ds

import "testing"

func TestLinkedList(t *testing.T) {
	list := &LinkedList{}
	list.LPush("a")
	list.RPush("b")
	value, ok := list.LPop()
	if !ok || value != "a" {
		t.Fatalf("LPop() = (%q, %v), want (a, true)", value, ok)
	}
	items := list.Range(0, 10)
	if len(items) != 1 || items[0] != "b" {
		t.Fatalf("Range() = %v, want [b]", items)
	}
}

func TestSkipList(t *testing.T) {
	sl := NewSkipList()
	sl.Insert("a", 2)
	sl.Insert("b", 1)
	sl.Insert("c", 3)
	items := sl.RangeByScore(1, 2)
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Member != "b" {
		t.Fatalf("first member = %s, want b", items[0].Member)
	}
}
