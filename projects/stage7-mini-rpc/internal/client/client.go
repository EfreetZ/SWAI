package client

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/protocol"
	"github.com/EfreetZ/SWAI/projects/stage7-mini-rpc/internal/resilience"
)

// Call RPC 调用。
type Call struct {
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Error         error
	Done          chan *Call
}

// Client RPC 客户端。
type Client struct {
	conn    net.Conn
	codec   protocol.Codec
	reqID   uint64
	mu      sync.Mutex
	closing bool
}

// Dial 建立客户端连接。
func Dial(addr string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, codec: protocol.GetCodec(protocol.JSON)}, nil
}

// Close 关闭连接。
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing {
		return nil
	}
	c.closing = true
	return c.conn.Close()
}

// Call 同步调用。
func (c *Client) Call(ctx context.Context, service, method string, args, reply interface{}) error {
	return resilience.CallWithTimeout(ctx, func() error {
		call := c.Go(service, method, args, reply, make(chan *Call, 1))
		done := <-call.Done
		return done.Error
	}, 2*time.Second)
}

// Go 异步调用。
func (c *Client) Go(service, method string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 1)
	}
	call := &Call{ServiceMethod: service + "." + method, Args: args, Reply: reply, Done: done}
	go func() {
		call.Error = c.invoke(service, method, args, reply)
		done <- call
	}()
	return call
}

func (c *Client) invoke(service, method string, args, reply interface{}) error {
	argsBytes, err := json.Marshal(args)
	if err != nil {
		return err
	}
	req := &protocol.RPCRequest{ServiceName: service, MethodName: method, Args: argsBytes, Metadata: map[string]string{}}
	payload, err := c.codec.Encode(req)
	if err != nil {
		return err
	}

	requestID := atomic.AddUint64(&c.reqID, 1)
	h := protocol.Header{Magic: protocol.MagicNumber, Version: 1, Type: protocol.Request, Codec: protocol.JSON, RequestID: requestID, PayloadLength: uint32(len(payload))}

	c.mu.Lock()
	if c.closing {
		c.mu.Unlock()
		return errors.New("client closed")
	}
	if _, err = c.conn.Write(protocol.EncodeHeader(h)); err != nil {
		c.mu.Unlock()
		return err
	}
	if _, err = c.conn.Write(payload); err != nil {
		c.mu.Unlock()
		return err
	}
	reader := bufio.NewReader(c.conn)
	head := make([]byte, protocol.HeaderSize)
	if _, err = io.ReadFull(reader, head); err != nil {
		c.mu.Unlock()
		return err
	}
	respHeader := protocol.DecodeHeader(head)
	respPayload := make([]byte, respHeader.PayloadLength)
	if _, err = io.ReadFull(reader, respPayload); err != nil {
		c.mu.Unlock()
		return err
	}
	c.mu.Unlock()

	resp := &protocol.RPCResponse{}
	if err = c.codec.Decode(respPayload, resp); err != nil {
		return err
	}
	if resp.Error != "" {
		return errors.New(resp.Error)
	}
	return json.Unmarshal(resp.Data, reply)
}
