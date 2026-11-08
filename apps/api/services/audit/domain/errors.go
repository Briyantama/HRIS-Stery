package domain

import "fmt"

var ErrInvalidTenantID = fmt.Errorf("invalid tenant_id")
var ErrAuditEntryNotFound = fmt.Errorf("audit entry not found")

// ErrInvalidEntry wraps every NewAuditEntry validation failure (mapped to 422 over HTTP).
var ErrInvalidEntry = fmt.Errorf("invalid audit entry")
