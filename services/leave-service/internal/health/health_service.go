package health

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// HealthService implements grpc.health.v1.HealthServer
type HealthService struct {
	grpc_health_v1.UnimplementedHealthServer
	logger *zap.Logger
}

func NewHealthService(logger *zap.Logger) *HealthService {
	return &HealthService{
		logger: logger,
	}
}

// Check implements Health.Check
func (h *HealthService) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

// Watch implements Health.Watch
func (h *HealthService) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	// Send initial response
	if err := stream.Send(&grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}); err != nil {
		h.logger.Error("failed to send health check response", zap.Error(err))
		return fmt.Errorf("failed to send health check response: %w", err)
	}

	// Keep the stream open for potential future status changes
	<-stream.Context().Done()
	return nil
}
