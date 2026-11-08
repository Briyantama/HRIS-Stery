package infrastructure

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
	"github.com/Briyantama/stery-kit/go/authn"
)

// The IDs must survive the token as plain strings: a refresh mints the next access token from them.
func TestTokensCarryIDsAsStrings(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	s, err := NewJWTSigner(
		pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}),
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}),
		"test", time.Hour, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	uid := domain.NewUserIDOrNil("2ec82033-fc6a-4cf2-818f-c785b85dba5a")
	tid := domain.NewTenantIDOrNil("a3516a09-e836-47ef-82da-1dad40fdb314")

	refresh, err := s.GenerateRefreshToken(uid, tid, "a@b.c", []string{"hr_admin"}, "h")
	if err != nil {
		t.Fatal(err)
	}
	rc, err := s.ValidateRefreshToken(refresh)
	if err != nil {
		t.Fatal(err)
	}
	if !rc.UserID.Equals(uid) || !rc.TenantID.Equals(tid) {
		t.Fatalf("refresh claims lost the IDs: %q %q", rc.UserID, rc.TenantID)
	}

	access, _, err := s.GenerateAccessToken(rc.UserID, rc.TenantID, rc.Email, rc.Roles)
	if err != nil {
		t.Fatal(err)
	}
	ac, err := s.ValidateAccessToken(access)
	if err != nil {
		t.Fatal(err)
	}
	if !ac.UserID.Equals(uid) || !ac.TenantID.Equals(tid) {
		t.Fatalf("access claims lost the IDs: %q %q", ac.UserID, ac.TenantID)
	}
	if strings.Count(access, ".") != 2 {
		t.Fatalf("not a JWT: %s", access)
	}
}

// A refresh token must not authenticate API routes, and an access token must not refresh.
func TestRefreshTokenIsNotAnAccessToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	s, err := NewJWTSigner(
		pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}),
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}),
		"test", time.Hour, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	uid := domain.NewUserIDOrNil("2ec82033-fc6a-4cf2-818f-c785b85dba5a")
	tid := domain.NewTenantIDOrNil("a3516a09-e836-47ef-82da-1dad40fdb314")
	refresh, err := s.GenerateRefreshToken(uid, tid, "a@b.c", []string{"hr_admin"}, "h")
	if err != nil {
		t.Fatal(err)
	}
	access, _, err := s.GenerateAccessToken(uid, tid, "a@b.c", []string{"hr_admin"})
	if err != nil {
		t.Fatal(err)
	}
	v := authn.NewRSAVerifier(&key.PublicKey, nil)
	if _, err := v.Verify(context.Background(), refresh); err == nil {
		t.Error("refresh token passed the kit verifier")
	}
	if _, err := v.Verify(context.Background(), access); err != nil {
		t.Errorf("access token rejected by kit verifier: %v", err)
	}
	if _, err := s.ValidateRefreshToken(access); err == nil {
		t.Error("access token accepted as refresh token")
	}
}
