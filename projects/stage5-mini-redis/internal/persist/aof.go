package persist

import (
	"bufio"
	"context"
	"os"
	"strings"
	"sync"
)

// AOF append-only file。
type AOF struct {
	mu     sync.Mutex
	file   *os.File
	writer *bufio.Writer
}

// NewAOF 创建 AOF。
func NewAOF(path string) (*AOF, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	return &AOF{file: file, writer: bufio.NewWriter(file)}, nil
}

// Close 关闭 AOF。
func (a *AOF) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.writer.Flush(); err != nil {
		return err
	}
	return a.file.Close()
}

// Append 追加命令。
func (a *AOF) Append(ctx context.Context, command []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	line := strings.Join(command, " ") + "\n"
	if _, err := a.writer.WriteString(line); err != nil {
		return err
	}
	return nil
}

// Flush 刷盘。
func (a *AOF) Flush(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.writer.Flush(); err != nil {
		return err
	}
	return a.file.Sync()
}

// Replay 读取全部命令行。
func (a *AOF) Replay(ctx context.Context) ([][]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, err := a.file.Seek(0, 0); err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(a.file)
	commands := make([][]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		commands = append(commands, strings.Fields(line))
	}
	if _, err := a.file.Seek(0, 2); err != nil {
		return nil, err
	}
	return commands, scanner.Err()
}
