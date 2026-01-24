package async

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	traceParentKey = "traceparent"
	traceStateKey  = "tracestate"
)

// TraceCarrier stores W3C trace context for async task propagation.
type TraceCarrier struct {
	TraceParent string `json:"traceparent,omitempty"`
	TraceState  string `json:"tracestate,omitempty"`
}

// InjectTrace stores the current trace context into the carrier.
func InjectTrace(ctx context.Context, carrier *TraceCarrier) {
	if carrier == nil {
		return
	}
	sc := trace.SpanFromContext(ctx).SpanContext()
	if !sc.IsValid() {
		return
	}
	m := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, m)
	carrier.TraceParent = m[traceParentKey]
	carrier.TraceState = m[traceStateKey]
}

// ExtractTrace restores the trace context from the carrier.
func ExtractTrace(ctx context.Context, carrier TraceCarrier) context.Context {
	if carrier.TraceParent == "" && carrier.TraceState == "" {
		return ctx
	}
	m := propagation.MapCarrier{}
	if carrier.TraceParent != "" {
		m[traceParentKey] = carrier.TraceParent
	}
	if carrier.TraceState != "" {
		m[traceStateKey] = carrier.TraceState
	}
	return otel.GetTextMapPropagator().Extract(ctx, m)
}
