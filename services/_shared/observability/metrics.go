package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// MetricsRegistry holds all service metrics
type MetricsRegistry struct {
	// RPC metrics
	RPCRequestsTotal   metric.Int64Counter
	RPCDurationSeconds metric.Float64Histogram

	// Database metrics
	DBQueryDuration metric.Float64Histogram

	// NATS metrics
	NATSMessagesPublished metric.Int64Counter
	NATSMessagesConsumed  metric.Int64Counter

	// Email metrics
	SMTPSendDuration metric.Float64Histogram

	logger *zap.Logger
}

// InitMetrics initializes all metrics
func InitMetrics(ctx context.Context, logger *zap.Logger) (*MetricsRegistry, error) {
	meter := otel.Meter("hris-stery")

	// RPC request counter
	rpcTotal, err := meter.Int64Counter("rpc_requests_total",
		metric.WithDescription("Total RPC requests"),
		metric.WithUnit("{requests}"),
	)
	if err != nil {
		logger.Warn("failed to create rpc_requests_total metric", zap.Error(err))
	}

	// RPC duration histogram
	rpcDuration, err := meter.Float64Histogram("rpc_duration_seconds",
		metric.WithDescription("RPC request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		logger.Warn("failed to create rpc_duration_seconds metric", zap.Error(err))
	}

	// DB query duration histogram
	dbDuration, err := meter.Float64Histogram("db_query_duration_seconds",
		metric.WithDescription("Database query duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		logger.Warn("failed to create db_query_duration_seconds metric", zap.Error(err))
	}

	// NATS published counter
	natsPublished, err := meter.Int64Counter("nats_messages_published_total",
		metric.WithDescription("Total NATS messages published"),
		metric.WithUnit("{messages}"),
	)
	if err != nil {
		logger.Warn("failed to create nats_messages_published_total metric", zap.Error(err))
	}

	// NATS consumed counter
	natsConsumed, err := meter.Int64Counter("nats_messages_consumed_total",
		metric.WithDescription("Total NATS messages consumed"),
		metric.WithUnit("{messages}"),
	)
	if err != nil {
		logger.Warn("failed to create nats_messages_consumed_total metric", zap.Error(err))
	}

	// SMTP send duration histogram
	smtpDuration, err := meter.Float64Histogram("smtp_send_duration_seconds",
		metric.WithDescription("SMTP email send duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		logger.Warn("failed to create smtp_send_duration_seconds metric", zap.Error(err))
	}

	logger.Info("Metrics initialized successfully")
	return &MetricsRegistry{
		RPCRequestsTotal:      rpcTotal,
		RPCDurationSeconds:    rpcDuration,
		DBQueryDuration:       dbDuration,
		NATSMessagesPublished: natsPublished,
		NATSMessagesConsumed:  natsConsumed,
		SMTPSendDuration:      smtpDuration,
		logger:                logger,
	}, nil
}

// RecordRPCRequest records an RPC request with attributes
func (m *MetricsRegistry) RecordRPCRequest(ctx context.Context, endpoint string, sc SpanContext, durationSec float64, success bool) {
	if m.RPCRequestsTotal == nil || m.RPCDurationSeconds == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("endpoint", endpoint),
		attribute.String(TenantIDKey, sc.TenantID),
		attribute.Bool("success", success),
	}

	m.RPCRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.RPCDurationSeconds.Record(ctx, durationSec, metric.WithAttributes(attrs...))
}

// RecordDBQuery records a database query execution
func (m *MetricsRegistry) RecordDBQuery(ctx context.Context, operation string, sc SpanContext, durationSec float64, success bool) {
	if m.DBQueryDuration == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("operation", operation),
		attribute.String(TenantIDKey, sc.TenantID),
		attribute.Bool("success", success),
	}

	m.DBQueryDuration.Record(ctx, durationSec, metric.WithAttributes(attrs...))
}

// RecordNATSPublish records a NATS message publish
func (m *MetricsRegistry) RecordNATSPublish(ctx context.Context, subject string, sc SpanContext, success bool) {
	if m.NATSMessagesPublished == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("subject", subject),
		attribute.String(TenantIDKey, sc.TenantID),
		attribute.Bool("success", success),
	}

	m.NATSMessagesPublished.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordNATSConsume records a NATS message consumption
func (m *MetricsRegistry) RecordNATSConsume(ctx context.Context, subject string, sc SpanContext, durationSec float64, success bool) {
	if m.NATSMessagesConsumed == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("subject", subject),
		attribute.String(TenantIDKey, sc.TenantID),
		attribute.Bool("success", success),
	}

	m.NATSMessagesConsumed.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordSMTPSend records SMTP email sending
func (m *MetricsRegistry) RecordSMTPSend(ctx context.Context, sc SpanContext, durationSec float64, success bool) {
	if m.SMTPSendDuration == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String(TenantIDKey, sc.TenantID),
		attribute.Bool("success", success),
	}

	m.SMTPSendDuration.Record(ctx, durationSec, metric.WithAttributes(attrs...))
}
