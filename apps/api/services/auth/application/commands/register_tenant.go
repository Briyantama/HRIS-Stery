package commands

import (
	"context"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
	"uuid"
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
// 4. Assign seeded permissions to roles
// 5. Create admin user
// 6. Assign admin user to hr_admin role
// 7. Publish events
func (h *RegisterTenantHandler) Handle(ctx context.Context, cmd RegisterTenantCommand) (*RegisterTenantResult, error) {
	// Validate slug
	slug, err := domain.NewTenantSlug(cmd.TenantSlug)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant slug: %w: %w", domain.ErrInvalidInput, err)
	}

	// Validate the admin user before anything is persisted, so bad input leaves no orphan tenant.
	tenantID := domain.MustNewTenantID(uuid.NewV7().String())
	userID := domain.GenerateUserID()
	adminUser, err := domain.NewUser(userID, tenantID, cmd.AdminEmail, cmd.AdminPassword, cmd.AdminName)
	if err != nil {
		return nil, fmt.Errorf("create admin user: %w: %w", domain.ErrInvalidInput, err)
	}
	adminUser.VerifyEmail()

	// Create tenant
	tenant := domain.NewTenant(tenantID, slug, cmd.CompanyName)

	if err := h.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	// Create roles
	roles := make(map[string]*domain.Role)
	for _, roleName := range domain.BuiltInRoles() {
		roleID := domain.MustNewRoleID(uuid.NewV7().String())
		role := domain.NewRole(roleID, tenantID, roleName, "", true)
		if err := h.roleRepo.Create(ctx, role); err != nil {
			return nil, fmt.Errorf("create role %s: %w", roleName, err)
		}
		roles[roleName] = role
	}

	// Permissions are global seed data (migration 0011); hris_app cannot insert them.
	// Only link them to the new roles.
	// ponytail: tenant, roles and links are separate statements, not one transaction;
	// a failure after the tenant insert leaves an orphan tenant until repos share a tx.
	for _, perm := range getAllPermissions() {
		for roleName, role := range roles {
			for _, defaultPerm := range domain.DefaultPermissionsForRole(roleName) {
				if defaultPerm.Equals(perm) {
					if err := h.permissionRepo.AssignToRole(ctx, role.ID(), perm); err != nil {
						return nil, fmt.Errorf("assign permission %s to role %s: %w", perm.String(), roleName, err)
					}
				}
			}
		}
	}

	if err := h.userRepo.Create(ctx, adminUser); err != nil {
		return nil, fmt.Errorf("persist admin user: %w", err)
	}

	// Assign admin user to hr_admin role
	hrAdminRole := roles["hr_admin"]
	if err := h.userRepo.SetRoles(ctx, tenantID, userID, []domain.RoleID{hrAdminRole.ID()}); err != nil {
		return nil, fmt.Errorf("assign admin role: %w", err)
	}

	err = h.eventPub.PublishSync(ctx, domain.NewTenantCreatedEvent(tenantID, cmd.TenantSlug, cmd.CompanyName))
	if err != nil {
		return nil, fmt.Errorf("publish tenant created event: %w", err)
	}
	err = h.eventPub.PublishSync(ctx, domain.NewUserRegisteredEvent(tenantID, userID, cmd.AdminEmail, cmd.AdminName))
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
