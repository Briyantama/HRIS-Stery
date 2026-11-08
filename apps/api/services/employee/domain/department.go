package domain

import (
	"fmt"
	"time"
)

// Department is the aggregate root for organizational departments.
type Department struct {
	id          DepartmentID
	tenantID    TenantID
	name        string
	description string
	headID      *EmployeeID
	createdAt   time.Time
	updatedAt   time.Time
}

// NewDepartment creates a new department.
func NewDepartment(tenantID TenantID, name string) (*Department, error) {
	if name == "" {
		return nil, fmt.Errorf("department name is required")
	}

	now := time.Now().UTC()
	return &Department{
		id:        GenerateDepartmentID(),
		tenantID:  tenantID,
		name:      name,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// RehydrateDepartment reconstructs a department from persisted data.
func RehydrateDepartment(
	id DepartmentID,
	tenantID TenantID,
	name, description string,
	headID *EmployeeID,
	createdAt, updatedAt time.Time,
) (*Department, error) {
	if name == "" {
		return nil, fmt.Errorf("department name is required")
	}

	return &Department{
		id:          id,
		tenantID:    tenantID,
		name:        name,
		description: description,
		headID:      headID,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

// ID returns the department's unique identifier.
func (d *Department) ID() DepartmentID {
	return d.id
}

// TenantID returns the tenant this department belongs to.
func (d *Department) TenantID() TenantID {
	return d.tenantID
}

// Name returns the department name.
func (d *Department) Name() string {
	return d.name
}

// Description returns the department description.
func (d *Department) Description() string {
	return d.description
}

// HeadID returns the department head's employee ID (if any).
func (d *Department) HeadID() *EmployeeID {
	return d.headID
}

// CreatedAt returns the creation timestamp.
func (d *Department) CreatedAt() time.Time {
	return d.createdAt
}

// UpdatedAt returns the last update timestamp.
func (d *Department) UpdatedAt() time.Time {
	return d.updatedAt
}

// SetHead assigns a department head.
func (d *Department) SetHead(headID *EmployeeID) error {
	d.headID = headID
	d.updatedAt = time.Now().UTC()
	return nil
}

// Rename changes the department name.
func (d *Department) Rename(newName string) error {
	if newName == "" {
		return fmt.Errorf("department name is required")
	}
	d.name = newName
	d.updatedAt = time.Now().UTC()
	return nil
}

// SetDescription updates the department description.
func (d *Department) SetDescription(description string) error {
	d.description = description
	d.updatedAt = time.Now().UTC()
	return nil
}
