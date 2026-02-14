package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/protocol"
)

func TestTCPServerHandleProduceFetch(t *testing.T) {
	addr := "127.0.0.1:19192"
	b := broker.NewBroker(1, addr, t.TempDir())
	_ = b.CreateTopic("events", broker.TopicConfig{NumPartitions: 1, SegmentBytes: 1024})
	h := NewHandler(b)
	s := NewTCPServer(addr, h, slog.New(slog.NewTextHandler(io.Discard, nil)))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = s.Start(ctx) }()
	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer func() { _ = conn.Close() }()

	produceReq := &protocol.Request{APIKey: protocol.ProduceRequest, CorrelationID: 1, ClientID: "c1", Body: []byte("events|0|k|v")}
	encoded, _ := protocol.EncodeRequest(produceReq)
	_, _ = conn.Write(encoded)
	if _, err := protocol.DecodeResponse(conn); err != nil {
		t.Fatalf("decode produce response failed: %v", err)
	}

	fetchReq := &protocol.Request{APIKey: protocol.FetchRequest, CorrelationID: 2, ClientID: "c1", Body: []byte("events|0|0")}
	encoded, _ = protocol.EncodeRequest(fetchReq)
	_, _ = conn.Write(encoded)
	resp, err := protocol.DecodeResponse(conn)
	if err != nil {
		t.Fatalf("decode fetch response failed: %v", err)
	}
	if string(resp.Body) != "k|v" {
		t.Fatalf("unexpected fetch body: %s", string(resp.Body))
	}
}
