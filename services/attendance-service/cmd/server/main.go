package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	attendancev1 "github.com/hris-stery/hris-stery/gen/go/hris/attendance/v1"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/application/queries"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/infrastructure/nats"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/infrastructure/postgres"
	grpchandlers "github.com/hris-stery/hris-stery/services/attendance-service/internal/interfaces/grpc"
	"github.com/jackc/pgx/v5/pgxpool"
	natslib "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	dbURL := envOr("DATABASE_URL", "postgres://hris_app:hris_app_secret@localhost:6432/hris_db?sslmode=disable")
	natsURL := envOr("NATS_URL", "nats://localhost:4222")
	grpcPort := envOr("GRPC_PORT", "50053")

	// Connect to PostgreSQL
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
	grpcServer := grpc.NewServer()
	attendanceService := grpchandlers.NewAttendanceServiceServer(
		checkInHandler,
		checkOutHandler,
		overrideHandler,
		getAttendanceHandler,
		listAttendanceHandler,
		getDailySummaryHandler,
	)
	attendancev1.RegisterAttendanceServiceServer(grpcServer, attendanceService)

	// Start listening on gRPC port
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
