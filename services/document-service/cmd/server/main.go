package main

import (
	"context"
	"log"
	"net"
	"os"

	pb "github.com/hris-stery/hris-stery/gen/go/hris/document/v1"
	"github.com/hris-stery/hris-stery/services/document-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/document-service/internal/application/queries"
	"github.com/hris-stery/hris-stery/services/document-service/internal/infrastructure/postgres"
	"github.com/hris-stery/hris-stery/services/document-service/internal/infrastructure/storage"
	grpchandlers "github.com/hris-stery/hris-stery/services/document-service/internal/interfaces/grpc"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Read configuration from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://hris_app:hris_app_secret@localhost:5432/hris_db?sslmode=disable"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50056"
	}

	// Connect to PostgreSQL
	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		logger.Fatal("failed to parse database URL", zap.Error(err))
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// Verify connection
	if err = pool.Ping(ctx); err != nil {
		logger.Fatal("failed to ping database", zap.Error(err))
	}

	logger.Info("connected to PostgreSQL")

	// Initialize MinIO
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000"
	}

	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		minioAccessKey = "minioadmin"
	}

	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		minioSecretKey = "minioadmin"
	}

	minioUseSSL := os.Getenv("MINIO_USE_SSL") == "true"

	minioClient, err := storage.NewMinIOClient(minioEndpoint, minioAccessKey, minioSecretKey, minioUseSSL, logger)
	if err != nil {
		logger.Fatal("failed to initialize MinIO client", zap.Error(err))
	}

	logger.Info("connected to MinIO", zap.String("endpoint", minioEndpoint))

	// Initialize repositories
	docRepo := postgres.NewDocumentRepository(pool)
	versionRepo := postgres.NewDocumentVersionRepository(pool)

	// Initialize storage service
	storageService := storage.NewDocumentStorageService(minioClient, logger)

	// Initialize command handlers
	uploadHandler := commands.NewUploadDocumentHandler(docRepo, storageService, logger)
	deleteHandler := commands.NewDeleteDocumentHandler(docRepo, logger)
	versionHandler := commands.NewCreateDocumentVersionHandler(docRepo, versionRepo, logger)

	// Initialize query handlers
	listHandler := queries.NewListDocumentsHandler(docRepo, logger)
	metadataHandler := queries.NewGetDocumentMetadataHandler(docRepo, versionRepo, logger)

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register Document Service
	documentServer := grpchandlers.NewDocumentServiceServer(
		uploadHandler,
		deleteHandler,
		versionHandler,
		listHandler,
		metadataHandler,
		logger,
	)
	pb.RegisterDocumentServiceServer(grpcServer, documentServer)

	// Register Health Check Service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("hris.document.v1.DocumentService", grpc_health_v1.HealthCheckResponse_SERVING)

	logger.Info("document-service initialized",
		zap.String("grpc_port", grpcPort),
	)

	// Start listening
	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Fatal("failed to listen on port", zap.String("port", grpcPort), zap.Error(err))
	}

	logger.Info("starting gRPC server", zap.String("port", grpcPort))

	if err = grpcServer.Serve(listener); err != nil {
		logger.Fatal("gRPC server error", zap.Error(err))
	}
}
