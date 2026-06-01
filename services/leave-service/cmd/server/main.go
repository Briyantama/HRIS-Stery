package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/hris-stery/hris-stery/services/leave-service/internal/interfaces/grpc"
	pb "github.com/hris-stery/hris-stery/gen/go/hris/leave/v1"
	"go.uber.org/zap"
	grpcsrv "google.golang.org/grpc"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	grpcPort := 50054
	if port, err := strconv.Atoi(os.Getenv("GRPC_PORT")); err == nil {
		grpcPort = port
	}

	logger.Info("starting leave-service", zap.Int("grpc_port", grpcPort))

	// Create gRPC server
	srv := grpcsrv.NewServer()
	leaveService := grpc.NewLeaveServiceServer(logger)
	pb.RegisterLeaveServiceServer(srv, leaveService)

	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		logger.Fatal("listening on port", zap.Error(err))
	}

	logger.Info("serving gRPC on port", zap.Int("port", grpcPort))
	if err := srv.Serve(listener); err != nil {
		logger.Fatal("serving gRPC", zap.Error(err))
	}
}
