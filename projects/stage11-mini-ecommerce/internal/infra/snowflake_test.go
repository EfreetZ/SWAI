package infra

import "testing"

func TestSnowflake(t *testing.T) {
	s, err := NewSnowflake(1)
	if err != nil {
		t.Fatalf("new snowflake failed: %v", err)
	}
	id1, _ := s.NextID()
	id2, _ := s.NextID()
	if id2 <= id1 {
		t.Fatal("id should increase")
	}
}
