package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent is the interface all domain events must implement.
type DomainEvent interface {
	EventID() string
	EventType() string
	OccurredAt() time.Time
}

// BaseDomainEvent contains common fields for all domain events.
type BaseDomainEvent struct {
	eventID    string
	eventType  string
	occurredAt time.Time
}

// NewBaseDomainEvent creates a new base domain event.
func NewBaseDomainEvent(eventType string) BaseDomainEvent {
	return BaseDomainEvent{
		eventID:    uuid.New().String(),
		eventType:  eventType,
		occurredAt: time.Now().UTC(),
	}
}

// EventID returns the event's unique identifier.
func (e BaseDomainEvent) EventID() string {
	return e.eventID
}

// EventType returns the event type string.
func (e BaseDomainEvent) EventType() string {
	return e.eventType
}

// OccurredAt returns the timestamp when the event occurred.
func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// EmployeeCreatedEvent is published when a new employee is created.
type EmployeeCreatedEvent struct {
	BaseDomainEvent
	EmployeeID   EmployeeID
	TenantID     TenantID
	Email        string
	FullName     string
	DepartmentID DepartmentID
	PositionID   PositionID
	ActorID      string // user who created the employee, or "system" for auto-registration
}

// NewEmployeeCreatedEvent creates a new employee created event.
func NewEmployeeCreatedEvent(employee *Employee, actorID string) *EmployeeCreatedEvent {
	return &EmployeeCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("hris.workforce.employee.created"),
		EmployeeID:      employee.ID(),
		TenantID:        employee.TenantID(),
		Email:           employee.Email(),
		FullName:        employee.FullName(),
		DepartmentID:    employee.DepartmentID(),
		PositionID:      employee.PositionID(),
		ActorID:         actorID,
	}
}

// EmployeeUpdatedEvent is published when an employee is updated.
type EmployeeUpdatedEvent struct {
	BaseDomainEvent
	EmployeeID EmployeeID
	TenantID   TenantID
	Email      string
	FullName   string
	Status     EmploymentStatus
	ActorID    string
}

// NewEmployeeUpdatedEvent creates a new employee updated event.
func NewEmployeeUpdatedEvent(employee *Employee, actorID string) *EmployeeUpdatedEvent {
	return &EmployeeUpdatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("hris.workforce.employee.updated"),
		EmployeeID:      employee.ID(),
		TenantID:        employee.TenantID(),
		Email:           employee.Email(),
		FullName:        employee.FullName(),
		Status:          employee.Status(),
		ActorID:         actorID,
	}
}

// EmployeeTerminatedEvent is published when an employee is terminated.
type EmployeeTerminatedEvent struct {
	BaseDomainEvent
	EmployeeID      EmployeeID
	TenantID        TenantID
	Email           string
	FullName        string
	TerminationDate time.Time
	ActorID         string
}

// NewEmployeeTerminatedEvent creates a new employee terminated event.
func NewEmployeeTerminatedEvent(employee *Employee, actorID string) *EmployeeTerminatedEvent {
	return &EmployeeTerminatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("hris.workforce.employee.terminated"),
		EmployeeID:      employee.ID(),
		TenantID:        employee.TenantID(),
		Email:           employee.Email(),
		FullName:        employee.FullName(),
		TerminationDate: *employee.TerminationDate(),
		ActorID:         actorID,
	}
}
