package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
)

// RegisterTenantCommand creates a new tenant with its first admin user.
type RegisterTenantCommand struct {
	CompanyName   string
	TenantSlug    string
	AdminEmail    string
	AdminPassword string
	AdminName     string
}

// RegisterTenantResult contains IDs of the created entities.
type RegisterTenantResult struct {
	TenantID domain.TenantID
	UserID   domain.UserID
	Email    string
}

// RegisterTenantHandler executes the register tenant command.
type RegisterTenantHandler struct {
	tenantRepo     domain.TenantRepository
	userRepo       domain.UserRepository
	roleRepo       domain.RoleRepository
	permissionRepo domain.PermissionRepository
	tokenSvc       TokenService
	eventPub       EventPublisher
}

// NewRegisterTenantHandler creates a new register tenant handler.
func NewRegisterTenantHandler(
	tenantRepo domain.TenantRepository,
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	permissionRepo domain.PermissionRepository,
	tokenSvc TokenService,
	eventPub EventPublisher,
) *RegisterTenantHandler {
	return &RegisterTenantHandler{
		tenantRepo:     tenantRepo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		tokenSvc:       tokenSvc,
		eventPub:       eventPub,
	}
}

// Handle executes the tenant registration.
// Steps:
// 1. Validate slug and create tenant
// 2. Persist tenant
// 3. Create and seed roles
// 4. Create and seed permissions
// 5. Create admin user
// 6. Assign admin user to hr_admin role
// 7. Publish events
func (h *RegisterTenantHandler) Handle(ctx context.Context, cmd RegisterTenantCommand) (*RegisterTenantResult, error) {
	// Validate slug
	slug, err := domain.NewTenantSlug(cmd.TenantSlug)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant slug: %w", err)
	}

	// Create tenant
	tenantID := domain.MustNewTenantID(uuid.New().String())
	tenant := domain.NewTenant(tenantID, slug, cmd.CompanyName)

	if err := h.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	// Create roles
	roles := make(map[string]*domain.Role)
	for _, roleName := range domain.BuiltInRoles() {
		roleID := domain.MustNewRoleID(uuid.New().String())
		role := domain.NewRole(roleID, tenantID, roleName, "", true)
		if err := h.roleRepo.Create(ctx, role); err != nil {
			return nil, fmt.Errorf("create role %s: %w", roleName, err)
		}
		roles[roleName] = role
	}

	// Create and assign permissions
	for _, perm := range getAllPermissions() {
		if err := h.permissionRepo.Create(ctx, perm.String(), ""); err != nil {
			// Silently ignore if permission already exists
			_ = err
		}

		// Assign to appropriate roles
		for roleName, role := range roles {
			defaultPerms := domain.DefaultPermissionsForRole(roleName)
			for _, defaultPerm := range defaultPerms {
				if defaultPerm.Equals(perm) {
					_ = h.permissionRepo.AssignToRole(ctx, role.ID(), perm)
				}
			}
		}
	}

	// Create admin user
	userID := domain.GenerateUserID()
	adminUser, err := domain.NewUser(userID, tenantID, cmd.AdminEmail, cmd.AdminPassword, cmd.AdminName)
	if err != nil {
		return nil, fmt.Errorf("create admin user: %w", err)
	}
	adminUser.VerifyEmail()

	if err := h.userRepo.Create(ctx, adminUser); err != nil {
		return nil, fmt.Errorf("persist admin user: %w", err)
	}

	// Assign admin user to hr_admin role
	hrAdminRole := roles["hr_admin"]
	if err := h.userRepo.SetRoles(ctx, tenantID, userID, []domain.RoleID{hrAdminRole.ID()}); err != nil {
		return nil, fmt.Errorf("assign admin role: %w", err)
	}

	err = h.eventPub.PublishAsync(ctx, domain.NewTenantCreatedEvent(tenantID, cmd.TenantSlug, cmd.CompanyName))
	if err != nil {
		return nil, fmt.Errorf("publish tenant created event: %w", err)
	}
	err = h.eventPub.PublishAsync(ctx, domain.NewUserRegisteredEvent(tenantID, userID, cmd.AdminEmail, cmd.AdminName))
	if err != nil {
		return nil, fmt.Errorf("publish user registered event: %w", err)
	}

	return &RegisterTenantResult{
		TenantID: tenantID,
		UserID:   userID,
		Email:    cmd.AdminEmail,
	}, nil
}

// getAllPermissions returns all application permissions.
func getAllPermissions() []domain.Permission {
	return []domain.Permission{
		domain.PermissionEmployeeRead,
		domain.PermissionEmployeeWrite,
		domain.PermissionEmployeeDelete,
		domain.PermissionAttendanceRead,
		domain.PermissionAttendanceWrite,
		domain.PermissionLeaveRead,
		domain.PermissionLeaveWrite,
		domain.PermissionLeaveApprove,
		domain.PermissionNotificationRead,
		domain.PermissionAuditRead,
		domain.PermissionSettingsWrite,
	}
}
