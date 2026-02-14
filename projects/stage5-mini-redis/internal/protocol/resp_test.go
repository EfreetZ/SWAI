package protocol

import (
	"bufio"
	"bytes"
	"testing"
)

func TestRESPParseSerialize(t *testing.T) {
	input := []byte("*2\r\n$4\r\nPING\r\n$4\r\nPONG\r\n")
	value, err := Parse(bufio.NewReader(bytes.NewReader(input)))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if value.Type != Array || len(value.Array) != 2 {
		t.Fatalf("parsed value unexpected: %+v", value)
	}

	encoded := Serialize(&Value{Type: SimpleString, Str: "OK"})
	if string(encoded) != "+OK\r\n" {
		t.Fatalf("Serialize() = %q, want %q", string(encoded), "+OK\\r\\n")
	}
}

func BenchmarkRESPParse(b *testing.B) {
	payload := []byte("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")
	for i := 0; i < b.N; i++ {
		reader := bufio.NewReader(bytes.NewReader(payload))
		if _, err := Parse(reader); err != nil {
			b.Fatalf("Parse() error = %v", err)
		}
	}
}
