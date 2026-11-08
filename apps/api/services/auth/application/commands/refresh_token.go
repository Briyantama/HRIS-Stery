package commands

import (
	"context"
	"fmt"
)

// RefreshTokenCommand exchanges a refresh token for new access token.
type RefreshTokenCommand struct {
	RefreshToken string
}

// RefreshTokenResult contains the new tokens.
type RefreshTokenResult struct {
	AccessToken    string
	RefreshToken   string
	AccessTokenTTL int32
}

// RefreshTokenHandler executes the refresh token command.
type RefreshTokenHandler struct {
	tokenSvc TokenService
}

// NewRefreshTokenHandler creates a new refresh token handler.
func NewRefreshTokenHandler(tokenSvc TokenService) *RefreshTokenHandler {
	return &RefreshTokenHandler{tokenSvc: tokenSvc}
}

// Handle exchanges a refresh token for new tokens.
func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd RefreshTokenCommand) (*RefreshTokenResult, error) {
	newAccessToken, newRefreshToken, ttl, err := h.tokenSvc.RefreshAccessToken(ctx, cmd.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token invalid or expired: %w", err)
	}

	return &RefreshTokenResult{
		AccessToken:    newAccessToken,
		RefreshToken:   newRefreshToken,
		AccessTokenTTL: ttl,
	}, nil
}
