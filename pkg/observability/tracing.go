package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracer returns a tracer for the given name.
// Use this to create spans in application code.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// StartSpan starts a new span with the given name and returns the new context and span.
// The caller is responsible for calling span.End().
//
// Example:
//
//	ctx, span := observability.StartSpan(ctx, "auth", "Register")
//	defer span.End()
func StartSpan(ctx context.Context, tracerName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, spanName, opts...)
}

// RecordError records an error on the span and sets the span status to Error.
func RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetSpanAttributes sets attributes on the span.
func SetSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// SpanFromContext returns the span from the context.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// WithSpan is a helper that wraps a function with a span.
// It automatically records any errors and ends the span.
//
// Example:
//
//	result, err := observability.WithSpan(ctx, "auth", "Register", func(ctx context.Context) (*AuthResult, error) {
//	    return s.doRegister(ctx, email, password)
//	})
func WithSpan[T any](ctx context.Context, tracerName, spanName string, fn func(context.Context) (T, error)) (T, error) {
	ctx, span := StartSpan(ctx, tracerName, spanName)
	defer span.End()

	result, err := fn(ctx)
	if err != nil {
		RecordError(span, err)
	}
	return result, err
}

// WithSpanVoid is like WithSpan but for functions that don't return a value.
func WithSpanVoid(ctx context.Context, tracerName, spanName string, fn func(context.Context) error) error {
	ctx, span := StartSpan(ctx, tracerName, spanName)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		RecordError(span, err)
	}
	return err
}
