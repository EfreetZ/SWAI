package transport

import (
	"net"
	"time"
)

// DialTCP 连接 TCP。
func DialTCP(addr string, timeout time.Duration) (net.Conn, error) {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return net.DialTimeout("tcp", addr, timeout)
}
