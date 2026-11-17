package domain

import "strings"

// Viewer is who is asking for a document. HR (the hr_admin role) sees every document of the tenant;
// anyone else sees only the documents they own and never SECRET ones.
type Viewer struct {
	UserID string
	HR     bool
}

// Owns reports whether the document is the viewer's own: uploaded by them, or filed under their
// employee record (the employee id is the user id, as in the leave service).
func (v Viewer) Owns(d *DocumentFile) bool {
	if v.UserID == "" {
		return false
	}
	return strings.EqualFold(d.UploadedBy(), v.UserID) ||
		(d.EntityType() == EntityTypeEmployee && strings.EqualFold(d.EntityID(), v.UserID))
}

// CanView is the access rule for get, download and presign. The list queries repeat it in SQL
// (see postgres.restrictTo), so change both together.
func (v Viewer) CanView(d *DocumentFile) bool {
	return v.HR || (v.Owns(d) && d.Classification() != ClassificationSecret)
}
