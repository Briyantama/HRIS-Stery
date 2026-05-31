//go:build integration

package integration_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"github.com/hris-stery/hris-stery/services/auth-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/application/queries"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/infrastructure"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/infrastructure/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func TestIntegrationAuthFlow(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("set INTEGRATION=true to run")
	}

	ctx := context.Background()
	dbURL := env("DATABASE_URL", "postgres://hris_app:hris_app_secret@localhost:5432/hris_db?sslmode=disable")
	redisURL := env("REDIS_URL", "redis://localhost:6379/0")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	t.Cleanup(pool.Close)

	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("parse redis URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)
	t.Cleanup(func() { _ = redisClient.Close() })
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatalf("connect redis: %v", err)
	}

	privPEM, pubPEM := testRSAKeys(t)
	signer, err := infrastructure.NewJWTSigner(privPEM, pubPEM, "auth-service", 15*time.Minute, time.Hour)
	if err != nil {
		t.Fatalf("jwt signer: %v", err)
	}
	tokenSvc := infrastructure.NewTokenService(signer, redisClient, time.Hour)
	eventPub := &noopPublisher{}

	tenantRepo := postgres.NewTenantRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	roleRepo := postgres.NewRoleRepository(pool)
	permissionRepo := postgres.NewPermissionRepository(pool)

	register := commands.NewRegisterTenantHandler(tenantRepo, userRepo, roleRepo, permissionRepo, tokenSvc, eventPub)
	login := commands.NewLoginHandler(tenantRepo, userRepo, tokenSvc, eventPub)
	refresh := commands.NewRefreshTokenHandler(tokenSvc)
	revoke := commands.NewRevokeTokenHandler(tokenSvc)
	validate := queries.NewValidateTokenHandler(tokenSvc)

	slug := "integration-" + randomSuffix()
	_, err = register.Handle(ctx, commands.RegisterTenantCommand{
		CompanyName:   "Integration Co",
		TenantSlug:    slug,
		AdminEmail:    "admin@" + slug + ".test",
		AdminPassword: "ValidPass123!",
		AdminName:     "Admin User",
	})
	if err != nil {
		t.Fatalf("register tenant: %v", err)
	}

	loginResult, err := login.Handle(ctx, commands.LoginCommand{
		Email:      "admin@" + slug + ".test",
		Password:   "ValidPass123!",
		TenantSlug: slug,
		IPAddr:     "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	refreshResult, err := refresh.Handle(ctx, commands.RefreshTokenCommand{RefreshToken: loginResult.RefreshToken})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshResult.RefreshToken == loginResult.RefreshToken {
		t.Fatal("expected rotated refresh token")
	}

	claims, err := validate.Handle(ctx, queries.ValidateTokenQuery{AccessToken: refreshResult.AccessToken})
	if err != nil {
		t.Fatalf("validate refreshed access token: %v", err)
	}
	if claims.Email == "" {
		t.Fatal("expected email in access token claims")
	}

	if err := revoke.Handle(ctx, commands.RevokeTokenCommand{RefreshToken: refreshResult.RefreshToken}); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	_, err = refresh.Handle(ctx, commands.RefreshTokenCommand{RefreshToken: refreshResult.RefreshToken})
	if err == nil {
		t.Fatal("expected refresh to fail after revoke")
	}

	_, err = login.Handle(ctx, commands.LoginCommand{
		Email:      "admin@" + slug + ".test",
		Password:   "WrongPassword123!",
		TenantSlug: slug,
		IPAddr:     "127.0.0.1",
	})
	if err == nil {
		t.Fatal("expected login failure for invalid password")
	}

	_, err = login.Handle(ctx, commands.LoginCommand{
		Email:      "admin@" + slug + ".test",
		Password:   "ValidPass123!",
		TenantSlug: "nonexistent-tenant",
		IPAddr:     "127.0.0.1",
	})
	if err == nil {
		t.Fatal("expected login failure for invalid tenant")
	}
}

type noopPublisher struct{}

func (n *noopPublisher) PublishAsync(context.Context, domain.DomainEvent) error { return nil }
func (n *noopPublisher) PublishSync(context.Context, domain.DomainEvent) error  { return nil }

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func randomSuffix() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmtHex(b)
}

func fmtHex(b []byte) string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = hexdigits[v>>4]
		out[i*2+1] = hexdigits[v&0x0f]
	}
	return string(out)
}

func testRSAKeys(t *testing.T) ([]byte, []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	privDER := x509.MarshalPKCS1PrivateKey(key)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDER})
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return privPEM, pubPEM
}
