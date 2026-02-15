package client

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/server"
)

type addArgs struct {
	A int `json:"a"`
	B int `json:"b"`
}

type addReply struct {
	Sum int `json:"sum"`
}

type testSvc struct{}

func (s *testSvc) Add(args *addArgs, reply *addReply) error {
	reply.Sum = args.A + args.B
	return nil
}

func TestClientCall(t *testing.T) {
	addr := "127.0.0.1:18081"
	srv := server.NewServer(addr, registry.NewMemoryRegistry(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := srv.Register(&testSvc{}); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	go func() { _ = srv.Start() }()
	time.Sleep(100 * time.Millisecond)
	defer func() { _ = srv.Stop() }()

	c, err := Dial(addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer func() { _ = c.Close() }()

	var reply addReply
	if err = c.Call(nil, "testSvc", "Add", &addArgs{A: 1, B: 2}, &reply); err != nil {
		t.Fatalf("call failed: %v", err)
	}
	if reply.Sum != 3 {
		t.Fatalf("unexpected sum: %d", reply.Sum)
	}
}
