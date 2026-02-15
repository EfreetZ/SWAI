package protocol

import "testing"

func TestHeaderEncodeDecode(t *testing.T) {
	h := Header{Magic: MagicNumber, Version: 1, Type: Request, Codec: JSON, RequestID: 7, PayloadLength: 12}
	encoded := EncodeHeader(h)
	decoded := DecodeHeader(encoded)
	if decoded != h {
		t.Fatalf("header mismatch: %+v vs %+v", decoded, h)
	}
}

func TestJSONCodec(t *testing.T) {
	codec := &JSONCodec{}
	payload, err := codec.Encode(map[string]int{"a": 1})
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	var out map[string]int
	if err = codec.Decode(payload, &out); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if out["a"] != 1 {
		t.Fatalf("unexpected value: %v", out)
	}
}
