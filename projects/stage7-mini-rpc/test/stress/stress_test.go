package stress

import (
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/client"
	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/registry"
	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/server"
)

type sArgs struct {
	A int `json:"a"`
	B int `json:"b"`
}

type sReply struct {
	Sum int `json:"sum"`
}

type sSvc struct{}

func (s *sSvc) Add(in *sArgs, out *sReply) error {
	out.Sum = in.A + in.B
	return nil
}

func TestRPCStress(t *testing.T) {
	addr := "127.0.0.1:18085"
	srv := server.NewServer(addr, registry.NewMemoryRegistry(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := srv.Register(&sSvc{}); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	go func() { _ = srv.Start() }()
	time.Sleep(100 * time.Millisecond)
	defer func() { _ = srv.Stop() }()

	const workers = 20
	const each = 200
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			cli, err := client.Dial(addr)
			if err != nil {
				t.Errorf("dial failed: %v", err)
				return
			}
			defer func() { _ = cli.Close() }()
			for i := 0; i < each; i++ {
				var out sReply
				if callErr := cli.Call(nil, "sSvc", "Add", &sArgs{A: i, B: i}, &out); callErr != nil {
					t.Errorf("call failed: %v", callErr)
					return
				}
			}
		}()
	}
	wg.Wait()
}
