package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/hris-stery/hris-stery/services/_shared/database"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/hris-stery/hris-stery/gen/go/hris/notification/v1"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/application/consumers"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/application/queries"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/health"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/infrastructure/channels"
	infracons "github.com/hris-stery/hris-stery/services/notification-service/internal/infrastructure/nats"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/infrastructure/postgres"
	grpchandlers "github.com/hris-stery/hris-stery/services/notification-service/internal/interfaces/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// Setup logging
	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Fatal("DATABASE_URL not set")
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	grpcPort := 50055
	if portStr := os.Getenv("GRPC_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			grpcPort = port
		}
	}

	// Parse SMTP configuration
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpPort := 587 // Default to TLS port
	if smtpPortStr != "" {
		if port, err := strconv.Atoi(smtpPortStr); err == nil {
			smtpPort = port
		}
	}
	smtpUser := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFromEmail := os.Getenv("SMTP_FROM_EMAIL")
	if smtpFromEmail == "" {
		smtpFromEmail = "noreply@hris.local"
	}
	smtpFromName := os.Getenv("SMTP_FROM_NAME")
	if smtpFromName == "" {
		smtpFromName = "HRIS System"
	}
	smtpTLSEnabled := os.Getenv("SMTP_TLS_ENABLED") != "false" // Default true
	logger.Info("SMTP configuration loaded",
		zap.String("smtp_host", smtpHost),
		zap.Int("smtp_port", smtpPort),
		zap.String("smtp_from_email", smtpFromEmail),
		zap.Bool("smtp_tls_enabled", smtpTLSEnabled),
	)

	// Connect PostgreSQL
	logger.Info("connecting to PostgreSQL")
	pool, err := database.SetupPool(ctx, databaseURL)
	if err != nil {
		logger.Fatal("failed to setup database pool", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Fatal("failed to ping PostgreSQL", zap.Error(err))
	}
	logger.Info("connected to PostgreSQL")

	// Connect NATS
	logger.Info("connecting to NATS", zap.String("nats_url", natsURL))
	nc, err := nats.Connect(natsURL)
	if err != nil {
		logger.Fatal("failed to connect NATS", zap.Error(err))
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		logger.Fatal("failed to get JetStream context", zap.Error(err))
	}
	logger.Info("connected to NATS JetStream")

	// Instantiate repositories
	notificationRepo := postgres.NewNotificationRepository(pool)
	channelConfigRepo := postgres.NewChannelConfigRepository(pool)

	// Instantiate channel adapters (for future use in sending notifications)
	emailAdapter := channels.NewEmailAdapter(
		smtpHost,
		smtpPort,
		smtpUser,
		smtpPassword,
		smtpFromEmail,
		smtpFromName,
		smtpTLSEnabled,
		logger,
	)
	inAppAdapter := channels.NewInAppAdapter()
	_ = emailAdapter // Placeholder for future use
	_ = inAppAdapter // Placeholder for future use

	// Instantiate command handlers
	sendHandler := commands.NewSendNotificationHandler(notificationRepo)
	markReadHandler := commands.NewMarkNotificationReadHandler(notificationRepo)
	updateConfigHandler := commands.NewUpdateChannelConfigHandler(channelConfigRepo)

	// Instantiate query handlers
	listNotificationsHandler := queries.NewListNotificationsHandler(notificationRepo)
	getDeliveryStatusHandler := queries.NewGetDeliveryStatusHandler(notificationRepo)
	getChannelConfigHandler := queries.NewGetChannelConfigHandler(channelConfigRepo)

	// Instantiate application layer consumers
	appLeaveConsumer := consumers.NewLeaveEventConsumer(sendHandler, channelConfigRepo)
	appEmployeeConsumer := consumers.NewEmployeeEventConsumer(sendHandler, channelConfigRepo)
	appAuthConsumer := consumers.NewAuthEventConsumer(sendHandler, channelConfigRepo)

	// Instantiate NATS event consumers
	leaveConsumer := infracons.NewLeaveConsumer(js, pool, appLeaveConsumer, logger)
	employeeConsumer := infracons.NewEmployeeConsumer(js, pool, appEmployeeConsumer, logger)
	authConsumer := infracons.NewAuthConsumer(js, pool, appAuthConsumer, logger)

	// Subscribe to events
	logger.Info("subscribing to NATS events")
	if err := leaveConsumer.Subscribe(ctx); err != nil {
		logger.Fatal("failed to subscribe to leave events", zap.Error(err))
	}
	if err := employeeConsumer.Subscribe(ctx); err != nil {
		logger.Fatal("failed to subscribe to employee events", zap.Error(err))
	}
	if err := authConsumer.Subscribe(ctx); err != nil {
		logger.Fatal("failed to subscribe to auth events", zap.Error(err))
	}
	logger.Info("subscribed to all NATS events")

	// Create gRPC server
	grpcServer := grpc.NewServer()
	notificationService := grpchandlers.NewNotificationServiceServer(
		logger,
		sendHandler,
		markReadHandler,
		updateConfigHandler,
		listNotificationsHandler,
		getDeliveryStatusHandler,
		getChannelConfigHandler,
	)
	pb.RegisterNotificationServiceServer(grpcServer, notificationService)

	// Register health check service
	healthService := health.NewHealthService(pool, nc, logger)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthService)
	logger.Info("gRPC Health service registered")

	// Start gRPC server
	logger.Info("starting gRPC server", zap.Int("port", grpcPort))
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		logger.Fatal("failed to listen on gRPC port", zap.Error(err))
	}

	logger.Info("notification-service ready", zap.Int("grpc_port", grpcPort))
	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("gRPC server error", zap.Error(err))
	}
}
