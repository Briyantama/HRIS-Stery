package domain

import (
	"context"
	"time"
)

// AttendanceRepository defines the interface for persistence of attendance records.
// All methods must enforce RLS via WithTenantTx() and return domain errors.
type AttendanceRepository interface {
	// Create saves a new attendance record.
	Create(ctx context.Context, attendance *AttendanceRecord) error

	// GetByID retrieves an attendance record by ID within the tenant context.
	GetByID(ctx context.Context, tenantID TenantID, id AttendanceID) (*AttendanceRecord, error)

	// GetByEmployeeAndDate retrieves an attendance record for a specific employee on a specific date.
	GetByEmployeeAndDate(ctx context.Context, tenantID TenantID, employeeID EmployeeID, date time.Time) (*AttendanceRecord, error)

	// ListByTenant retrieves all attendance records for a tenant with optional filters.
	ListByTenant(ctx context.Context, tenantID TenantID, filters AttendanceFilters) ([]*AttendanceRecord, error)

	// Update persists changes to an attendance record.
	Update(ctx context.Context, attendance *AttendanceRecord) error

	// ListByEmployee retrieves all attendance records for a specific employee within date range.
	ListByEmployee(ctx context.Context, tenantID TenantID, employeeID EmployeeID, from time.Time, to time.Time) ([]*AttendanceRecord, error)

	// ListActiveCheckIns retrieves all currently checked-in records (no check-out) for a tenant.
	ListActiveCheckIns(ctx context.Context, tenantID TenantID) ([]*AttendanceRecord, error)
}

// AttendanceFilters defines optional filters for querying attendance records.
type AttendanceFilters struct {
	EmployeeID   *EmployeeID
	DepartmentID *string // From employee-service via gRPC
	DateFrom     *time.Time
	DateTo       *time.Time
	Status       *AttendanceStatus
	Offset       int
	Limit        int
}

// EventPublisher publishes domain events to NATS JetStream.
type EventPublisher interface {
	// Publish publishes a domain event with tenant and actor context.
	Publish(ctx context.Context, event DomainEvent, tenantID TenantID, actorID string) error
}
