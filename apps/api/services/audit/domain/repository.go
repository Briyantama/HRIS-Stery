package domain

import "context"

type QueryFilters struct {
	ActorID      string
	Action       AuditAction
	ResourceType ResourceType
	ResourceID   string
	Limit        int
	Offset       int
}

type AuditRepository interface {
	Record(ctx context.Context, entry *AuditEntry) error
	GetByID(ctx context.Context, tenantID TenantID, entryID AuditEntryID) (*AuditEntry, error)
	Query(ctx context.Context, tenantID TenantID, filters QueryFilters) ([]*AuditEntry, int, error)
}
