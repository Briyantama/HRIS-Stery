package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
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
	tenantRepo domain.TenantRepository
	userRepo   domain.UserRepository
	tokenSvc   TokenService
}

// NewRefreshTokenHandler creates a new refresh token handler.
func NewRefreshTokenHandler(tenantRepo domain.TenantRepository, userRepo domain.UserRepository, tokenSvc TokenService) *RefreshTokenHandler {
	return &RefreshTokenHandler{tenantRepo: tenantRepo, userRepo: userRepo, tokenSvc: tokenSvc}
}

// Handle spends the refresh token, then issues a new pair from the tenant, user and roles as they
// are now, so deactivations and role changes take effect at the next refresh.
func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd RefreshTokenCommand) (*RefreshTokenResult, error) {
	claims, err := h.tokenSvc.ConsumeRefreshToken(ctx, cmd.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token invalid or expired: %w", err)
	}

	tenant, err := h.tenantRepo.GetByID(ctx, claims.TenantID)
	if err != nil {
		return nil, refreshLookupErr("tenant", err)
	}
	if !tenant.IsActive() {
		return nil, fmt.Errorf("tenant is inactive: %w", domain.ErrUnauthenticated)
	}

	user, err := h.userRepo.GetByID(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		return nil, refreshLookupErr("user", err)
	}
	if !user.IsActive() {
		return nil, fmt.Errorf("account is inactive: %w", domain.ErrUnauthenticated)
	}

	roles, err := h.userRepo.GetRoles(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetch roles: %w", err)
	}
	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name()
	}

	accessToken, refreshToken, ttl, err := h.tokenSvc.GenerateTokenPair(ctx, user.ID(), tenant.ID(), user.Email(), roleNames)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	return &RefreshTokenResult{AccessToken: accessToken, RefreshToken: refreshToken, AccessTokenTTL: ttl}, nil
}

// refreshLookupErr maps a missing tenant/user to unauthenticated; other failures stay internal errors.
func refreshLookupErr(what string, err error) error {
	if errors.Is(err, domain.ErrNotFound) {
		return fmt.Errorf("%s not found: %w", what, domain.ErrUnauthenticated)
	}
	return fmt.Errorf("load %s: %w", what, err)
}
