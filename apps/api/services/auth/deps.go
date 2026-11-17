package auth

import (
	"fmt"
	"time"

	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/Briyantama/stery-kit/go/events"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/Briyantama/HRIS-Stery/internal/config"
	"github.com/Briyantama/HRIS-Stery/services/auth/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/auth/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/auth/infrastructure"
	"github.com/Briyantama/HRIS-Stery/services/auth/infrastructure/postgres"
)

// Deps is everything the auth service needs from the host.
type Deps struct {
	Config   config.Config
	Pool     *pgxpool.Pool
	Redis    *redis.Client
	Events   *events.Publisher
	Verifier authn.Verifier
}

// Build creates the repositories, token service and handlers.
func Build(d Deps) (*HTTPHandler, error) {
	accessTTL := time.Duration(d.Config.AccessTokenHours) * time.Hour
	refreshTTL := time.Duration(d.Config.RefreshTokenDays) * 24 * time.Hour
	signer, err := infrastructure.NewJWTSigner(d.Config.AuthPrivateKeyPEM, d.Config.AuthPublicKeyPEM, "auth-service", accessTTL, refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("jwt signer: %w", err)
	}
	tokens := infrastructure.NewTokenService(signer, d.Redis, refreshTTL)
	pub := infrastructure.NewEventPublisher(d.Events)

	tenants := postgres.NewTenantRepository(d.Pool)
	users := postgres.NewUserRepository(d.Pool)
	roles := postgres.NewRoleRepository(d.Pool)
	perms := postgres.NewPermissionRepository(d.Pool)

	return NewHTTPHandler(
		commands.NewLoginHandler(tenants, users, tokens, pub),
		commands.NewRefreshTokenHandler(tenants, users, tokens),
		commands.NewRevokeTokenHandler(tokens),
		commands.NewRegisterTenantHandler(tenants, users, roles, perms, tokens, pub),
		queries.NewGetPermissionsHandler(users, roles, perms),
		d.Verifier,
	), nil
}
