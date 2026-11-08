package auth_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/auth/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/auth/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
	"github.com/Briyantama/HRIS-Stery/services/auth/infrastructure"
	"github.com/Briyantama/HRIS-Stery/services/auth/infrastructure/postgres"
	"github.com/alicebob/miniredis/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/Briyantama/HRIS-Stery/internal/testdb"
)

func TestIntegrationAuthFlow(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.NewMigratedURL(t))
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	t.Cleanup(pool.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: miniredis.RunT(t).Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

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

func (n *noopPublisher) PublishSync(context.Context, domain.DomainEvent) error { return nil }

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
