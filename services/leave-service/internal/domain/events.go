package domain

import "time"

// DomainEvent is the interface that all domain events must implement.
type DomainEvent interface {
	EventID() string
	EventType() string
	OccurredAt() time.Time
}

// LeaveRequestedEvent is published when a leave request is created.
type LeaveRequestedEvent struct {
	eventID       string
	eventType     string
	tenantID      TenantID
	employeeID    EmployeeID
	leaveRequestID LeaveRequestID
	leaveTypeID   LeaveTypeID
	startDate     time.Time
	endDate       time.Time
	daysCount     int
	reason        string
	occurredAt    time.Time
}

// NewLeaveRequestedEvent creates a new leave requested event.
func NewLeaveRequestedEvent(
	eventID string,
	tenantID TenantID,
	employeeID EmployeeID,
	leaveRequestID LeaveRequestID,
	leaveTypeID LeaveTypeID,
	startDate time.Time,
	endDate time.Time,
	daysCount int,
	reason string,
) *LeaveRequestedEvent {
	return &LeaveRequestedEvent{
		eventID:        eventID,
		eventType:      "hris.operations.leave.requested",
		tenantID:       tenantID,
		employeeID:     employeeID,
		leaveRequestID: leaveRequestID,
		leaveTypeID:    leaveTypeID,
		startDate:      startDate,
		endDate:        endDate,
		daysCount:      daysCount,
		reason:         reason,
		occurredAt:     time.Now().UTC(),
	}
}

func (e *LeaveRequestedEvent) EventID() string {
	return e.eventID
}

func (e *LeaveRequestedEvent) EventType() string {
	return e.eventType
}

func (e *LeaveRequestedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *LeaveRequestedEvent) TenantID() TenantID {
	return e.tenantID
}

func (e *LeaveRequestedEvent) EmployeeID() EmployeeID {
	return e.employeeID
}

func (e *LeaveRequestedEvent) LeaveRequestID() LeaveRequestID {
	return e.leaveRequestID
}

// LeaveApprovedEvent is published when a leave request is approved.
type LeaveApprovedEvent struct {
	eventID        string
	eventType      string
	tenantID       TenantID
	employeeID     EmployeeID
	leaveRequestID LeaveRequestID
	approvedByID   EmployeeID
	occurredAt     time.Time
}

// NewLeaveApprovedEvent creates a new leave approved event.
func NewLeaveApprovedEvent(
	eventID string,
	tenantID TenantID,
	employeeID EmployeeID,
	leaveRequestID LeaveRequestID,
	approvedByID EmployeeID,
) *LeaveApprovedEvent {
	return &LeaveApprovedEvent{
		eventID:        eventID,
		eventType:      "hris.operations.leave.approved",
		tenantID:       tenantID,
		employeeID:     employeeID,
		leaveRequestID: leaveRequestID,
		approvedByID:   approvedByID,
		occurredAt:     time.Now().UTC(),
	}
}

func (e *LeaveApprovedEvent) EventID() string {
	return e.eventID
}

func (e *LeaveApprovedEvent) EventType() string {
	return e.eventType
}

func (e *LeaveApprovedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *LeaveApprovedEvent) TenantID() TenantID {
	return e.tenantID
}

func (e *LeaveApprovedEvent) EmployeeID() EmployeeID {
	return e.employeeID
}

func (e *LeaveApprovedEvent) LeaveRequestID() LeaveRequestID {
	return e.leaveRequestID
}

func (e *LeaveApprovedEvent) ApprovedByID() EmployeeID {
	return e.approvedByID
}

// LeaveRejectedEvent is published when a leave request is rejected.
type LeaveRejectedEvent struct {
	eventID        string
	eventType      string
	tenantID       TenantID
	employeeID     EmployeeID
	leaveRequestID LeaveRequestID
	rejectionReason string
	occurredAt     time.Time
}

