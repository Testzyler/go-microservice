package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Span creates a span and returns a defer-friendly end function.
// Use with named return values to automatically capture errors.
//
// Usage:
//
//	func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (result *Order, err error) {
//	    ctx, end := observability.Span(ctx, "GetOrder", observability.AttrID("order", id))
//	    defer end(&err)
//
//	    // Your business logic here - no wrapping needed
//	    return s.repo.GetByID(ctx, id)
//	}
func Span(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, func(*error)) {
	ctx, span := otel.Tracer("app").Start(ctx, name)
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	return ctx, func(errPtr *error) {
		if errPtr != nil && *errPtr != nil {
			span.RecordError(*errPtr)
			span.SetStatus(codes.Error, (*errPtr).Error())
		}
		span.End()
	}
}

// SpanWithTracer is like Span but with a custom tracer name.
func SpanWithTracer(ctx context.Context, tracerName, spanName string, attrs ...attribute.KeyValue) (context.Context, func(*error)) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, spanName)
	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	return ctx, func(errPtr *error) {
		if errPtr != nil && *errPtr != nil {
			span.RecordError(*errPtr)
			span.SetStatus(codes.Error, (*errPtr).Error())
		}
		span.End()
	}
}

// AddSpanAttrs adds attributes to the current span from context.
// Use this to enrich spans with output data after the operation.
func AddSpanAttrs(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		span.SetAttributes(attrs...)
	}
}

// SpanEvent adds an event to the current span.
func SpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// --- Attribute Helpers ---

// Attr creates a string attribute.
func Attr(key, value string) attribute.KeyValue {
	return attribute.String(key, value)
}

// AttrInt creates an int attribute.
func AttrInt(key string, value int) attribute.KeyValue {
	return attribute.Int(key, value)
}

// AttrInt64 creates an int64 attribute.
func AttrInt64(key string, value int64) attribute.KeyValue {
	return attribute.Int64(key, value)
}

// AttrBool creates a bool attribute.
func AttrBool(key string, value bool) attribute.KeyValue {
	return attribute.Bool(key, value)
}

// AttrID creates an ID attribute (converts to string).
// Example: AttrID("order", orderID) → "order.id" = "uuid-value"
func AttrID(name string, id interface{}) attribute.KeyValue {
	var s string
	switch v := id.(type) {
	case string:
		s = v
	case interface{ String() string }:
		s = v.String()
	default:
		s = ""
	}
	return attribute.String(name+".id", s)
}

// AttrError creates an error attribute (if err is not nil).
func AttrError(err error) attribute.KeyValue {
	if err == nil {
		return attribute.String("error", "")
	}
	return attribute.String("error", err.Error())
}
