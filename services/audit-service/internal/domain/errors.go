package domain

import "fmt"

var ErrInvalidTenantID = fmt.Errorf("invalid tenant_id")
var ErrAuditEntryNotFound = fmt.Errorf("audit entry not found")

type AuditNotFoundError struct {
	entryID string
}

func (e *AuditNotFoundError) Error() string {
	return fmt.Sprintf("audit entry not found: %s", e.entryID)
}
