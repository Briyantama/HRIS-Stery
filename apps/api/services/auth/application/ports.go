package application

import (
	"context"

	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
)

// TokenService generates and validates JWT tokens.
type TokenService interface {
	// GenerateTokenPair creates both access and refresh tokens.
	// Returns: accessToken, refreshToken, accessTokenTTLSeconds, error
	GenerateTokenPair(
		ctx context.Context,
		userID domain.UserID,
		tenantID domain.TenantID,
		email string,
		roles []string,
	) (string, string, int32, error)

	// ValidateAccessToken verifies and parses an access token, returning claims.
	ValidateAccessToken(ctx context.Context, token string) (*TokenClaims, error)

	// RefreshAccessToken exchanges a refresh token for rotated access + refresh tokens.
	RefreshAccessToken(ctx context.Context, refreshToken string) (accessToken string, newRefreshToken string, ttl int32, err error)

	// RevokeRefreshToken marks a refresh token as revoked.
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
}

// TokenClaims represents the parsed claims from a JWT.
type TokenClaims struct {
	UserID    domain.UserID
	TenantID  domain.TenantID
	Email     string
	Roles     []string
	ExpiresAt int64 // Unix timestamp
}

// EventPublisher sends domain events to the event stream.
type EventPublisher interface {
	// PublishSync publishes an event and waits for confirmation.
	PublishSync(ctx context.Context, event domain.DomainEvent) error
}
