package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	auditv1 "github.com/hris-stery/hris-stery/gen/go/hris/audit/v1"
	"github.com/hris-stery/hris-stery/services/_shared/database"
	"github.com/hris-stery/hris-stery/services/_shared/observability"
	"github.com/hris-stery/hris-stery/services/_shared/server"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/application/queries"
	healthsvc "github.com/hris-stery/hris-stery/services/audit-service/internal/health"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/infrastructure/nats"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/infrastructure/postgres"
	grpchandlers "github.com/hris-stery/hris-stery/services/audit-service/internal/interfaces/grpc"
	natsconn "github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	ctx := context.Background()

	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	// Start metrics server
	observability.StartMetricsServer(logger)

	// Read configuration from environment
	dbURL := envOr("DATABASE_URL", "postgres://hris_app:hris_app_secret@localhost:6432/hris_db?sslmode=disable")
	natsURL := envOr("NATS_URL", "nats://localhost:4222")
	grpcPort := envOr("GRPC_PORT", "50054")

	// Connect to PostgreSQL
	pool, err := database.SetupPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("setup database pool: %v", err)
	}
	defer pool.Close()
	logger.Info("connected to PostgreSQL", zap.Int32("maxConns", 25))

	// Connect to NATS
	nc, err := natsconn.Connect(natsURL)
	if err != nil {
		log.Fatalf("connect to NATS: %v", err)
	}
	defer nc.Close()
	logger.Info("connected to NATS")

	// Get JetStream context
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("get JetStream context: %v", err)
	}

	// Ensure HRIS_EVENTS stream exists
	_, err = js.AddStream(&natsconn.StreamConfig{
		Name:     "HRIS_EVENTS",
		Subjects: []string{"hris.>"},
	})
	if err != nil && err != natsconn.ErrStreamNameAlreadyInUse {
		log.Fatalf("create HRIS_EVENTS stream: %v", err)
	}
	logger.Info("NATS JetStream ready")

	// Initialize repositories
	auditRepo := postgres.NewAuditRepository(pool)

	// Initialize handlers
	recordHandler := commands.NewRecordAuditHandler(auditRepo)
	queryHandler := queries.NewQueryAuditTrailHandler(auditRepo)

	// Initialize NATS consumers
	identityConsumer := nats.NewIdentityConsumer(js, pool, recordHandler, logger)
	workforceConsumer := nats.NewWorkforceConsumer(js, pool, recordHandler, logger)
	operationsConsumer := nats.NewOperationsConsumer(js, pool, recordHandler, logger)
	notificationConsumer := nats.NewNotificationConsumer(js, pool, recordHandler, logger)

	// Subscribe to NATS subjects
	if err := identityConsumer.Subscribe(ctx); err != nil {
		log.Fatalf("subscribe to identity events: %v", err)
	}
	logger.Info("subscribed to identity events")

	if err := workforceConsumer.Subscribe(ctx); err != nil {
		log.Fatalf("subscribe to workforce events: %v", err)
	}
	logger.Info("subscribed to workforce events")

	if err := operationsConsumer.Subscribe(ctx); err != nil {
		log.Fatalf("subscribe to operations events: %v", err)
	}
	logger.Info("subscribed to operations events")

	if err := notificationConsumer.Subscribe(ctx); err != nil {
		log.Fatalf("subscribe to notification events: %v", err)
	}
	logger.Info("subscribed to notification events")

	// Create gRPC server with options
	grpcServer := grpc.NewServer(server.ServerOptionsWithLogging(logger)...)

	// Register audit service
	auditService := grpchandlers.NewAuditServiceServer(recordHandler, queryHandler, logger)
	auditv1.RegisterAuditServiceServer(grpcServer, auditService)
	logger.Info("registered audit service")

	// Register health service
	healthImpl := healthsvc.NewHealthService(pool, nc, logger)
	healthgrpc.RegisterHealthServer(grpcServer, healthImpl)
	logger.Info("registered health service")

	// Start listening
	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("listen on port %s: %v", grpcPort, err)
	}

	logger.Info(fmt.Sprintf("starting gRPC server on port %s", grpcPort))

	// Start gRPC server in goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	// Setup graceful shutdown
	shutdown := server.NewGracefulShutdown(grpcServer, logger)
	shutdown.RegisterNATSDrain(nc)
	shutdown.RegisterDatabaseClose(pool)

	// Wait for shutdown signal
	shutdown.WaitForShutdown()
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
