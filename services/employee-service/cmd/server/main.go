package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	employeev1 "github.com/hris-stery/hris-stery/gen/go/hris/employee/v1"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/application/queries"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/infrastructure/postgres"
	grpchandlers "github.com/hris-stery/hris-stery/services/employee-service/internal/interfaces/grpc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dbURL := envOr("DATABASE_URL", "postgres://hris_app:hris_app_secret@localhost:6432/hris_db?sslmode=disable")
	natsURL := envOr("NATS_URL", "nats://localhost:4222")
	grpcPort := envOr("GRPC_PORT", "50052")

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("parse database URL: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()
	logger.Info("connected to PostgreSQL")

	natsConn, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("connect to NATS: %v", err)
	}
	defer natsConn.Close()
	logger.Info("connected to NATS")

	// Create repositories
	employeeRepo := postgres.NewEmployeeRepository(pool)
	departmentRepo := postgres.NewDepartmentRepository(pool)
	positionRepo := postgres.NewPositionRepository(pool)

	// Create event publisher
	eventPub := &noopPublisher{} // TODO: Use NATSPublisher

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

	// Create gRPC server
	grpcServer := grpc.NewServer()
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

	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("listen on port %s: %v", grpcPort, err)
	}

	logger.Info(fmt.Sprintf("starting gRPC server on port %s", grpcPort))
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("serve gRPC: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// noopPublisher is a temporary event publisher that does nothing.
type noopPublisher struct{}

func (n *noopPublisher) PublishAsync(ctx context.Context, event interface{}) error {
	return nil
}

func (n *noopPublisher) PublishSync(ctx context.Context, event interface{}) error {
	return nil
}
