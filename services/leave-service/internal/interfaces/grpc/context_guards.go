package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

/**
 * Context guard functions for fine-grained RBAC enforcement at the service layer.
 *
 * These functions extract user context from gRPC metadata and enforce role-based
 * access control before business logic execution. Defense-in-depth: even if
 * the gateway middleware is bypassed, the service layer enforces permissions.
 */

// extractRolesFromContext retrieves user roles from gRPC metadata.
// Roles are passed in the x-user-roles header as comma-separated values.
func extractRolesFromContext(ctx context.Context) ([]string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return []string{}, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Extract roles from x-user-roles header
	rolesHeader := md.Get("x-user-roles")
	if len(rolesHeader) == 0 {
		return []string{}, nil
	}

	// Split comma-separated roles
	roles := strings.Split(rolesHeader[0], ",")
	result := make([]string, 0, len(roles))
	for _, role := range roles {
		trimmed := strings.TrimSpace(role)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result, nil
}

// extractUserIDFromContext retrieves the authenticated user ID from gRPC metadata.
func extractUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	userIDHeader := md.Get("x-user-id")
	if len(userIDHeader) == 0 {
		return "", status.Error(codes.Unauthenticated, "x-user-id not found in metadata")
	}

	return userIDHeader[0], nil
}

// extractTenantIDFromContext retrieves the tenant ID from gRPC metadata.
func extractTenantIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	tenantIDHeader := md.Get("x-tenant-id")
	if len(tenantIDHeader) == 0 {
		return "", status.Error(codes.Unauthenticated, "x-tenant-id not found in metadata")
	}

	return tenantIDHeader[0], nil
}

// hasRole checks if the user has at least one of the required roles.
// Returns true if the user has any of the required roles.
func hasRole(userRoles []string, requiredRoles ...string) bool {
	for _, userRole := range userRoles {
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

// requireApprovalPermission enforces that the user has manager or hr_admin role.
// Used for ApproveLeave and RejectLeave operations.
func requireApprovalPermission(ctx context.Context) error {
	roles, err := extractRolesFromContext(ctx)
	if err != nil {
		return err
	}

	if !hasRole(roles, "manager", "hr_admin") {
		return status.Errorf(
			codes.PermissionDenied,
			"only managers and hr_admins can perform this operation. user roles: %v",
			roles,
		)
	}

	return nil
}

// verifyCancelPermission enforces that the user can cancel the specified leave request.
// Rules:
// - Employees can only cancel their own leave requests (actor_id == employee_id)
// - HR admins can cancel any leave request
func verifyCancelPermission(ctx context.Context, ownerEmployeeID string) error {
	roles, err := extractRolesFromContext(ctx)
	if err != nil {
		return err
	}

	// HR admin can cancel any leave
	if hasRole(roles, "hr_admin") {
		return nil
	}

	// Other users can only cancel their own leave
	userID, err := extractUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	if userID != ownerEmployeeID {
		return status.Errorf(
			codes.PermissionDenied,
			"employees can only cancel their own leave requests. user_id=%s, leave_owner=%s",
			userID,
			ownerEmployeeID,
		)
	}

	return nil
}

// requireAuthenticatedUser verifies that a user is authenticated by checking
// that user_id and tenant_id are present in the request metadata.
func requireAuthenticatedUser(ctx context.Context) error {
	userID, err := extractUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	tenantID, err := extractTenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	if userID == "" || tenantID == "" {
		return status.Error(codes.Unauthenticated, "invalid authentication context")
	}

	return nil
}
