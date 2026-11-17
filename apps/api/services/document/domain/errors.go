package domain

import "errors"

var (
	// ErrInvalidInput wraps every failed input validation; the HTTP layer maps it to 422.
	ErrInvalidInput          = errors.New("invalid input")
	ErrInvalidTenantID       = errors.New("invalid tenant_id")
	ErrInvalidDocumentID     = errors.New("invalid document_id")
	ErrInvalidVersionID      = errors.New("invalid version_id")
	ErrInvalidFileName       = errors.New("file_name is required")
	ErrInvalidMimeType       = errors.New("mime_type is required")
	ErrInvalidEntityType     = errors.New("invalid entity_type")
	ErrInvalidClassification = errors.New("invalid classification")
	ErrInvalidFileStatus     = errors.New("invalid file status")
	ErrDocumentNotFound      = errors.New("document not found")
	ErrVersionNotFound       = errors.New("version not found")
	ErrDocumentDeleted       = errors.New("document is deleted")
	ErrInvalidVersionNumber  = errors.New("invalid version number")
	ErrRetentionDaysNegative = errors.New("retention_days cannot be negative")
	ErrFileTooLarge          = errors.New("size_bytes exceeds the upload limit")
)

// MaxFileBytes is the largest file one upload may declare (50 MiB).
const MaxFileBytes int64 = 50 << 20
