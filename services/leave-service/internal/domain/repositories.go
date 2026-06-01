package domain

import (
	"context"
	"time"
)

// LeaveRequestRepository defines the interface for persistence of leave requests.
// All methods must enforce RLS via WithTenantTx() and return domain errors.
type LeaveRequestRepository interface {
	// Create saves a new leave request.
	Create(ctx context.Context, request *LeaveRequest) error

	// GetByID retrieves a leave request by ID within the tenant context.
	GetByID(ctx context.Context, tenantID TenantID, id LeaveRequestID) (*LeaveRequest, error)

	// ListByTenant retrieves all leave requests for a tenant with optional filters.
	ListByTenant(ctx context.Context, tenantID TenantID, filters LeaveFilters) ([]*LeaveRequest, error)

	// ListByEmployee retrieves all leave requests for a specific employee within a year.
	ListByEmployee(ctx context.Context, tenantID TenantID, employeeID EmployeeID, year int) ([]*LeaveRequest, error)

	// Update persists changes to a leave request.
	Update(ctx context.Context, request *LeaveRequest) error
}

// LeaveFilters defines optional filters for querying leave requests.
type LeaveFilters struct {
	EmployeeID   *EmployeeID
	Status       *LeaveStatus
	ApprovedByID *EmployeeID
	Year         *int
	DateFrom     *time.Time
	DateTo       *time.Time
	Offset       int
	Limit        int
}

// LeaveTypeRepository defines the interface for persistence of leave types.
type LeaveTypeRepository interface {
	// Create saves a new leave type.
	Create(ctx context.Context, leaveType *LeaveType) error

	// GetByID retrieves a leave type by ID within the tenant context.
	GetByID(ctx context.Context, tenantID TenantID, id LeaveTypeID) (*LeaveType, error)

	// ListByTenant retrieves all leave types for a tenant.
	ListByTenant(ctx context.Context, tenantID TenantID) ([]*LeaveType, error)
}

// LeaveBalanceRepository defines the interface for persistence of leave balances.
type LeaveBalanceRepository interface {
	// Create saves a new leave balance.
	Create(ctx context.Context, balance *LeaveBalance) error

	// GetByEmployeeAndType retrieves a balance for a specific employee and leave type.
	GetByEmployeeAndType(ctx context.Context, tenantID TenantID, employeeID EmployeeID, leaveTypeID LeaveTypeID, year int) (*LeaveBalance, error)

	// ListByEmployee retrieves all balances for a specific employee in a year.
	ListByEmployee(ctx context.Context, tenantID TenantID, employeeID EmployeeID, year int) ([]*LeaveBalance, error)

	// Update persists changes to a leave balance.
	Update(ctx context.Context, balance *LeaveBalance) error
}

// EventPublisher publishes domain events to NATS JetStream.
type EventPublisher interface {
	// Publish publishes a domain event with tenant and actor context.
	Publish(ctx context.Context, event DomainEvent, tenantID TenantID, actorID string) error
}

// LeaveType represents a configured leave type (annual, sick, etc.).
type LeaveType struct {
	id               LeaveTypeID
	tenantID         TenantID
	code             string
	name             string
	maxDaysPerYear   float64
	requiresDocument bool
	isPaid           bool
	isActive         bool
}

// NewLeaveType creates a new leave type.
func NewLeaveType(
	tenantID TenantID,
	code string,
	name string,
	maxDaysPerYear float64,
	requiresDocument bool,
	isPaid bool,
) (*LeaveType, error) {
	if tenantID.IsZero() {
		return nil, ErrMissingRequiredField
	}
	if code == "" || name == "" {
		return nil, ErrMissingRequiredField
	}
	if maxDaysPerYear < 0 {
		return nil, ErrInvalidDaysCount
	}

	return &LeaveType{
		id:               GenerateLeaveTypeID(),
		tenantID:         tenantID,
		code:             code,
		name:             name,
		maxDaysPerYear:   maxDaysPerYear,
		requiresDocument: requiresDocument,
		isPaid:           isPaid,
		isActive:         true,
	}, nil
}

// RehydrateLeaveType reconstructs a leave type from persisted data.
func RehydrateLeaveType(
	id LeaveTypeID,
	tenantID TenantID,
	code string,
	name string,
	maxDaysPerYear float64,
	requiresDocument bool,
	isPaid bool,
	isActive bool,
) *LeaveType {
	return &LeaveType{
		id:               id,
		tenantID:         tenantID,
		code:             code,
		name:             name,
		maxDaysPerYear:   maxDaysPerYear,
		requiresDocument: requiresDocument,
		isPaid:           isPaid,
		isActive:         isActive,
	}
}

// Getters

func (lt *LeaveType) ID() LeaveTypeID {
	return lt.id
}

func (lt *LeaveType) TenantID() TenantID {
	return lt.tenantID
}

func (lt *LeaveType) Code() string {
	return lt.code
}

func (lt *LeaveType) Name() string {
	return lt.name
}

func (lt *LeaveType) MaxDaysPerYear() float64 {
	return lt.maxDaysPerYear
}

func (lt *LeaveType) RequiresDocument() bool {
	return lt.requiresDocument
}

func (lt *LeaveType) IsPaid() bool {
	return lt.isPaid
}

func (lt *LeaveType) IsActive() bool {
	return lt.isActive
}
