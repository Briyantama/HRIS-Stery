package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// GracefulShutdown orchestrates graceful shutdown of gRPC server and resources.
type GracefulShutdown struct {
	grpcServer  *grpc.Server
	logger      *zap.Logger
	drainPeriod time.Duration
	cleanupFns  []func(ctx context.Context) error
}

// NewGracefulShutdown creates a new graceful shutdown orchestrator.
func NewGracefulShutdown(grpcServer *grpc.Server, logger *zap.Logger) *GracefulShutdown {
	return &GracefulShutdown{
		grpcServer:  grpcServer,
		logger:      logger,
		drainPeriod: 30 * time.Second,
		cleanupFns:  []func(context.Context) error{},
	}
}

// RegisterCleanup registers a cleanup function to be called during shutdown.
// Functions are called in reverse order (LIFO).
func (g *GracefulShutdown) RegisterCleanup(fn func(ctx context.Context) error) {
	g.cleanupFns = append(g.cleanupFns, fn)
}

// RegisterNATSDrain registers NATS connection for graceful drain.
func (g *GracefulShutdown) RegisterNATSDrain(nc *nats.Conn) {
	g.RegisterCleanup(func(ctx context.Context) error {
		g.logger.Info("draining NATS subscriptions")
		if err := nc.Drain(); err != nil {
			g.logger.Error("failed to drain NATS", zap.Error(err))
			return fmt.Errorf("drain NATS: %w", err)
		}
		g.logger.Info("NATS subscriptions drained")
		return nil
	})
}

// RegisterRedisClose registers Redis client for cleanup.
func (g *GracefulShutdown) RegisterRedisClose(client *redis.Client) {
	g.RegisterCleanup(func(ctx context.Context) error {
		g.logger.Info("closing Redis client")
		if err := client.Close(); err != nil {
			g.logger.Error("failed to close Redis", zap.Error(err))
			return fmt.Errorf("close Redis: %w", err)
		}
		g.logger.Info("Redis client closed")
		return nil
	})
}

// RegisterMinIOClose registers MinIO client for cleanup.
func (g *GracefulShutdown) RegisterMinIOClose(closeFn func() error) {
	g.RegisterCleanup(func(ctx context.Context) error {
		g.logger.Info("closing MinIO client")
		if err := closeFn(); err != nil {
			g.logger.Error("failed to close MinIO", zap.Error(err))
			return fmt.Errorf("close MinIO: %w", err)
		}
		g.logger.Info("MinIO client closed")
		return nil
	})
}

// RegisterDatabaseClose registers database pool for cleanup.
func (g *GracefulShutdown) RegisterDatabaseClose(pool *pgxpool.Pool) {
	g.RegisterCleanup(func(ctx context.Context) error {
		g.logger.Info("closing database pool")
		pool.Close()
		g.logger.Info("database pool closed")
		return nil
	})
}

// WaitForShutdown blocks until a shutdown signal (SIGTERM or SIGINT) is received,
// then gracefully shuts down the gRPC server and calls all registered cleanup functions.
func (g *GracefulShutdown) WaitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigChan
	g.logger.Info("shutdown signal received", zap.String("signal", sig.String()))

	// Create context with drain timeout
	ctx, cancel := context.WithTimeout(context.Background(), g.drainPeriod)
	defer cancel()

	g.logger.Info("initiating graceful shutdown", zap.Duration("drain_period", g.drainPeriod))

	// Stop accepting new connections and wait for in-flight requests
	g.logger.Info("stopping gRPC server")
	g.grpcServer.GracefulStop()
	g.logger.Info("gRPC server stopped")

	// Call cleanup functions in reverse order (LIFO)
	for i := len(g.cleanupFns) - 1; i >= 0; i-- {
		if err := g.cleanupFns[i](ctx); err != nil {
			g.logger.Error("cleanup error", zap.Error(err), zap.Int("index", i))
		}
	}

	g.logger.Info("graceful shutdown complete")
}

// DefaultGRPCServerOptions returns production-ready gRPC server options.
func DefaultGRPCServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxConcurrentStreams(1000),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    10 * time.Second,
			Timeout: 1 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}
}
