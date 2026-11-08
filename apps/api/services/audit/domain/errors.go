package domain

import "fmt"

var ErrInvalidTenantID = fmt.Errorf("invalid tenant_id")
var ErrAuditEntryNotFound = fmt.Errorf("audit entry not found")
