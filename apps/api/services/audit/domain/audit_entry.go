package domain

import (
	"fmt"
	"time"
)

type AuditAction string
type ResourceType string

const (
	ActionLogin          AuditAction = "LOGIN"
	ActionLogout         AuditAction = "LOGOUT"
	ActionEmployeeCreate AuditAction = "EMPLOYEE_CREATED"
	ActionLeaveRequested AuditAction = "LEAVE_REQUESTED"
	ActionLeaveApproved  AuditAction = "LEAVE_APPROVED"
)

const (
	ResourceUser         ResourceType = "USER"
	ResourceEmployee     ResourceType = "EMPLOYEE"
	ResourceLeave        ResourceType = "LEAVE"
	ResourceNotification ResourceType = "NOTIFICATION"
)

// AuditEntry is an immutable append-only aggregate
type AuditEntry struct {
	id           AuditEntryID
	tenantID     TenantID
	actorID      string // User UUID who performed the action
	action       AuditAction
	resourceType ResourceType
	resourceID   string
	description  string
	success      bool
	errorMessage string
	changes      map[string]string // before/after JSON pairs
	createdAt    time.Time
}

func NewAuditEntry(
	tenantID TenantID,
	actorID string,
	action AuditAction,
	resourceType ResourceType,
	resourceID string,
	description string,
	success bool,
	errorMessage string,
	changes map[string]string,
) (*AuditEntry, error) {
	if tenantID.IsZero() {
		return nil, ErrInvalidTenantID
	}
	if actorID == "" {
		return nil, fmt.Errorf("actor_id required")
	}
	if action == "" {
		return nil, fmt.Errorf("action required")
	}
	if resourceType == "" {
		return nil, fmt.Errorf("resource_type required")
	}

	return &AuditEntry{
		id:           NewAuditEntryID(),
		tenantID:     tenantID,
		actorID:      actorID,
		action:       action,
		resourceType: resourceType,
		resourceID:   resourceID,
		description:  description,
		success:      success,
		errorMessage: errorMessage,
		changes:      changes,
		createdAt:    time.Now().UTC(),
	}, nil
}

// Immutable accessors
func (e *AuditEntry) ID() AuditEntryID           { return e.id }
func (e *AuditEntry) TenantID() TenantID         { return e.tenantID }
func (e *AuditEntry) ActorID() string            { return e.actorID }
func (e *AuditEntry) Action() AuditAction        { return e.action }
func (e *AuditEntry) ResourceType() ResourceType { return e.resourceType }
func (e *AuditEntry) ResourceID() string         { return e.resourceID }
func (e *AuditEntry) Description() string        { return e.description }
func (e *AuditEntry) Success() bool              { return e.success }
func (e *AuditEntry) ErrorMessage() string       { return e.errorMessage }
func (e *AuditEntry) Changes() map[string]string { return e.changes }
func (e *AuditEntry) CreatedAt() time.Time       { return e.createdAt }

// RehydrateAuditEntry reconstructs an AuditEntry from stored data (for repository reads).
func RehydrateAuditEntry(
	id AuditEntryID,
	tenantID TenantID,
	actorID string,
	action AuditAction,
	resourceType ResourceType,
	resourceID string,
	description string,
	success bool,
	errorMessage string,
	changes map[string]string,
	createdAtInterface interface{},
) (*AuditEntry, error) {
	createdAt, ok := createdAtInterface.(time.Time)
	if !ok {
		return nil, fmt.Errorf("invalid created_at type")
	}

	return &AuditEntry{
		id:           id,
		tenantID:     tenantID,
		actorID:      actorID,
		action:       action,
		resourceType: resourceType,
		resourceID:   resourceID,
		description:  description,
		success:      success,
		errorMessage: errorMessage,
		changes:      changes,
		createdAt:    createdAt,
	}, nil
}

// Invariants: No update or delete methods exist
// Immutability enforced at compile time
