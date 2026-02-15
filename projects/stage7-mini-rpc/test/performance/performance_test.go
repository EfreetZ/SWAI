package performance

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/client"
	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/server"
)

type perfArgs struct {
	A int `json:"a"`
	B int `json:"b"`
}

type perfReply struct {
	Sum int `json:"sum"`
}

type perfSvc struct{}

func (s *perfSvc) Add(in *perfArgs, out *perfReply) error {
	out.Sum = in.A + in.B
	return nil
}

func TestRPCPerformance(t *testing.T) {
	addr := "127.0.0.1:18084"
	srv := server.NewServer(addr, registry.NewMemoryRegistry(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := srv.Register(&perfSvc{}); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	go func() { _ = srv.Start() }()
	time.Sleep(100 * time.Millisecond)
	defer func() { _ = srv.Stop() }()

	cli, err := client.Dial(addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer func() { _ = cli.Close() }()

	start := time.Now()
	for i := 0; i < 2000; i++ {
		var out perfReply
		if err = cli.Call(nil, "perfSvc", "Add", &perfArgs{A: i, B: i}, &out); err != nil {
			t.Fatalf("call failed: %v", err)
		}
	}
	if time.Since(start) > 3*time.Second {
		t.Fatalf("performance regression: %s", time.Since(start))
	}
}
