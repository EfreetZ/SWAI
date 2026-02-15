package test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/client"
	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/server"
)

type args struct {
	A int `json:"a"`
	B int `json:"b"`
}

type reply struct {
	Sum int `json:"sum"`
}

type svc struct{}

func (s *svc) Add(in *args, out *reply) error {
	out.Sum = in.A + in.B
	return nil
}

func TestRPCIntegration(t *testing.T) {
	addr := "127.0.0.1:18082"
	srv := server.NewServer(addr, registry.NewMemoryRegistry(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := srv.Register(&svc{}); err != nil {
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

	proxy := client.NewProxy(cli, "svc")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var out reply
	if err = proxy.Invoke(ctx, "Add", &args{A: 5, B: 6}, &out); err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if out.Sum != 11 {
		t.Fatalf("unexpected sum: %d", out.Sum)
	}
}
