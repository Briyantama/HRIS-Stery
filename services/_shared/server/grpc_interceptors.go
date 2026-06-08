package server

import (
	"context"
	"time"

	"github.com/hris-stery/hris-stery/services/_shared/observability"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerLoggingInterceptor logs entry and exit of unary RPC calls
func UnaryServerLoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract context information
		requestID := observability.RequestIDFromContext(ctx)
		tenantID := observability.TenantIDFromContext(ctx)
		userID := observability.UserIDFromContext(ctx)

		fields := observability.StandardFields{
			RequestID: requestID,
			TenantID:  tenantID,
			UserID:    userID,
			Operation: info.FullMethod,
		}

		// Log entry
		observability.LogHandlerEntry(logger, info.FullMethod, fields)

		// Call handler
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// Log exit
		if err != nil {
			st, _ := status.FromError(err)
			fields.ErrorDetail = st.Code().String()
			observability.LogHandlerExit(logger, info.FullMethod, duration, err, fields)
		} else {
			observability.LogHandlerExit(logger, info.FullMethod, duration, nil, fields)
		}

		return resp, err
	}
}

// StreamServerLoggingInterceptor logs stream RPC calls
func StreamServerLoggingInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Extract context information
		ctx := ss.Context()
		requestID := observability.RequestIDFromContext(ctx)
		tenantID := observability.TenantIDFromContext(ctx)
		userID := observability.UserIDFromContext(ctx)

		fields := observability.StandardFields{
			RequestID: requestID,
			TenantID:  tenantID,
			UserID:    userID,
			Operation: info.FullMethod,
		}

		// Log entry
		observability.LogHandlerEntry(logger, info.FullMethod, fields)

		// Call handler
		start := time.Now()
		err := handler(srv, ss)
		duration := time.Since(start)

		// Log exit
		observability.LogHandlerExit(logger, info.FullMethod, duration, err, fields)

		return err
	}
}

// ServerOptionsWithLogging returns gRPC server options with logging interceptors
func ServerOptionsWithLogging(logger *zap.Logger) []grpc.ServerOption {
	opts := DefaultGRPCServerOptions()
	opts = append(opts,
		grpc.UnaryInterceptor(UnaryServerLoggingInterceptor(logger)),
		grpc.StreamInterceptor(StreamServerLoggingInterceptor(logger)),
	)
	return opts
}
