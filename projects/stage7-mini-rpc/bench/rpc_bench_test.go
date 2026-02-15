package bench

import (
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

func BenchmarkRPCCall(b *testing.B) {
	addr := "127.0.0.1:18083"
	srv := server.NewServer(addr, registry.NewMemoryRegistry(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	_ = srv.Register(&svc{})
	go func() { _ = srv.Start() }()
	time.Sleep(100 * time.Millisecond)
	defer func() { _ = srv.Stop() }()

	cli, err := client.Dial(addr)
	if err != nil {
		b.Fatalf("dial failed: %v", err)
	}
	defer func() { _ = cli.Close() }()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out reply
		if err = cli.Call(nil, "svc", "Add", &args{A: i, B: i}, &out); err != nil {
			b.Fatalf("call failed: %v", err)
		}
	}
}
