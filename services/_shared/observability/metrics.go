package observability

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MetricsRegistry holds Prometheus metrics
type MetricsRegistry struct {
	// RPC metrics
	rpcRequestsTotal   prometheus.Counter
	rpcDurationSeconds prometheus.Histogram
	rpcErrorsTotal     prometheus.Counter

	// Database metrics
	dbQueryDurationMs prometheus.Histogram
	dbErrorsTotal     prometheus.Counter

	// NATS metrics
	natsMessagesPublished prometheus.Counter
	natsMessagesConsumed  prometheus.Counter
	natsErrorsTotal       prometheus.Counter
}

// NewMetricsRegistry initializes a new metrics registry
func NewMetricsRegistry(serviceName string) (*MetricsRegistry, error) {
	m := &MetricsRegistry{
		rpcRequestsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "hris",
			Subsystem: "grpc",
			Name:      "requests_total",
			Help:      "Total number of RPC requests",
			ConstLabels: prometheus.Labels{
				"service": serviceName,
			},
		}),
		rpcDurationSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "hris",
			Subsystem: "grpc",
			Name:      "duration_seconds",
			Help:      "RPC request duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
			ConstLabels: prometheus.Labels{
				"service": serviceName,
			},
		}),
		rpcErrorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "hris",
			Subsystem: "grpc",
			Name:      "errors_total",
			Help:      "Total number of RPC errors",
			ConstLabels: prometheus.Labels{
				"service": serviceName,
			},
		}),
		dbQueryDurationMs: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "hris",
			Subsystem: "db",
			Name:      "query_duration_ms",
			Help:      "Database query duration in milliseconds",
			Buckets:   []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500},
			ConstLabels: prometheus.Labels{
				"service": serviceName,
			},
		}),
		dbErrorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "hris",
			Subsystem: "db",
			Name:      "errors_total",
			Help:      "Total number of database errors",
			ConstLabels: prometheus.Labels{
				"service": serviceName,
			},
		}),
		natsMessagesPublished: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "hris",
			Subsystem: "nats",
			Name:      "messages_published_total",
			Help:      "Total number of NATS messages published",
			ConstLabels: prometheus.Labels{
				"service": serviceName,
			},
		}),
		natsMessagesConsumed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "hris",
			Subsystem: "nats",
			Name:      "messages_consumed_total",
			Help:      "Total number of NATS messages consumed",
			ConstLabels: prometheus.Labels{
				"service": serviceName,
			},
		}),
		natsErrorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "hris",
			Subsystem: "nats",
			Name:      "errors_total",
			Help:      "Total number of NATS errors",
			ConstLabels: prometheus.Labels{
				"service": serviceName,
			},
		}),
	}

	// Register metrics with default registry
	if err := prometheus.Register(m.rpcRequestsTotal); err != nil {
		return nil, fmt.Errorf("register rpcRequestsTotal: %w", err)
	}
	if err := prometheus.Register(m.rpcDurationSeconds); err != nil {
		return nil, fmt.Errorf("register rpcDurationSeconds: %w", err)
	}
	if err := prometheus.Register(m.rpcErrorsTotal); err != nil {
		return nil, fmt.Errorf("register rpcErrorsTotal: %w", err)
	}
	if err := prometheus.Register(m.dbQueryDurationMs); err != nil {
		return nil, fmt.Errorf("register dbQueryDurationMs: %w", err)
	}
	if err := prometheus.Register(m.dbErrorsTotal); err != nil {
		return nil, fmt.Errorf("register dbErrorsTotal: %w", err)
	}
	if err := prometheus.Register(m.natsMessagesPublished); err != nil {
		return nil, fmt.Errorf("register natsMessagesPublished: %w", err)
	}
	if err := prometheus.Register(m.natsMessagesConsumed); err != nil {
		return nil, fmt.Errorf("register natsMessagesConsumed: %w", err)
	}
	if err := prometheus.Register(m.natsErrorsTotal); err != nil {
		return nil, fmt.Errorf("register natsErrorsTotal: %w", err)
	}

	return m, nil
}

// RecordRPCRequest records an RPC request
func (m *MetricsRegistry) RecordRPCRequest() {
	m.rpcRequestsTotal.Inc()
}

// RecordRPCDuration records RPC request duration
func (m *MetricsRegistry) RecordRPCDuration(duration time.Duration) {
	m.rpcDurationSeconds.Observe(duration.Seconds())
}

// RecordRPCError records an RPC error
func (m *MetricsRegistry) RecordRPCError() {
	m.rpcErrorsTotal.Inc()
}

// RecordDBQuery records database query duration
func (m *MetricsRegistry) RecordDBQuery(duration time.Duration) {
	m.dbQueryDurationMs.Observe(float64(duration.Milliseconds()))
}

// RecordDBError records a database error
func (m *MetricsRegistry) RecordDBError() {
	m.dbErrorsTotal.Inc()
}

// RecordNATSMessagePublished records a published NATS message
func (m *MetricsRegistry) RecordNATSMessagePublished() {
	m.natsMessagesPublished.Inc()
}

// RecordNATSMessageConsumed records a consumed NATS message
func (m *MetricsRegistry) RecordNATSMessageConsumed() {
	m.natsMessagesConsumed.Inc()
}

// RecordNATSError records a NATS error
func (m *MetricsRegistry) RecordNATSError() {
	m.natsErrorsTotal.Inc()
}

// StartMetricsServer starts the Prometheus metrics HTTP server on port 8081
func StartMetricsServer(logger *zap.Logger) {
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		listener, err := net.Listen("tcp", ":8081")
		if err != nil {
			logger.Error("failed to listen on metrics port", zap.Error(err))
			return
		}
		defer listener.Close()

		logger.Info("starting metrics server", zap.String("address", "0.0.0.0:8081"))
		if err := http.Serve(listener, nil); err != nil {
			logger.Error("metrics server error", zap.Error(err))
		}
	}()
}
