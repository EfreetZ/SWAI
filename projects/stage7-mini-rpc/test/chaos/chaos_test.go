package chaos

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/client"
	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/server"
)

type cArgs struct {
	A int `json:"a"`
	B int `json:"b"`
}

type cReply struct {
	Sum int `json:"sum"`
}

type cSvc struct{}

func (s *cSvc) Add(in *cArgs, out *cReply) error {
	out.Sum = in.A + in.B
	return nil
}

func TestServerStopChaos(t *testing.T) {
	addr := "127.0.0.1:18086"
	srv := server.NewServer(addr, registry.NewMemoryRegistry(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := srv.Register(&cSvc{}); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	go func() { _ = srv.Start() }()
	time.Sleep(100 * time.Millisecond)

	cli, err := client.Dial(addr)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer func() { _ = cli.Close() }()

	if err = srv.Stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}

	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		var out cReply
		err = cli.Call(nil, "cSvc", "Add", &cArgs{A: 1, B: 2}, &out)
		if err != nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("expected call fail after stop")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
