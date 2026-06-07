package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
)

// GetPermissionsQuery retrieves all permissions for a user based on their roles.
type GetPermissionsQuery struct {
	UserID   domain.UserID
	TenantID domain.TenantID
}

// GetPermissionsHandler executes the get permissions query.
type GetPermissionsHandler struct {
	userRepo       domain.UserRepository
	roleRepo       domain.RoleRepository
	permissionRepo domain.PermissionRepository
}

// NewGetPermissionsHandler creates a new get permissions handler.
func NewGetPermissionsHandler(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	permissionRepo domain.PermissionRepository,
) *GetPermissionsHandler {
	return &GetPermissionsHandler{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// Handle retrieves permissions for a user.
func (h *GetPermissionsHandler) Handle(ctx context.Context, query GetPermissionsQuery) ([]domain.Permission, error) {
	// Get user
	if _, err := h.userRepo.GetByID(ctx, query.TenantID, query.UserID); err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	roles, err := h.userRepo.GetRoles(ctx, query.TenantID, query.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetch roles: %w", err)
	}

	// Convert roles to roleIDs
	roleIDs := make([]domain.RoleID, len(roles))
	for i, role := range roles {
		roleIDs[i] = role.ID()
	}

	// Get all permissions for these roles in a single query (no N+1)
	perms, err := h.permissionRepo.GetForRoles(ctx, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("fetch permissions: %w", err)
	}

	return perms, nil
}
