package protocol

import (
	"bytes"
	"testing"
)

func TestCodecRequestRoundTrip(t *testing.T) {
	req := &Request{APIKey: ProduceRequest, CorrelationID: 7, ClientID: "c1", Body: []byte("hello")}
	encoded, err := EncodeRequest(req)
	if err != nil {
		t.Fatalf("encode request failed: %v", err)
	}
	decoded, err := DecodeRequest(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("decode request failed: %v", err)
	}
	if decoded.APIKey != req.APIKey || decoded.CorrelationID != req.CorrelationID || decoded.ClientID != req.ClientID || string(decoded.Body) != string(req.Body) {
		t.Fatal("request round trip mismatch")
	}
}

func TestCodecResponseRoundTrip(t *testing.T) {
	resp := &Response{CorrelationID: 8, ErrorCode: ErrCodeNone, Body: []byte("ok")}
	encoded, err := EncodeResponse(resp)
	if err != nil {
		t.Fatalf("encode response failed: %v", err)
	}
	decoded, err := DecodeResponse(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if decoded.CorrelationID != resp.CorrelationID || decoded.ErrorCode != resp.ErrorCode || string(decoded.Body) != string(resp.Body) {
		t.Fatal("response round trip mismatch")
	}
}
