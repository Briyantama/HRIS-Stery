package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	authv1 "github.com/hris-stery/hris-stery/gen/go/hris/auth/v1"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/application/queries"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/infrastructure"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/infrastructure/config"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/infrastructure/postgres"
	grpchandlers "github.com/hris-stery/hris-stery/services/auth-service/internal/interfaces/grpc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	logger, _ := zap.NewProduction()
	defer func() { _ = logger.Sync() }()

	dbURL := envOr("DATABASE_URL", "postgres://hris_app:hris_app_secret@localhost:6432/hris_db?sslmode=disable")
	redisURL := envOr("REDIS_URL", "redis://:hris_redis_secret@localhost:6379/0")
	natsURL := envOr("NATS_URL", "nats://localhost:4222")
	grpcPort := envOr("GRPC_PORT", "50051")

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

	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("parse redis URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("connect to Redis: %v", err)
	}
	logger.Info("connected to Redis")

	natsConn, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("connect to NATS: %v", err)
	}
	defer natsConn.Close()
	logger.Info("connected to NATS")

	tenantRepo := postgres.NewTenantRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	roleRepo := postgres.NewRoleRepository(pool)
	permissionRepo := postgres.NewPermissionRepository(pool)

	privateKeyPEM, publicKeyPEM, err := config.LoadPEM(
		"AUTH_PRIVATE_KEY", "AUTH_PUBLIC_KEY",
		"AUTH_PRIVATE_KEY_BASE64", "AUTH_PUBLIC_KEY_BASE64",
	)
	if err != nil {
		log.Fatalf("load RSA keys: %v", err)
	}

	jwtSigner, err := infrastructure.NewJWTSigner(privateKeyPEM, publicKeyPEM, "auth-service", 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		log.Fatalf("initialize JWT signer: %v", err)
	}

	tokenSvc := infrastructure.NewTokenService(jwtSigner, redisClient, 7*24*time.Hour)
	eventPub, err := infrastructure.NewNATSPublisher(natsConn)
	if err != nil {
		log.Fatalf("initialize NATS publisher: %v", err)
	}

	loginHandler := commands.NewLoginHandler(tenantRepo, userRepo, tokenSvc, eventPub)
	refreshTokenHandler := commands.NewRefreshTokenHandler(tokenSvc)
	revokeTokenHandler := commands.NewRevokeTokenHandler(tokenSvc)
	registerTenantHandler := commands.NewRegisterTenantHandler(tenantRepo, userRepo, roleRepo, permissionRepo, tokenSvc, eventPub)
	validateTokenHandler := queries.NewValidateTokenHandler(tokenSvc)
	getPermissionsHandler := queries.NewGetPermissionsHandler(userRepo, roleRepo, permissionRepo)

	grpcServer := grpc.NewServer()
	authService := grpchandlers.NewAuthServiceServer(
		loginHandler,
		refreshTokenHandler,
		revokeTokenHandler,
		registerTenantHandler,
		validateTokenHandler,
		getPermissionsHandler,
	)
	authv1.RegisterAuthServiceServer(grpcServer, authService)

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
