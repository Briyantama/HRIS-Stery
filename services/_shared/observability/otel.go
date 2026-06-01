package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// InitTracer initializes OpenTelemetry tracing with OTLP exporter
func InitTracer(ctx context.Context, serviceName, otlpEndpoint string, logger *zap.Logger) (trace.TracerProvider, error) {
	if otlpEndpoint == "" {
		logger.Info("OTEL_EXPORTER_OTLP_ENDPOINT not set, using no-op tracer")
		return trace.NewNoopTracerProvider(), nil
	}

	// Create OTLP exporter
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(otlpEndpoint))
	if err != nil {
		logger.Warn("failed to create OTLP exporter, using no-op tracer", zap.Error(err))
		return trace.NewNoopTracerProvider(), nil
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		logger.Warn("failed to create resource, using no-op tracer", zap.Error(err))
		return trace.NewNoopTracerProvider(), nil
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	logger.Info("OpenTelemetry tracer initialized", zap.String("endpoint", otlpEndpoint))
	return tp, nil
}

// GetTracer returns global tracer for the service
func GetTracer(serviceName string) trace.Tracer {
	return otel.Tracer(serviceName,
		trace.WithInstrumentationVersion("1.0.0"),
	)
}

// Common span attributes
const (
	TenantIDKey   = "tenant_id"
	RequestIDKey  = "request_id"
	ServiceKey    = "service_name"
	OperationKey  = "operation"
	ErrorKey      = "error"
)

// SpanContext holds request-scoped tracing context
type SpanContext struct {
	TenantID  string
	RequestID string
}

// ExtractContext extracts span context from gRPC or HTTP metadata
func ExtractContext(ctx context.Context) SpanContext {
	return SpanContext{
		TenantID:  extractValue(ctx, TenantIDKey),
		RequestID: extractValue(ctx, RequestIDKey),
	}
}

// extractValue is a helper to extract values from context
func extractValue(ctx context.Context, key string) string {
	if v := ctx.Value(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
