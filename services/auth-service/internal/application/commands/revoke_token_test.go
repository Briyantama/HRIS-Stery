package commands

import (
	"context"
	"fmt"
	"testing"

	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
)

type revokeTokenMock struct {
	err error
}

func (m *revokeTokenMock) GenerateTokenPair(context.Context, domain.UserID, domain.TenantID, string, []string) (string, string, int32, error) {
	return "", "", 0, fmt.Errorf("not implemented")
}

func (m *revokeTokenMock) ValidateAccessToken(context.Context, string) (*TokenClaims, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *revokeTokenMock) RefreshAccessToken(context.Context, string) (string, string, int32, error) {
	return "", "", 0, fmt.Errorf("not implemented")
}

func (m *revokeTokenMock) RevokeRefreshToken(context.Context, string) error {
	return m.err
}

func (m *revokeTokenMock) IsRefreshTokenRevoked(context.Context, string) (bool, error) {
	return false, nil
}

func TestRevokeTokenSuccess(t *testing.T) {
	handler := NewRevokeTokenHandler(&revokeTokenMock{})
	if err := handler.Handle(context.Background(), RevokeTokenCommand{RefreshToken: "refresh-1"}); err != nil {
		t.Fatalf("revoke failed: %v", err)
	}
}

func TestRevokeTokenInvalid(t *testing.T) {
	handler := NewRevokeTokenHandler(&revokeTokenMock{err: fmt.Errorf("invalid token")})
	if err := handler.Handle(context.Background(), RevokeTokenCommand{RefreshToken: "bad"}); err == nil {
		t.Fatal("expected revoke error")
	}
}
