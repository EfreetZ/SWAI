package client

import "context"

// Proxy 简单代理。
type Proxy struct {
	client  *Client
	service string
}

// NewProxy 创建服务代理。
func NewProxy(c *Client, service string) *Proxy {
	return &Proxy{client: c, service: service}
}

// Invoke 调用指定方法。
func (p *Proxy) Invoke(ctx context.Context, method string, args, reply interface{}) error {
	return p.client.Call(ctx, p.service, method, args, reply)
}
