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

// BaseDomainEvent provides common event fields.
type BaseDomainEvent struct {
	eventID   string
	eventType string
	occurredAt time.Time
}

// NewBaseDomainEvent creates a base event with UUID v7 ID and current timestamp.
func NewBaseDomainEvent(eventType string) BaseDomainEvent {
	return BaseDomainEvent{
		eventID:    uuid.New().String(), // In practice, use uuid v7 for time-ordering
		eventType:  eventType,
		occurredAt: time.Now().UTC(),
	}
}

// EventID returns the unique event ID.
func (e BaseDomainEvent) EventID() string {
	return e.eventID
}

// EventType returns the event type string (e.g., "UserRegistered").
func (e BaseDomainEvent) EventType() string {
	return e.eventType
}

// OccurredAt returns when the event occurred.
func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// UserRegisteredEvent fires when a new user is created.
type UserRegisteredEvent struct {
	BaseDomainEvent
	TenantID TenantID
	UserID   UserID
	Email    string
	FullName string
}

// NewUserRegisteredEvent creates the event.
func NewUserRegisteredEvent(tenantID TenantID, userID UserID, email, fullName string) UserRegisteredEvent {
	return UserRegisteredEvent{
		BaseDomainEvent: NewBaseDomainEvent("hris.identity.user.registered"),
		TenantID:        tenantID,
		UserID:          userID,
		Email:           email,
		FullName:        fullName,
	}
}

// UserLoggedInEvent fires when a user successfully authenticates.
type UserLoggedInEvent struct {
	BaseDomainEvent
	TenantID TenantID
	UserID   UserID
	Email    string
	IPAddr   string
}

// NewUserLoggedInEvent creates the event.
func NewUserLoggedInEvent(tenantID TenantID, userID UserID, email, ipAddr string) UserLoggedInEvent {
	return UserLoggedInEvent{
		BaseDomainEvent: NewBaseDomainEvent("hris.identity.user.login_succeeded"),
		TenantID:        tenantID,
		UserID:          userID,
		Email:           email,
		IPAddr:          ipAddr,
	}
}

// UserLoginFailedEvent fires when login fails (invalid credentials, inactive user, etc.).
type UserLoginFailedEvent struct {
	BaseDomainEvent
	Email    string
	TenantSlug string
	Reason   string
	IPAddr   string
}

// NewUserLoginFailedEvent creates the event.
func NewUserLoginFailedEvent(email, tenantSlug, reason, ipAddr string) UserLoginFailedEvent {
	return UserLoginFailedEvent{
		BaseDomainEvent: NewBaseDomainEvent("hris.identity.user.login_failed"),
		Email:           email,
		TenantSlug:      tenantSlug,
		Reason:          reason,
		IPAddr:          ipAddr,
	}
}

// PasswordChangedEvent fires when a user changes their password.
type PasswordChangedEvent struct {
	BaseDomainEvent
	TenantID TenantID
	UserID   UserID
}

// NewPasswordChangedEvent creates the event.
func NewPasswordChangedEvent(tenantID TenantID, userID UserID) PasswordChangedEvent {
	return PasswordChangedEvent{
		BaseDomainEvent: NewBaseDomainEvent("hris.identity.user.password_changed"),
		TenantID:        tenantID,
		UserID:          userID,
	}
}

// TenantCreatedEvent fires when a new tenant (company) is registered.
type TenantCreatedEvent struct {
	BaseDomainEvent
	TenantID    TenantID
	TenantSlug  string
	CompanyName string
}

// NewTenantCreatedEvent creates the event.
func NewTenantCreatedEvent(tenantID TenantID, tenantSlug, companyName string) TenantCreatedEvent {
	return TenantCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("hris.identity.tenant.created"),
		TenantID:        tenantID,
		TenantSlug:      tenantSlug,
		CompanyName:     companyName,
	}
}
