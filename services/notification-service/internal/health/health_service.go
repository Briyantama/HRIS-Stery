package health

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc/health/grpc_health_v1"
	"go.uber.org/zap"
)

// HealthService implements grpc.health.v1.HealthServer
type HealthService struct {
	grpc_health_v1.UnimplementedHealthServer
	pgPool *pgxpool.Pool
	nc     *nats.Conn
	logger *zap.Logger
}

func NewHealthService(pgPool *pgxpool.Pool, nc *nats.Conn, logger *zap.Logger) *HealthService {
	return &HealthService{
		pgPool: pgPool,
		nc:     nc,
		logger: logger,
	}
}

// Check implements Health.Check
func (h *HealthService) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	status := grpc_health_v1.HealthCheckResponse_SERVING

	// Verify PostgreSQL connectivity
	if h.pgPool != nil {
		if err := h.pgPool.Ping(ctx); err != nil {
			h.logger.Error("PostgreSQL health check failed", zap.Error(err))
			status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
		}
	}

	// Verify NATS connectivity
	if h.nc != nil && status == grpc_health_v1.HealthCheckResponse_SERVING {
		if h.nc.IsClosed() {
			h.logger.Error("NATS connection is closed")
			status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
		}
	}

	return &grpc_health_v1.HealthCheckResponse{Status: status}, nil
}

// Watch implements Health.Watch
func (h *HealthService) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	// Send initial status
	status := grpc_health_v1.HealthCheckResponse_SERVING

	// Verify dependencies
	ctx := stream.Context()
	if h.pgPool != nil {
		if err := h.pgPool.Ping(ctx); err != nil {
			status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
		}
	}

	if h.nc != nil && status == grpc_health_v1.HealthCheckResponse_SERVING {
		if h.nc.IsClosed() {
			status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
		}
	}

	// Send initial response
	if err := stream.Send(&grpc_health_v1.HealthCheckResponse{Status: status}); err != nil {
		h.logger.Error("failed to send health check response", zap.Error(err))
		return fmt.Errorf("failed to send health check response: %w", err)
	}

	// Keep the stream open for potential future status changes
	// In a production system, you might want to periodically check and send updates
	<-ctx.Done()
	return nil
}
