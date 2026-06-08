package observability

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// StandardFields contains standard field names for consistent logging across services
type StandardFields struct {
	TenantID    string
	UserID      string
	RequestID   string
	DurationMs  int64
	ErrorDetail string
	Operation   string
	Resource    string
}

// WithTenantID adds tenant_id to logger context
func WithTenantID(logger *zap.Logger, tenantID string) *zap.Logger {
	return logger.With(zap.String("tenant_id", tenantID))
}

// WithUserID adds user_id to logger context
func WithUserID(logger *zap.Logger, userID string) *zap.Logger {
	return logger.With(zap.String("user_id", userID))
}

// WithRequestID adds request_id to logger context
func WithRequestID(logger *zap.Logger, requestID string) *zap.Logger {
	return logger.With(zap.String("request_id", requestID))
}

// WithFields adds multiple standard fields to logger
func WithFields(logger *zap.Logger, fields StandardFields) *zap.Logger {
	opts := []zap.Field{}
	if fields.TenantID != "" {
		opts = append(opts, zap.String("tenant_id", fields.TenantID))
	}
	if fields.UserID != "" {
		opts = append(opts, zap.String("user_id", fields.UserID))
	}
	if fields.RequestID != "" {
		opts = append(opts, zap.String("request_id", fields.RequestID))
	}
	if fields.Operation != "" {
		opts = append(opts, zap.String("operation", fields.Operation))
	}
	if fields.Resource != "" {
		opts = append(opts, zap.String("resource", fields.Resource))
	}
	return logger.With(opts...)
}

// LogHandlerEntry logs entry to gRPC handler
func LogHandlerEntry(logger *zap.Logger, handler string, fields StandardFields) {
	WithFields(logger, fields).Info("handler.entry", zap.String("handler", handler))
}

// LogHandlerExit logs exit from gRPC handler with duration
func LogHandlerExit(logger *zap.Logger, handler string, duration time.Duration, err error, fields StandardFields) {
	durationMs := duration.Milliseconds()
	opts := []zap.Field{
		zap.String("handler", handler),
		zap.Int64("duration_ms", durationMs),
	}
	if err != nil {
		opts = append(opts, zap.Error(err))
		opts = append(opts, zap.String("status", "error"))
	} else {
		opts = append(opts, zap.String("status", "success"))
	}

	WithFields(logger, fields).Info("handler.exit", opts...)
}

// MeasureDuration measures operation duration
func MeasureDuration(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}

// LogSlowQuery logs database queries that exceed threshold
func LogSlowQuery(logger *zap.Logger, operation string, duration time.Duration, query string, fields StandardFields) {
	if duration.Milliseconds() > 100 {
		opts := []zap.Field{
			zap.String("operation", operation),
			zap.Int64("duration_ms", duration.Milliseconds()),
			zap.String("query", query),
		}
		WithFields(logger, fields).Warn("query.slow", opts...)
	}
}

// LogRepositoryEntry logs entry to repository method
func LogRepositoryEntry(logger *zap.Logger, repo string, method string, fields StandardFields) {
	WithFields(logger, fields).Debug("repository.entry",
		zap.String("repository", repo),
		zap.String("method", method),
	)
}

// LogRepositoryExit logs exit from repository method with duration
func LogRepositoryExit(logger *zap.Logger, repo string, method string, duration time.Duration, err error, fields StandardFields) {
	opts := []zap.Field{
		zap.String("repository", repo),
		zap.String("method", method),
		zap.Int64("duration_ms", duration.Milliseconds()),
	}
	if err != nil {
		opts = append(opts, zap.Error(err))
		opts = append(opts, zap.String("status", "error"))
	} else {
		opts = append(opts, zap.String("status", "success"))
	}

	WithFields(logger, fields).Debug("repository.exit", opts...)
}

// LogError logs error with context
func LogError(logger *zap.Logger, message string, err error, fields StandardFields) {
	WithFields(logger, fields).Error(message, zap.Error(err))
}

// RequestIDFromContext extracts request_id from context if present
func RequestIDFromContext(ctx context.Context) string {
	if val := ctx.Value("request_id"); val != nil {
		if rid, ok := val.(string); ok {
			return rid
		}
	}
	return ""
}

// TenantIDFromContext extracts tenant_id from context if present
func TenantIDFromContext(ctx context.Context) string {
	if val := ctx.Value("tenant_id"); val != nil {
		if tid, ok := val.(string); ok {
			return tid
		}
	}
	return ""
}

// UserIDFromContext extracts user_id from context if present
func UserIDFromContext(ctx context.Context) string {
	if val := ctx.Value("user_id"); val != nil {
		if uid, ok := val.(string); ok {
			return uid
		}
	}
	return ""
}
