package utils

import "testing"

func TestUint32BytesRoundTrip(t *testing.T) {
	value := uint32(123456)
	encoded := Uint32ToBytes(value)
	decoded := BytesToUint32(encoded)
	if decoded != value {
		t.Fatalf("decoded = %d, want %d", decoded, value)
	}
}
