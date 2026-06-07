package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	attendancev1 "github.com/hris-stery/hris-stery/gen/go/hris/attendance/v1"
	"github.com/hris-stery/hris-stery/services/_shared/database"
	"github.com/hris-stery/hris-stery/services/_shared/server"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/application/queries"
	healthsvc "github.com/hris-stery/hris-stery/services/attendance-service/internal/health"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/infrastructure/nats"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/infrastructure/postgres"
	grpchandlers "github.com/hris-stery/hris-stery/services/attendance-service/internal/interfaces/grpc"
	natslib "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	ctx := context.Background()

	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	dbURL := envOr("DATABASE_URL", "postgres://hris_app:hris_app_secret@localhost:6432/hris_db?sslmode=disable")
	natsURL := envOr("NATS_URL", "nats://localhost:4222")
	grpcPort := envOr("GRPC_PORT", "50053")

	// Connect to PostgreSQL
	pool, err := database.SetupPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("setup database pool: %v", err)
	}
	defer pool.Close()
	logger.Info("connected to PostgreSQL", zap.Int32("maxConns", 25))

	// Connect to NATS
	natsConn, err := natslib.Connect(natsURL)
	if err != nil {
		log.Fatalf("connect to NATS: %v", err)
	}
	defer natsConn.Close()
	logger.Info("connected to NATS")

	// Create JetStream context
	js, err := jetstream.New(natsConn)
	if err != nil {
		log.Fatalf("create JetStream context: %v", err)
	}

	// Create repositories
	attendanceRepo := postgres.NewAttendanceRepository(pool)

	// Create event publisher
	eventPub := nats.NewEventPublisher(js)

	// Create command handlers
	checkInHandler := commands.NewCheckInHandler(attendanceRepo, eventPub)
	checkOutHandler := commands.NewCheckOutHandler(attendanceRepo, eventPub)
	overrideHandler := commands.NewOverrideAttendanceHandler(attendanceRepo, eventPub)

	// Create query handlers
	getAttendanceHandler := queries.NewGetAttendanceHandler(attendanceRepo)
	listAttendanceHandler := queries.NewListAttendanceHandler(attendanceRepo)
	getDailySummaryHandler := queries.NewGetDailySummaryHandler(attendanceRepo)

	// Create gRPC server
	grpcServer := grpc.NewServer(server.DefaultGRPCServerOptions()...)
	attendanceService := grpchandlers.NewAttendanceServiceServer(
		checkInHandler,
		checkOutHandler,
		overrideHandler,
		getAttendanceHandler,
		listAttendanceHandler,
		getDailySummaryHandler,
	)
	attendancev1.RegisterAttendanceServiceServer(grpcServer, attendanceService)

	// Register health check service
	healthService := healthsvc.NewHealthService(pool, natsConn, logger)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)
	logger.Info("gRPC Health service registered")

	// Start listening on gRPC port
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
