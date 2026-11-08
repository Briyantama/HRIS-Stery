package domain

import "context"

// TenantRepository defines the interface for tenant persistence.
type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id TenantID) (*Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id TenantID) error
}

// UserRepository defines the interface for user persistence.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, tenantID TenantID, id UserID) (*User, error)
	GetByTenantAndEmail(ctx context.Context, tenantID TenantID, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, tenantID TenantID, id UserID) error
	GetRoles(ctx context.Context, tenantID TenantID, userID UserID) ([]*Role, error)
	SetRoles(ctx context.Context, tenantID TenantID, userID UserID, roleIDs []RoleID) error
}

// RoleRepository defines the interface for role persistence.
type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, tenantID TenantID, id RoleID) (*Role, error)
	GetByTenantAndName(ctx context.Context, tenantID TenantID, name string) (*Role, error)
	ListByTenant(ctx context.Context, tenantID TenantID) ([]*Role, error)
	Update(ctx context.Context, role *Role) error
}

// PermissionRepository defines the interface for permission persistence.
type PermissionRepository interface {
	Create(ctx context.Context, name, description string) error
	GetByName(ctx context.Context, name string) (Permission, error)
	List(ctx context.Context) ([]Permission, error)
	AssignToRole(ctx context.Context, roleID RoleID, permission Permission) error
	RemoveFromRole(ctx context.Context, roleID RoleID, permission Permission) error
	GetForRole(ctx context.Context, roleID RoleID) ([]Permission, error)
	GetForRoles(ctx context.Context, roleIDs []RoleID) ([]Permission, error)
}
