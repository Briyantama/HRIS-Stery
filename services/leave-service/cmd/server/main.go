package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/hris-stery/hris-stery/gen/go/hris/leave/v1"
	"github.com/hris-stery/hris-stery/services/_shared/database"
	"github.com/hris-stery/hris-stery/services/_shared/observability"
	"github.com/hris-stery/hris-stery/services/_shared/server"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/application/queries"
	healthsvc "github.com/hris-stery/hris-stery/services/leave-service/internal/health"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/infrastructure/nats"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/infrastructure/postgres"
	grpchandlers "github.com/hris-stery/hris-stery/services/leave-service/internal/interfaces/grpc"
	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	grpcsrv "google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	ctx := context.Background()

	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	// Start metrics server
	observability.StartMetricsServer(logger)

	dbURL := envOr("DATABASE_URL", "postgres://hris_app:hris_app_secret@localhost:6432/hris_db?sslmode=disable")
	natsURL := envOr("NATS_URL", "nats://localhost:4222")
	grpcPort := envOr("GRPC_PORT", "50054")

	// Setup PostgreSQL connection pool
	pool, err := database.SetupPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("setup database pool: %v", err)
	}
	defer pool.Close()
	logger.Info("connected to PostgreSQL", zap.Int32("maxConns", 25))

	// Setup NATS connection and JetStream
	natsConn, err := natsgo.Connect(natsURL)
	if err != nil {
		log.Fatalf("connect to NATS: %v", err)
	}
	defer natsConn.Close()
	logger.Info("connected to NATS")

	js, err := jetstream.New(natsConn)
	if err != nil {
		log.Fatalf("initialize JetStream: %v", err)
	}

	// Create repositories
	leaveRequestRepo := postgres.NewLeaveRequestRepository(pool)
	leaveTypeRepo := postgres.NewLeaveTypeRepository(pool)
	leaveBalanceRepo := postgres.NewLeaveBalanceRepository(pool)

	// Create event publisher
	eventPub := nats.NewEventPublisher(js)

	// Create NATS consumer for employee created events
	employeeConsumer := nats.NewEmployeeCreatedConsumer(js, pool, leaveBalanceRepo, leaveTypeRepo, logger)
	if err := employeeConsumer.Subscribe(ctx); err != nil {
		log.Fatalf("subscribe to employee events: %v", err)
	}
	defer employeeConsumer.Close()

	// Create command handlers
	createLeaveRequestHandler := commands.NewCreateLeaveRequestHandler(
		leaveRequestRepo,
		leaveBalanceRepo,
		eventPub,
		&noopEmployeeValidator{},
	)
	approveLeaveRequestHandler := commands.NewApproveLeaveRequestHandler(leaveRequestRepo, leaveBalanceRepo, eventPub)
	rejectLeaveRequestHandler := commands.NewRejectLeaveRequestHandler(leaveRequestRepo, leaveBalanceRepo, eventPub)
	cancelLeaveRequestHandler := commands.NewCancelLeaveRequestHandler(leaveRequestRepo, leaveBalanceRepo, eventPub)

	// Create query handlers
	getLeaveRequestHandler := queries.NewGetLeaveRequestHandler(leaveRequestRepo)
	listLeaveRequestsHandler := queries.NewListLeaveRequestsHandler(leaveRequestRepo)
	getLeaveBalanceHandler := queries.NewGetLeaveBalanceHandler(leaveBalanceRepo)
	listLeaveTypesHandler := queries.NewListLeaveTypesHandler(leaveTypeRepo)

	// Create gRPC server
	grpcServer := grpcsrv.NewServer(server.ServerOptionsWithLogging(logger)...)
	leaveService := grpchandlers.NewLeaveServiceServer(
		logger,
		createLeaveRequestHandler,
		approveLeaveRequestHandler,
		rejectLeaveRequestHandler,
		cancelLeaveRequestHandler,
		getLeaveRequestHandler,
		listLeaveRequestsHandler,
		getLeaveBalanceHandler,
		listLeaveTypesHandler,
	)
	pb.RegisterLeaveServiceServer(grpcServer, leaveService)

	// Register health check service
	healthService := healthsvc.NewHealthService(logger)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)
	logger.Info("gRPC Health service registered")

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
	shutdown.RegisterNATSDrain(natsConn)
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

// noopEmployeeValidator is a temporary no-op implementation.
// In production, this should call employee-service to validate.
type noopEmployeeValidator struct{}

func (v *noopEmployeeValidator) ValidateEmployeeExists(ctx context.Context, tenantID, employeeID string) (bool, error) {
	return true, nil
}
