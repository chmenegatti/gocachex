// Package tracing provides OpenTelemetry tracing support for GoCacheX.
package tracing

import (
	"context"
	"fmt"

	"github.com/chmenegatti/gocachex/pkg/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracer wraps OpenTelemetry tracing functionality.
type Tracer struct {
	config config.TracingConfig
	tracer trace.Tracer
}

// New creates a new tracer instance.
func New(cfg config.TracingConfig) (*Tracer, error) {
	if !cfg.Enabled {
		return &Tracer{config: cfg}, nil
	}

	// Get the global tracer
	tracer := otel.Tracer(cfg.ServiceName)

	return &Tracer{
		config: cfg,
		tracer: tracer,
	}, nil
}

// StartSpan starts a new tracing span.
func (t *Tracer) StartSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	if t == nil || t.tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	return t.tracer.Start(ctx, operationName)
}

// SpanFromContext returns the span from the context.
func (t *Tracer) SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddEvent adds an event to the current span.
func (t *Tracer) AddEvent(ctx context.Context, name string, attributes ...trace.EventOption) {
	if span := trace.SpanFromContext(ctx); span != nil {
		span.AddEvent(name, attributes...)
	}
}

// SetAttributes sets attributes on the current span.
func (t *Tracer) SetAttributes(ctx context.Context, attributes ...trace.SpanStartEventOption) {
	if span := trace.SpanFromContext(ctx); span != nil {
		// Note: This is a simplified implementation
		// In a real implementation, you would convert SpanStartEventOption to Attribute
		_ = attributes
	}
}

// RecordError records an error on the current span.
func (t *Tracer) RecordError(ctx context.Context, err error) {
	if span := trace.SpanFromContext(ctx); span != nil {
		span.RecordError(err)
	}
}

// Close closes the tracer and flushes any pending spans.
func (t *Tracer) Close() error {
	// In a real implementation, you would flush/close the tracer provider
	return nil
}

// IsEnabled returns whether tracing is enabled.
func (t *Tracer) IsEnabled() bool {
	return t != nil && t.config.Enabled && t.tracer != nil
}

// GetServiceName returns the service name for tracing.
func (t *Tracer) GetServiceName() string {
	if t == nil {
		return ""
	}
	return t.config.ServiceName
}

// Helper functions for common tracing patterns

// TraceOperation traces a cache operation.
func (t *Tracer) TraceOperation(ctx context.Context, operation, backend string, fn func(ctx context.Context) error) error {
	if !t.IsEnabled() {
		return fn(ctx)
	}

	spanName := fmt.Sprintf("cache.%s", operation)
	ctx, span := t.StartSpan(ctx, spanName)
	defer span.End()

	// Set attributes
	span.SetAttributes(
		attribute.String("cache.operation", operation),
		attribute.String("cache.backend", backend),
	)

	// Execute the operation
	err := fn(ctx)
	if err != nil {
		t.RecordError(ctx, err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}

// TraceBatchOperation traces a batch cache operation.
func (t *Tracer) TraceBatchOperation(ctx context.Context, operation, backend string, keyCount int, fn func(ctx context.Context) error) error {
	if !t.IsEnabled() {
		return fn(ctx)
	}

	spanName := fmt.Sprintf("cache.%s_multi", operation)
	ctx, span := t.StartSpan(ctx, spanName)
	defer span.End()

	// Set attributes
	span.SetAttributes(
		attribute.String("cache.operation", operation),
		attribute.String("cache.backend", backend),
		attribute.Int("cache.key_count", keyCount),
	)

	// Execute the operation
	err := fn(ctx)
	if err != nil {
		t.RecordError(ctx, err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}

// NoOpTracer is a no-operation tracer for when tracing is disabled.
type NoOpTracer struct{}

// StartSpan is a no-op implementation.
func (n *NoOpTracer) StartSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	return ctx, trace.SpanFromContext(ctx)
}

// SpanFromContext is a no-op implementation.
func (n *NoOpTracer) SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddEvent is a no-op implementation.
func (n *NoOpTracer) AddEvent(ctx context.Context, name string, attributes ...trace.EventOption) {
	// No-op
}

// SetAttributes is a no-op implementation.
func (n *NoOpTracer) SetAttributes(ctx context.Context, attributes ...trace.SpanStartEventOption) {
	// No-op
}

// RecordError is a no-op implementation.
func (n *NoOpTracer) RecordError(ctx context.Context, err error) {
	// No-op
}

// Close is a no-op implementation.
func (n *NoOpTracer) Close() error {
	return nil
}

// IsEnabled returns false for no-op tracer.
func (n *NoOpTracer) IsEnabled() bool {
	return false
}

// GetServiceName returns empty string for no-op tracer.
func (n *NoOpTracer) GetServiceName() string {
	return ""
}

// TraceOperation is a no-op implementation.
func (n *NoOpTracer) TraceOperation(ctx context.Context, operation, backend string, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// TraceBatchOperation is a no-op implementation.
func (n *NoOpTracer) TraceBatchOperation(ctx context.Context, operation, backend string, keyCount int, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
