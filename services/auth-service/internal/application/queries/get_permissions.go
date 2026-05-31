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

	permissionSet := make(map[string]domain.Permission)
	for _, role := range roles {
		perms, err := h.permissionRepo.GetForRole(ctx, role.ID())
		if err != nil {
			return nil, fmt.Errorf("get permissions for role %s: %w", role.ID().String(), err)
		}
		for _, perm := range perms {
			permissionSet[perm.String()] = perm
		}
	}

	// Return as slice
	result := make([]domain.Permission, 0, len(permissionSet))
	for _, perm := range permissionSet {
		result = append(result, perm)
	}
	return result, nil
}
