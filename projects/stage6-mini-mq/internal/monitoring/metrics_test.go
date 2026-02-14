package monitoring

import "testing"

func TestBrokerMetrics(t *testing.T) {
	m := BrokerMetrics{ProducedTotal: 100, ConsumedTotal: 90, LagTotal: 10}
	if m.Throughput() != 190 {
		t.Fatalf("unexpected throughput: %d", m.Throughput())
	}
	if !m.LagHealthy(20) {
		t.Fatal("lag should be healthy")
	}
	if m.LagHealthy(5) {
		t.Fatal("lag should be unhealthy")
	}
}
