package tcp

import (
	"bufio"
	"context"
	"net"
	"testing"
	"time"
)

func TestEchoServerRoundTrip(t *testing.T) {
	server := NewEchoServer("127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start(ctx)
	}()

	addr, err := waitForAddress(server, time.Second)
	if err != nil {
		t.Fatalf("waitForAddress() error = %v", err)
	}

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("DialTimeout() error = %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.Write([]byte("hello\n"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatalf("ReadString() error = %v", err)
	}
	if line != "hello\n" {
		t.Fatalf("echo line = %q, want %q", line, "hello\\n")
	}

	cancel()
	select {
	case runErr := <-errCh:
		if runErr != nil {
			t.Fatalf("Start() error = %v", runErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not exit in time")
	}
}

func TestEchoServerStartWithCanceledContext(t *testing.T) {
	server := NewEchoServer("127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := server.Start(ctx); err == nil {
		t.Fatal("Start() error = nil, want context canceled")
	}
}

func BenchmarkEchoServerRoundTrip(b *testing.B) {
	server := NewEchoServer("127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start(ctx)
	}()

	addr, err := waitForAddress(server, time.Second)
	if err != nil {
		b.Fatalf("waitForAddress() error = %v", err)
	}

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		b.Fatalf("DialTimeout() error = %v", err)
	}
	defer func() {
		_ = conn.Close()
		cancel()
		<-errCh
	}()

	reader := bufio.NewReader(conn)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, writeErr := conn.Write([]byte("ping\n")); writeErr != nil {
			b.Fatalf("Write() error = %v", writeErr)
		}
		if _, readErr := reader.ReadString('\n'); readErr != nil {
			b.Fatalf("ReadString() error = %v", readErr)
		}
	}
}

func waitForAddress(server *EchoServer, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		addr := server.Addr()
		if addr != "" && addr != "127.0.0.1:0" {
			return addr, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return "", context.DeadlineExceeded
}
