package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http/httptest"
	"testing"

	"github.com/Briyantama/stery-kit/go/objstore"
	"github.com/alicebob/miniredis/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/Briyantama/HRIS-Stery/internal/config"
)

// TestRoutes builds the real wiring (the pool never connects) and checks that the routes register
// without a pattern conflict, liveness is public and each service's routes require a token.
func TestRoutes(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	cfg := config.Config{
		AuthPrivateKeyPEM: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}),
		AuthPublicKeyPEM:  pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}),
		AccessTokenHours:  1,
		RefreshTokenDays:  1,
	}
	pool, err := pgxpool.New(context.Background(), "postgres://nobody@127.0.0.1:1/none")
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	rdb := redis.NewClient(&redis.Options{Addr: miniredis.RunT(t).Addr()})

	handlers, consumers, err := build(cfg, pool, rdb, objstore.NewMemory())
	if err != nil {
		t.Fatal(err)
	}
	if len(consumers) != 4 {
		t.Errorf("consumer groups = %d, want 4", len(consumers))
	}
	mux := newMux(handlers...)

	rows := []struct {
		method, path string
		want         int
	}{
		{"GET", "/healthz", 200},
		{"GET", "/api/v1/auth/permissions/u-1", 401},
		{"GET", "/api/v1/employees", 401},
		{"GET", "/api/v1/leaves", 401},
		{"GET", "/api/v1/attendance", 401},
		{"GET", "/api/v1/audit/entries", 401},
		{"GET", "/api/v1/documents", 401},
		{"GET", "/api/v1/notifications", 401},
	}
	for _, r := range rows {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(r.method, r.path, nil))
		if rec.Code != r.want {
			t.Errorf("%s %s: got %d want %d", r.method, r.path, rec.Code, r.want)
		}
	}
}
