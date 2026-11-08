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

	// ConsumeRefreshToken verifies a refresh token and atomically spends it (single use).
	// Replayed, revoked, expired or unknown tokens return domain.ErrUnauthenticated.
	ConsumeRefreshToken(ctx context.Context, refreshToken string) (*TokenClaims, error)

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
