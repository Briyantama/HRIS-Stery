package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/auth-service/internal/application"
)

// ValidateTokenQuery extracts and validates an access token.
type ValidateTokenQuery struct {
	AccessToken string
}

// ValidateTokenHandler executes the validate token query.
type ValidateTokenHandler struct {
	tokenSvc application.TokenService
}

// NewValidateTokenHandler creates a new validate token handler.
func NewValidateTokenHandler(tokenSvc application.TokenService) *ValidateTokenHandler {
	return &ValidateTokenHandler{tokenSvc: tokenSvc}
}

// Handle validates the token and returns its claims.
func (h *ValidateTokenHandler) Handle(ctx context.Context, query ValidateTokenQuery) (*application.TokenClaims, error) {
	claims, err := h.tokenSvc.ValidateAccessToken(ctx, query.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}
	return claims, nil
}
