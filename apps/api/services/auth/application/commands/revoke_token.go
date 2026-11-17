package commands

import (
	"context"
	"fmt"
)

// RevokeTokenCommand revokes (logs out) a refresh token.
type RevokeTokenCommand struct {
	RefreshToken string
}

// RevokeTokenHandler executes the revoke token command.
type RevokeTokenHandler struct {
	tokenSvc TokenService
}

// NewRevokeTokenHandler creates a new revoke token handler.
func NewRevokeTokenHandler(tokenSvc TokenService) *RevokeTokenHandler {
	return &RevokeTokenHandler{tokenSvc: tokenSvc}
}

// Handle revokes the token.
func (h *RevokeTokenHandler) Handle(ctx context.Context, cmd RevokeTokenCommand) error {
	if err := h.tokenSvc.RevokeRefreshToken(ctx, cmd.RefreshToken); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	return nil
}
