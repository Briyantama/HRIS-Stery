package commands

import (
	"context"
	"fmt"
	"testing"

	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
)

type rotatingTokenMock struct {
	access  string
	refresh string
	ttl     int32
	err     error
}

func (m *rotatingTokenMock) GenerateTokenPair(context.Context, domain.UserID, domain.TenantID, string, []string) (string, string, int32, error) {
	return "", "", 0, fmt.Errorf("not implemented")
}

func (m *rotatingTokenMock) ValidateAccessToken(context.Context, string) (*TokenClaims, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *rotatingTokenMock) RefreshAccessToken(context.Context, string) (string, string, int32, error) {
	return m.access, m.refresh, m.ttl, m.err
}

func (m *rotatingTokenMock) RevokeRefreshToken(context.Context, string) error {
	return fmt.Errorf("not implemented")
}

func (m *rotatingTokenMock) IsRefreshTokenRevoked(context.Context, string) (bool, error) {
	return false, nil
}

func TestRefreshTokenRotatesPair(t *testing.T) {
	mock := &rotatingTokenMock{
		access:  "access-2",
		refresh: "refresh-2",
		ttl:     900,
	}
	handler := NewRefreshTokenHandler(mock)

	result, err := handler.Handle(context.Background(), RefreshTokenCommand{RefreshToken: "refresh-1"})
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if result.AccessToken != "access-2" || result.RefreshToken != "refresh-2" {
		t.Fatalf("unexpected token pair: %+v", result)
	}
}

func TestRefreshTokenRejectsInvalid(t *testing.T) {
	mock := &rotatingTokenMock{err: fmt.Errorf("invalid refresh token")}
	handler := NewRefreshTokenHandler(mock)

	_, err := handler.Handle(context.Background(), RefreshTokenCommand{RefreshToken: "bad"})
	if err == nil {
		t.Fatal("expected refresh error")
	}
}
