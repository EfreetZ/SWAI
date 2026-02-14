package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"strings"

	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/broker"
	"github.com/EfreetZ/SWAI/projects/stage6-mini-mq/internal/protocol"
)

// Handler 处理协议请求。
type Handler struct {
	broker *broker.Broker
}

// NewHandler 创建处理器。
func NewHandler(b *broker.Broker) *Handler {
	return &Handler{broker: b}
}

// HandleRequest 处理请求。
func (h *Handler) HandleRequest(ctx context.Context, req *protocol.Request) *protocol.Response {
	if req == nil {
		return &protocol.Response{ErrorCode: protocol.ErrCodeInvalidReq, Body: []byte("nil request")}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return &protocol.Response{CorrelationID: req.CorrelationID, ErrorCode: protocol.ErrCodeInternal, Body: []byte(err.Error())}
	}

	switch req.APIKey {
	case protocol.ProduceRequest:
		return h.handleProduce(ctx, req)
	case protocol.FetchRequest:
		return h.handleFetch(ctx, req)
	case protocol.MetadataRequest:
		return h.handleMetadata(req)
	default:
		return &protocol.Response{CorrelationID: req.CorrelationID, ErrorCode: protocol.ErrCodeInvalidReq, Body: []byte("unsupported api key")}
	}
}

func (h *Handler) handleProduce(ctx context.Context, req *protocol.Request) *protocol.Response {
	parts := strings.SplitN(string(req.Body), "|", 4)
	if len(parts) < 4 {
		return &protocol.Response{CorrelationID: req.CorrelationID, ErrorCode: protocol.ErrCodeInvalidReq, Body: []byte("invalid produce body")}
	}
	topic := parts[0]
	partition := atoi(parts[1])
	key := []byte(parts[2])
	value := []byte(parts[3])
	offset, err := h.broker.Produce(ctx, topic, partition, key, value)
	if err != nil {
		return &protocol.Response{CorrelationID: req.CorrelationID, ErrorCode: protocol.ErrCodeInternal, Body: []byte(err.Error())}
	}
	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.BigEndian, offset)
	return &protocol.Response{CorrelationID: req.CorrelationID, ErrorCode: protocol.ErrCodeNone, Body: buf.Bytes()}
}

func (h *Handler) handleFetch(ctx context.Context, req *protocol.Request) *protocol.Response {
	parts := strings.SplitN(string(req.Body), "|", 3)
	if len(parts) < 3 {
		return &protocol.Response{CorrelationID: req.CorrelationID, ErrorCode: protocol.ErrCodeInvalidReq, Body: []byte("invalid fetch body")}
	}
	topic := parts[0]
	partition := atoi(parts[1])
	offset := atoi64(parts[2])
	msg, err := h.broker.Fetch(ctx, topic, partition, offset)
	if err != nil {
		return &protocol.Response{CorrelationID: req.CorrelationID, ErrorCode: protocol.ErrCodeOffsetOutOfRange, Body: []byte(err.Error())}
	}
	payload := []byte(string(msg.Key) + "|" + string(msg.Value))
	return &protocol.Response{CorrelationID: req.CorrelationID, ErrorCode: protocol.ErrCodeNone, Body: payload}
}

func (h *Handler) handleMetadata(req *protocol.Request) *protocol.Response {
	topics := h.broker.ListTopics()
	return &protocol.Response{CorrelationID: req.CorrelationID, ErrorCode: protocol.ErrCodeNone, Body: []byte(strings.Join(topics, ","))}
}

func atoi(raw string) int {
	v := 0
	for i := 0; i < len(raw); i++ {
		if raw[i] < '0' || raw[i] > '9' {
			break
		}
		v = v*10 + int(raw[i]-'0')
	}
	return v
}

func atoi64(raw string) int64 {
	var v int64
	for i := 0; i < len(raw); i++ {
		if raw[i] < '0' || raw[i] > '9' {
			break
		}
		v = v*10 + int64(raw[i]-'0')
	}
	return v
}
