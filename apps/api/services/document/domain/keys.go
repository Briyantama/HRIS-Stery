package domain

import "fmt"

// ObjectKey is the storage key of one version of a file. The bucket is shared by all tenants, so
// the key is prefixed with the tenant id; the rest keeps the old per-tenant-bucket layout.
func ObjectKey(tenantID TenantID, entityType EntityType, entityID string, fileID DocumentID, versionNumber int32) string {
	return fmt.Sprintf("%s/%s/%s/%d/%s", tenantID, entityType, entityID, versionNumber, fileID)
}