// NewLeaveRejectedEvent creates a new leave rejected event.
func NewLeaveRejectedEvent(
	eventID string,
	tenantID TenantID,
	employeeID EmployeeID,
	leaveRequestID LeaveRequestID,
	rejectionReason string,
) *LeaveRejectedEvent {
	return &LeaveRejectedEvent{
		eventID:         eventID,
		eventType:       "hris.operations.leave.rejected",
		tenantID:        tenantID,
		employeeID:      employeeID,
		leaveRequestID:  leaveRequestID,
		rejectionReason: rejectionReason,
		occurredAt:      time.Now().UTC(),
	}
}

func (e *LeaveRejectedEvent) EventID() string {
	return e.eventID
}

func (e *LeaveRejectedEvent) EventType() string {
	return e.eventType
}

func (e *LeaveRejectedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *LeaveRejectedEvent) TenantID() TenantID {
	return e.tenantID
}

func (e *LeaveRejectedEvent) EmployeeID() EmployeeID {
	return e.employeeID
}

func (e *LeaveRejectedEvent) LeaveRequestID() LeaveRequestID {
	return e.leaveRequestID
}

// LeaveCancelledEvent is published when a leave request is cancelled.
type LeaveCancelledEvent struct {
	eventID        string
	eventType      string
	tenantID       TenantID
	employeeID     EmployeeID
	leaveRequestID LeaveRequestID
	reason         string
	occurredAt     time.Time
}

// NewLeaveCancelledEvent creates a new leave cancelled event.
func NewLeaveCancelledEvent(
	eventID string,
	tenantID TenantID,
	employeeID EmployeeID,
	leaveRequestID LeaveRequestID,
	reason string,
) *LeaveCancelledEvent {
	return &LeaveCancelledEvent{
		eventID:        eventID,
		eventType:      "hris.operations.leave.cancelled",
		tenantID:       tenantID,
		employeeID:     employeeID,
		leaveRequestID: leaveRequestID,
		reason:         reason,
		occurredAt:     time.Now().UTC(),
	}
}

func (e *LeaveCancelledEvent) EventID() string {
	return e.eventID
}

func (e *LeaveCancelledEvent) EventType() string {
	return e.eventType
}

func (e *LeaveCancelledEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *LeaveCancelledEvent) TenantID() TenantID {
	return e.tenantID
}

func (e *LeaveCancelledEvent) EmployeeID() EmployeeID {
	return e.employeeID
}

func (e *LeaveCancelledEvent) LeaveRequestID() LeaveRequestID {
	return e.leaveRequestID
}

// LeaveBalanceUpdatedEvent is published when a leave balance is updated.
type LeaveBalanceUpdatedEvent struct {
	eventID       string
	eventType     string
	tenantID      TenantID
	employeeID    EmployeeID
	leaveTypeID   LeaveTypeID
	year          int
	entitledDays  float64
	usedDays      float64
	pendingDays   float64
	occurredAt    time.Time
}

// NewLeaveBalanceUpdatedEvent creates a new leave balance updated event.
func NewLeaveBalanceUpdatedEvent(
	eventID string,
	tenantID TenantID,
	employeeID EmployeeID,
	leaveTypeID LeaveTypeID,
	year int,
	entitledDays float64,
	usedDays float64,
	pendingDays float64,
) *LeaveBalanceUpdatedEvent {
	return &LeaveBalanceUpdatedEvent{
		eventID:      eventID,
		eventType:    "hris.operations.leave.balance_updated",
		tenantID:     tenantID,
		employeeID:   employeeID,
		leaveTypeID:  leaveTypeID,
		year:         year,
		entitledDays: entitledDays,
		usedDays:     usedDays,
		pendingDays:  pendingDays,
		occurredAt:   time.Now().UTC(),
	}
}

func (e *LeaveBalanceUpdatedEvent) EventID() string {
	return e.eventID
}

func (e *LeaveBalanceUpdatedEvent) EventType() string {
	return e.eventType
}

func (e *LeaveBalanceUpdatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *LeaveBalanceUpdatedEvent) TenantID() TenantID {
	return e.tenantID
}

func (e *LeaveBalanceUpdatedEvent) EmployeeID() EmployeeID {
	return e.employeeID
}
