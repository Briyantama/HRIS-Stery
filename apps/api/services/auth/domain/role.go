package domain

import (
	"fmt"
	"time"
)

// RoleID is a typed UUID for role identity.
type RoleID struct {
	value string
}

// NewRoleID creates a RoleID from a UUID string.
func NewRoleID(id string) (RoleID, error) {
	if id == "" {
		return RoleID{}, fmt.Errorf("role ID cannot be empty")
	}
	return RoleID{value: id}, nil
}

// MustNewRoleID panics if the ID is invalid. Use in tests and trusted deserialization.
func MustNewRoleID(id string) RoleID {
	roleID, err := NewRoleID(id)
	if err != nil {
		panic(err)
	}
	return roleID
}

// String returns the UUID string.
func (r RoleID) String() string {
	return r.value
}

// Role represents an RBAC role assigned within a tenant.
type Role struct {
	id          RoleID
	tenantID    TenantID
	name        string
	description string
	isSystem    bool
	createdAt   time.Time
}

// NewRole creates a new Role.
func NewRole(id RoleID, tenantID TenantID, name, description string, isSystem bool) *Role {
	return &Role{
		id:          id,
		tenantID:    tenantID,
		name:        name,
		description: description,
		isSystem:    isSystem,
		createdAt:   time.Now().UTC(),
	}
}

// ID returns the role's identifier.
func (r *Role) ID() RoleID {
	return r.id
}

// TenantID returns the tenant this role belongs to.
func (r *Role) TenantID() TenantID {
	return r.tenantID
}

// Name returns the role name.
func (r *Role) Name() string {
	return r.name
}

// Description returns the role description.
func (r *Role) Description() string {
	return r.description
}

// IsSystem returns whether this role is system-managed and cannot be deleted.
func (r *Role) IsSystem() bool {
	return r.isSystem
}

// CreatedAt returns the creation timestamp.
func (r *Role) CreatedAt() time.Time {
	return r.createdAt
}

// BuiltInRoles returns the standard roles created for every tenant on signup.
func BuiltInRoles() []string {
	return []string{"hr_admin", "manager", "employee"}
}
