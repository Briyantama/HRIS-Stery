package application

import (
	"context"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
)

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	PublishAsync(ctx context.Context, event domain.DomainEvent) error
	PublishSync(ctx context.Context, event domain.DomainEvent) error
}

// TokenClaims represents the claims in a validated JWT token.
type TokenClaims struct {
	UserID    string
	TenantID  string
	Email     string
	Roles     []string
	ExpiresAt int64
}

// TokenService defines the interface for token validation.
type TokenService interface {
	ValidateAccessToken(ctx context.Context, token string) (*TokenClaims, error)
}

// PermissionService defines the interface for permission resolution.
type PermissionService interface {
	GetPermissions(ctx context.Context, userID, tenantID string) ([]string, error)
}
