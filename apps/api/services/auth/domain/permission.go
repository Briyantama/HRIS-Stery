package domain

// Permission is a string value object representing an access right.
// Format: "resource:action" (e.g., "employee:read", "leave:approve")
type Permission string

// Common permissions in the HRIS domain.
const (
	PermissionEmployeeRead     Permission = "employee:read"
	PermissionEmployeeWrite    Permission = "employee:write"
	PermissionEmployeeDelete   Permission = "employee:delete"
	PermissionAttendanceRead   Permission = "attendance:read"
	PermissionAttendanceWrite  Permission = "attendance:write"
	PermissionLeaveRead        Permission = "leave:read"
	PermissionLeaveWrite       Permission = "leave:write"
	PermissionLeaveApprove     Permission = "leave:approve"
	PermissionNotificationRead Permission = "notification:read"
	PermissionAuditRead        Permission = "audit:read"
	PermissionSettingsWrite    Permission = "settings:write"
)

// String returns the permission string.
func (p Permission) String() string {
	return string(p)
}

// Equals compares two permissions.
func (p Permission) Equals(other Permission) bool {
	return string(p) == string(other)
}

// DefaultPermissionsForRole returns the standard permissions granted to a built-in role.
func DefaultPermissionsForRole(roleName string) []Permission {
	switch roleName {
	case "hr_admin":
		return []Permission{
			PermissionEmployeeRead, PermissionEmployeeWrite, PermissionEmployeeDelete,
			PermissionAttendanceRead, PermissionAttendanceWrite,
			PermissionLeaveRead, PermissionLeaveApprove,
			PermissionNotificationRead,
			PermissionAuditRead,
			PermissionSettingsWrite,
		}
	case "manager":
		return []Permission{
			PermissionEmployeeRead,
			PermissionAttendanceRead,
			PermissionLeaveRead, PermissionLeaveApprove,
			PermissionNotificationRead,
		}
	case "employee":
		return []Permission{
			PermissionEmployeeRead,
			PermissionLeaveRead, PermissionLeaveWrite,
			PermissionNotificationRead,
		}
	default:
		return []Permission{}
	}
}
