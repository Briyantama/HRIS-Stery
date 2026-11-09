package infrastructure

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestTokenService(t *testing.T) (*TokenServiceImpl, *miniredis.Miniredis) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	signer, err := NewJWTSigner(
		pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}),
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}),
		"test", time.Hour, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	mr := miniredis.RunT(t)
	rc := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rc.Close() })
	return NewTokenService(signer, rc, time.Hour), mr
}

func newRefreshToken(t *testing.T, s *TokenServiceImpl) string {
	t.Helper()
	_, refresh, _, err := s.GenerateTokenPair(context.Background(),
		domain.NewUserIDOrNil("2ec82033-fc6a-4cf2-818f-c785b85dba5a"),
		domain.NewTenantIDOrNil("a3516a09-e836-47ef-82da-1dad40fdb314"), "a@b.c", []string{"hr_admin"})
	if err != nil {
		t.Fatal(err)
	}
	return refresh
}

func TestConsumeRefreshTokenRejectsReplay(t *testing.T) {
	s, _ := newTestTokenService(t)
	refresh := newRefreshToken(t, s)

	if _, err := s.ConsumeRefreshToken(context.Background(), refresh); err != nil {
		t.Fatalf("first consume: %v", err)
	}
	if _, err := s.ConsumeRefreshToken(context.Background(), refresh); !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("replay must be unauthenticated, got %v", err)
	}
}

func TestConsumeRefreshTokenConcurrentReplayWinsOnce(t *testing.T) {
	s, _ := newTestTokenService(t)
	refresh := newRefreshToken(t, s)

	const n = 20
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	for range n {
		wg.Go(func() {
			if _, err := s.ConsumeRefreshToken(context.Background(), refresh); err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("exactly one concurrent consume may succeed, got %d", wins)
	}
}

func TestConsumeRefreshTokenRejectsRevoked(t *testing.T) {
	s, _ := newTestTokenService(t)
	refresh := newRefreshToken(t, s)
	if err := s.RevokeRefreshToken(context.Background(), refresh); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ConsumeRefreshToken(context.Background(), refresh); !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("revoked token must be unauthenticated, got %v", err)
	}
}

// A broken revocation store must not read as "not revoked".
func TestConsumeRefreshTokenFailsClosedWhenStoreErrors(t *testing.T) {
	s, mr := newTestTokenService(t)
	refresh := newRefreshToken(t, s)

	mr.SetError("redis down")
	_, err := s.ConsumeRefreshToken(context.Background(), refresh)
	if err == nil {
		t.Fatal("expected error when the revocation store fails")
	}
	if errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("store failure is an internal error, not a bad credential: %v", err)
	}

	mr.SetError("")
	if _, err := s.ConsumeRefreshToken(context.Background(), refresh); err != nil {
		t.Fatalf("token must stay usable after a failed attempt: %v", err)
	}
}
