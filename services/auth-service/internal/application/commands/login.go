package commands

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
)

// LoginCommand represents a user login attempt.
type LoginCommand struct {
	Email      string
	Password   string
	TenantSlug string
	IPAddr     string
}

// LoginResult contains the outcome of a successful login.
type LoginResult struct {
	UserID         domain.UserID
	TenantID       domain.TenantID
	Email          string
	FullName       string
	AccessToken    string
	RefreshToken   string
	AccessTokenTTL int32 // seconds (typically 900 = 15 minutes)
	Roles          []string
}

// LoginHandler executes the login command.
type LoginHandler struct {
	tenantRepo domain.TenantRepository
	userRepo   domain.UserRepository
	tokenSvc   TokenService
	eventPub   EventPublisher
}

// NewLoginHandler creates a new login handler.
func NewLoginHandler(
	tenantRepo domain.TenantRepository,
	userRepo domain.UserRepository,
	tokenSvc TokenService,
	eventPub EventPublisher,
) *LoginHandler {
	return &LoginHandler{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
		tokenSvc:   tokenSvc,
		eventPub:   eventPub,
	}
}

// Handle executes the login logic.
// 1. Look up tenant by slug
// 2. Look up user by email within that tenant
// 3. Verify password
// 4. Generate access and refresh tokens
// 5. Publish login event
// 6. Record login in database
func (h *LoginHandler) Handle(ctx context.Context, cmd LoginCommand) (*LoginResult, error) {
	// Tenant lookup
	tenant, err := h.tenantRepo.GetBySlug(ctx, cmd.TenantSlug)
	if err != nil {
		h.publishLoginFailed(ctx, cmd.Email, cmd.TenantSlug, "tenant_not_found", cmd.IPAddr)
		return nil, fmt.Errorf("tenant not found")
	}

	if !tenant.IsActive() {
		h.publishLoginFailed(ctx, cmd.Email, cmd.TenantSlug, "tenant_inactive", cmd.IPAddr)
		return nil, fmt.Errorf("tenant is inactive")
	}

	// User lookup
	user, err := h.userRepo.GetByTenantAndEmail(ctx, tenant.ID(), cmd.Email)
	if err != nil {
		h.publishLoginFailed(ctx, cmd.Email, cmd.TenantSlug, "user_not_found", cmd.IPAddr)
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.IsActive() {
		h.publishLoginFailed(ctx, cmd.Email, cmd.TenantSlug, "user_inactive", cmd.IPAddr)
		return nil, fmt.Errorf("account is inactive")
	}

	// Password verification
	if err := user.VerifyPassword(cmd.Password); err != nil {
		h.publishLoginFailed(ctx, cmd.Email, cmd.TenantSlug, "invalid_password", cmd.IPAddr)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Record login
	user.RecordLogin(cmd.IPAddr)
	if err := h.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("record login: %w", err)
	}

	// Fetch roles and permissions
	roles, err := h.userRepo.GetRoles(ctx, tenant.ID(), user.ID())
	if err != nil {
		return nil, fmt.Errorf("fetch roles: %w", err)
	}

	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name()
	}

	// Generate tokens
	accessToken, refreshToken, ttl, err := h.tokenSvc.GenerateTokenPair(
		ctx,
		user.ID(),
		tenant.ID(),
		user.Email(),
		roleNames,
	)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	_ = h.eventPub.PublishAsync(ctx, domain.NewUserLoggedInEvent(
		tenant.ID(),
		user.ID(),
		user.Email(),
		cmd.IPAddr,
	))

	return &LoginResult{
		UserID:         user.ID(),
		TenantID:       tenant.ID(),
		Email:          user.Email(),
		FullName:       user.FullName(),
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		AccessTokenTTL: ttl,
		Roles:          roleNames,
	}, nil
}

func (h *LoginHandler) publishLoginFailed(ctx context.Context, email, tenantSlug, reason, ipAddr string) {
	_ = h.eventPub.PublishAsync(ctx, domain.NewUserLoginFailedEvent(email, tenantSlug, reason, ipAddr))
}
