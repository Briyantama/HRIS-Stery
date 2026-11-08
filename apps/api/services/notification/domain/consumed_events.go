package domain

import (
	"time"
)

type LeaveRequestedEvent struct {
	EventID    string
	TenantID   TenantID
	ActorID    string
	OccurredAt time.Time
	EmployeeID string
	LeaveType  string
	StartDate  string
	EndDate    string
	ApproverId string
}

type LeaveApprovedEvent struct {
	EventID    string
	TenantID   TenantID
	ActorID    string
	OccurredAt time.Time
	EmployeeID string
	LeaveType  string
	StartDate  string
	EndDate    string
	ApproverId string
}

type LeaveRejectedEvent struct {
	EventID    string
	TenantID   TenantID
	ActorID    string
	OccurredAt time.Time
	EmployeeID string
	LeaveType  string
	StartDate  string
	EndDate    string
	Reason     string
}

type EmployeeCreatedEvent struct {
	EventID    string
	TenantID   TenantID
	ActorID    string
	OccurredAt time.Time
	EmployeeID string
	Email      string
	FirstName  string
	LastName   string
}

type EmployeeTerminatedEvent struct {
	EventID    string
	TenantID   TenantID
	ActorID    string
	OccurredAt time.Time
	EmployeeID string
	Email      string
	FirstName  string
	LastName   string
}

type UserRegisteredEvent struct {
	EventID    string
	TenantID   TenantID
	ActorID    string
	OccurredAt time.Time
	UserID     string
	Email      string
	FirstName  string
	LastName   string
}
