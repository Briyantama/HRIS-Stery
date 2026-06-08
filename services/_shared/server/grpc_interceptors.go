package server

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hris-stery/hris-stery/services/_shared/observability"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerRequestIDInterceptor generates request_id for tracing if not present
func UnaryServerRequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		requestID := observability.RequestIDFromContext(ctx)
		if requestID == "" {
			requestID = uuid.New().String()
			ctx = context.WithValue(ctx, "request_id", requestID)
		}
		return handler(ctx, req)
	}
}

// StreamServerRequestIDInterceptor generates request_id for tracing if not present
func StreamServerRequestIDInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		requestID := observability.RequestIDFromContext(ctx)
		if requestID == "" {
			requestID = uuid.New().String()
			ctx = context.WithValue(ctx, "request_id", requestID)
		}
		return handler(srv, &wrappedStream{ServerStream: ss, ctx: ctx})
	}
}

// wrappedStream wraps grpc.ServerStream to replace its context
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

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

// ServerOptionsWithLogging returns gRPC server options with request tracing and logging
func ServerOptionsWithLogging(logger *zap.Logger) []grpc.ServerOption {
	opts := DefaultGRPCServerOptions()
	opts = append(opts,
		// Request ID generation (outermost to catch all requests)
		grpc.ChainUnaryInterceptor(
			UnaryServerRequestIDInterceptor(),
			UnaryServerLoggingInterceptor(logger),
		),
		grpc.ChainStreamInterceptor(
			StreamServerRequestIDInterceptor(),
			StreamServerLoggingInterceptor(logger),
		),
	)
	return opts
}
