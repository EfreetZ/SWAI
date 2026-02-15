package tracing

import (
	"context"
	"testing"
)

func TestStartSpan(t *testing.T) {
	ctx, span := StartSpan(context.Background())
	if span.TraceID == "" || span.SpanID == "" {
		t.Fatal("empty span ids")
	}
	nextCtx, child := StartSpan(ctx)
	_ = nextCtx
	if child.TraceID != span.TraceID {
		t.Fatal("trace id should be propagated")
	}
}
