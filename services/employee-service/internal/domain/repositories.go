package domain

import "context"

// EmployeeRepository defines the interface for employee persistence.
type EmployeeRepository interface {
	Create(ctx context.Context, employee *Employee) error
	GetByID(ctx context.Context, tenantID TenantID, id EmployeeID) (*Employee, error)
	GetByTenantAndEmail(ctx context.Context, tenantID TenantID, email string) (*Employee, error)
	Update(ctx context.Context, employee *Employee) error
	Delete(ctx context.Context, tenantID TenantID, id EmployeeID) error
	ListByTenant(ctx context.Context, tenantID TenantID, filters EmployeeFilters) ([]*Employee, error)
	GetByManager(ctx context.Context, tenantID TenantID, managerID EmployeeID) ([]*Employee, error)
}

// EmployeeFilters defines filtering options for listing employees.
type EmployeeFilters struct {
	Status       *EmploymentStatus
	DepartmentID *DepartmentID
	SearchText   string
	Limit        int32
	Offset       int32
}

// DepartmentRepository defines the interface for department persistence.
type DepartmentRepository interface {
	Create(ctx context.Context, department *Department) error
	GetByID(ctx context.Context, tenantID TenantID, id DepartmentID) (*Department, error)
	ListByTenant(ctx context.Context, tenantID TenantID) ([]*Department, error)
	Update(ctx context.Context, department *Department) error
}

// PositionRepository defines the interface for position persistence.
type PositionRepository interface {
	Create(ctx context.Context, position *Position) error
	GetByID(ctx context.Context, tenantID TenantID, id PositionID) (*Position, error)
	ListByTenant(ctx context.Context, tenantID TenantID) ([]*Position, error)
	Update(ctx context.Context, position *Position) error
}
