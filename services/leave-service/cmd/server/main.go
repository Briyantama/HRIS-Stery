package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	pb "github.com/hris-stery/hris-stery/gen/go/hris/leave/v1"
	"github.com/hris-stery/hris-stery/services/_shared/server"
	healthsvc "github.com/hris-stery/hris-stery/services/leave-service/internal/health"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/interfaces/grpc"
	"go.uber.org/zap"
	grpcsrv "google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	grpcPort := 50054
	if port, err := strconv.Atoi(os.Getenv("GRPC_PORT")); err == nil {
		grpcPort = port
	}

	logger.Info("starting leave-service", zap.Int("grpc_port", grpcPort))

	// Create gRPC server
	srv := grpcsrv.NewServer(server.DefaultGRPCServerOptions()...)
	leaveService := grpc.NewLeaveServiceServer(logger)
	pb.RegisterLeaveServiceServer(srv, leaveService)

	// Register health check service
	healthService := healthsvc.NewHealthService(logger)
	grpc_health_v1.RegisterHealthServer(srv, healthService)
	logger.Info("gRPC Health service registered")

	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		logger.Fatal("listening on port", zap.Error(err))
	}

	logger.Info("serving gRPC on port", zap.Int("port", grpcPort))

	// Start gRPC server in goroutine
	go func() {
		if err := srv.Serve(listener); err != nil {
			logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	// Setup graceful shutdown
	shutdown := server.NewGracefulShutdown(srv, logger)

	// Wait for shutdown signal
	shutdown.WaitForShutdown()
}
