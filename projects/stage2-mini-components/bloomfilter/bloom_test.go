package bloomfilter

import "testing"

func TestBloomFilterContains(t *testing.T) {
	filter := New(1000, 0.01)
	filter.Add([]byte("alice"))
	filter.Add([]byte("bob"))

	if !filter.Contains([]byte("alice")) {
		t.Fatal("Contains(alice) should be true")
	}
	if !filter.Contains([]byte("bob")) {
		t.Fatal("Contains(bob) should be true")
	}
}

func TestBloomFilterEstimateRate(t *testing.T) {
	filter := New(1000, 0.01)
	for i := 0; i < 200; i++ {
		filter.Add([]byte{byte(i), byte(i >> 1)})
	}

	rate := filter.EstimateFalsePositiveRate()
	if rate <= 0 || rate >= 0.1 {
		t.Fatalf("EstimateFalsePositiveRate() = %f, want (0, 0.1)", rate)
	}
}

func BenchmarkBloomFilterContains(b *testing.B) {
	filter := New(10000, 0.01)
	for i := 0; i < 10000; i++ {
		filter.Add([]byte{byte(i), byte(i >> 1), byte(i >> 2)})
	}
	data := []byte("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filter.Contains(data)
	}
}
