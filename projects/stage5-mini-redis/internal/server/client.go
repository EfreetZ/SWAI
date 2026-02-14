package server

import "net"

// Client 客户端连接。
type Client struct {
	Conn net.Conn
}
