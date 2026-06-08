package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	employeev1 "github.com/hris-stery/hris-stery/gen/go/hris/employee/v1"
	"github.com/hris-stery/hris-stery/services/_shared/database"
	"github.com/hris-stery/hris-stery/services/_shared/observability"
	"github.com/hris-stery/hris-stery/services/_shared/server"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/application/queries"
	healthsvc "github.com/hris-stery/hris-stery/services/employee-service/internal/health"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/infrastructure"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/infrastructure/cache"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/infrastructure/postgres"
	grpchandlers "github.com/hris-stery/hris-stery/services/employee-service/internal/interfaces/grpc"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
	grpcPort := envOr("GRPC_PORT", "50052")

	pool, err := database.SetupPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("setup database pool: %v", err)
	}
	defer pool.Close()
	logger.Info("connected to PostgreSQL", zap.Int32("maxConns", 25))

	natsConn, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("connect to NATS: %v", err)
	}
	defer natsConn.Close()
	logger.Info("connected to NATS")

	// Create repositories with caching for read-heavy data
	employeeRepo := postgres.NewEmployeeRepository(pool)
	departmentRepo := cache.NewDepartmentCacheRepository(postgres.NewDepartmentRepository(pool))
	positionRepo := cache.NewPositionCacheRepository(postgres.NewPositionRepository(pool))

	// Create event publisher
	eventPub, err := infrastructure.NewNATSPublisher(natsConn)
	if err != nil {
		log.Fatalf("initialize NATS publisher: %v", err)
	}

	// Create command handlers
	createEmployeeHandler := commands.NewCreateEmployeeHandler(employeeRepo, departmentRepo, positionRepo, eventPub)
	terminateEmployeeHandler := commands.NewTerminateEmployeeHandler(employeeRepo, eventPub)
	createDepartmentHandler := commands.NewCreateDepartmentHandler(departmentRepo)
	createPositionHandler := commands.NewCreatePositionHandler(positionRepo)

	// Create query handlers
	getEmployeeHandler := queries.NewGetEmployeeHandler(employeeRepo)
	listEmployeesHandler := queries.NewListEmployeesHandler(employeeRepo)
	listDepartmentsHandler := queries.NewListDepartmentsHandler(departmentRepo)
	listPositionsHandler := queries.NewListPositionsHandler(positionRepo)

	// Create event consumer for user registration
	consumer := infrastructure.NewUserRegisteredConsumer(pool, createEmployeeHandler, logger)
	if err := consumer.Subscribe(natsConn); err != nil {
		log.Fatalf("subscribe to events: %v", err)
	}
	defer consumer.Close()

	// Create gRPC server
	grpcServer := grpc.NewServer(server.ServerOptionsWithLogging(logger)...)
	employeeService := grpchandlers.NewEmployeeServiceServer(
		createEmployeeHandler,
		terminateEmployeeHandler,
		createDepartmentHandler,
		createPositionHandler,
		getEmployeeHandler,
		listEmployeesHandler,
		listDepartmentsHandler,
		listPositionsHandler,
	)
	employeev1.RegisterEmployeeServiceServer(grpcServer, employeeService)

	// Register health check service
	healthService := healthsvc.NewHealthService(pool, natsConn, logger)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)
	logger.Info("gRPC Health service registered")

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
