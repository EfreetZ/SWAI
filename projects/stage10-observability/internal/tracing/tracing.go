package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type traceKey struct{}

// SpanContext 链路上下文。
type SpanContext struct {
	TraceID string
	SpanID  string
}

// StartSpan 启动 span。
func StartSpan(ctx context.Context) (context.Context, SpanContext) {
	if ctx == nil {
		ctx = context.Background()
	}
	parent := FromContext(ctx)
	traceID := parent.TraceID
	if traceID == "" {
		traceID = newID(16)
	}
	span := SpanContext{TraceID: traceID, SpanID: newID(8)}
	return context.WithValue(ctx, traceKey{}, span), span
}

// FromContext 读取 span。
func FromContext(ctx context.Context) SpanContext {
	if ctx == nil {
		return SpanContext{}
	}
	v, _ := ctx.Value(traceKey{}).(SpanContext)
	return v
}

func newID(n int) string {
	buf := make([]byte, n)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
