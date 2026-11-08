package domain

import (
	"context"
	"time"
)

// AttendanceRepository defines the interface for persistence of attendance records.
// All methods enforce RLS via pgtenant.WithTenantTx and return domain errors.
type AttendanceRepository interface {
	// Create saves a new attendance record.
	Create(ctx context.Context, attendance *AttendanceRecord) error

	// GetByID retrieves an attendance record by ID within the tenant context.
	GetByID(ctx context.Context, tenantID TenantID, id AttendanceID) (*AttendanceRecord, error)

	// GetByEmployeeAndDate retrieves an attendance record for a specific employee on a specific date.
	// It returns (nil, nil) when there is none.
	GetByEmployeeAndDate(ctx context.Context, tenantID TenantID, employeeID EmployeeID, date time.Time) (*AttendanceRecord, error)

	// ListByTenant retrieves attendance records for a tenant with optional filters and paging.
	ListByTenant(ctx context.Context, tenantID TenantID, filters AttendanceFilters) ([]*AttendanceRecord, error)

	// CountByTenant counts the records ListByTenant would return without paging.
	CountByTenant(ctx context.Context, tenantID TenantID, filters AttendanceFilters) (int, error)

	// Update persists changes to an attendance record.
	Update(ctx context.Context, attendance *AttendanceRecord) error
}

// AttendanceFilters defines optional filters for querying attendance records.
type AttendanceFilters struct {
	EmployeeID *EmployeeID
	DateFrom   *time.Time
	DateTo     *time.Time
	Status     *AttendanceStatus
	Offset     int
	Limit      int
}

// EventPublisher publishes domain events to the event stream.
type EventPublisher interface {
	// Publish publishes a domain event with tenant and actor context.
	Publish(ctx context.Context, event DomainEvent, tenantID TenantID, actorID string) error
}
