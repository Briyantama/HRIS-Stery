package domain

import (
	"fmt"
	"time"
)

// DocumentFile is the main aggregate root for file management.
// It tracks file metadata, versions, and access control.
type DocumentFile struct {
	id             DocumentID
	tenantID       TenantID
	uploadedBy     string // User UUID
	entityType     EntityType
	entityID       string // UUID of the owner entity
	fileName       string
	mimeType       string
	sizeBytes      int64
	minioBucket    string
	minioKey       string
	status         FileStatus
	classification FileClassification
	retentionDays  int32 // 0 = no expiry
	currentVersion int32
	totalVersions  int32
	createdAt      time.Time
	updatedAt      time.Time
}

// NewDocumentFile creates a new document file aggregate
func NewDocumentFile(
	tenantID TenantID,
	uploadedBy string,
	entityType EntityType,
	entityID string,
	fileName string,
	mimeType string,
	sizeBytes int64,
	minioBucket string,
	minioKey string,
	classification FileClassification,
	retentionDays int32,
) (*DocumentFile, error) {
	// Validate inputs
	if tenantID.IsZero() {
		return nil, ErrInvalidTenantID
	}
	if uploadedBy == "" {
		return nil, fmt.Errorf("uploaded_by is required")
	}
	if !ValidateEntityType(entityType) {
		return nil, ErrInvalidEntityType
	}
	if entityID == "" {
		return nil, fmt.Errorf("entity_id is required")
	}
	if fileName == "" {
		return nil, ErrInvalidFileName
	}
	if mimeType == "" {
		return nil, ErrInvalidMimeType
	}
	if sizeBytes <= 0 {
		return nil, fmt.Errorf("size_bytes must be positive")
	}
	if minioBucket == "" {
		return nil, fmt.Errorf("minio_bucket is required")
	}
	if minioKey == "" {
		return nil, fmt.Errorf("minio_key is required")
	}
	if !ValidateClassification(classification) {
		return nil, ErrInvalidClassification
	}
	if retentionDays < 0 {
		return nil, ErrRetentionDaysNegative
	}

	now := time.Now().UTC()
	return &DocumentFile{
		id:             NewDocumentID(),
		tenantID:       tenantID,
		uploadedBy:     uploadedBy,
		entityType:     entityType,
		entityID:       entityID,
		fileName:       fileName,
		mimeType:       mimeType,
		sizeBytes:      sizeBytes,
		minioBucket:    minioBucket,
		minioKey:       minioKey,
		status:         StatusUploaded,
		classification: classification,
		retentionDays:  retentionDays,
		currentVersion: 1,
		totalVersions:  1,
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

// RehydrateDocumentFile reconstructs a DocumentFile from stored data
func RehydrateDocumentFile(
	id DocumentID,
	tenantID TenantID,
	uploadedBy string,
	entityType EntityType,
	entityID string,
	fileName string,
	mimeType string,
	sizeBytes int64,
	minioBucket string,
	minioKey string,
	status FileStatus,
	classification FileClassification,
	retentionDays int32,
	currentVersion int32,
	totalVersions int32,
	createdAt time.Time,
	updatedAt time.Time,
) (*DocumentFile, error) {
	if !ValidateFileStatus(status) {
		return nil, ErrInvalidFileStatus
	}

	return &DocumentFile{
		id:             id,
		tenantID:       tenantID,
		uploadedBy:     uploadedBy,
		entityType:     entityType,
		entityID:       entityID,
		fileName:       fileName,
		mimeType:       mimeType,
		sizeBytes:      sizeBytes,
		minioBucket:    minioBucket,
		minioKey:       minioKey,
		status:         status,
		classification: classification,
		retentionDays:  retentionDays,
		currentVersion: currentVersion,
		totalVersions:  totalVersions,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}, nil
}

// Immutable accessors
func (d *DocumentFile) ID() DocumentID                     { return d.id }
func (d *DocumentFile) TenantID() TenantID                 { return d.tenantID }
func (d *DocumentFile) UploadedBy() string                 { return d.uploadedBy }
func (d *DocumentFile) EntityType() EntityType             { return d.entityType }
func (d *DocumentFile) EntityID() string                   { return d.entityID }
func (d *DocumentFile) FileName() string                   { return d.fileName }
func (d *DocumentFile) MimeType() string                   { return d.mimeType }
func (d *DocumentFile) SizeBytes() int64                   { return d.sizeBytes }
func (d *DocumentFile) MinioBucket() string                { return d.minioBucket }
func (d *DocumentFile) MinioKey() string                   { return d.minioKey }
func (d *DocumentFile) Status() FileStatus                 { return d.status }
func (d *DocumentFile) Classification() FileClassification { return d.classification }
func (d *DocumentFile) RetentionDays() int32               { return d.retentionDays }
func (d *DocumentFile) CurrentVersion() int32              { return d.currentVersion }
func (d *DocumentFile) TotalVersions() int32               { return d.totalVersions }
func (d *DocumentFile) CreatedAt() time.Time               { return d.createdAt }
func (d *DocumentFile) UpdatedAt() time.Time               { return d.updatedAt }

// IsDeleted returns true if document is soft-deleted
func (d *DocumentFile) IsDeleted() bool {
	return d.status == StatusDeleted
}

// UpdateStatus transitions document to a new status
func (d *DocumentFile) UpdateStatus(newStatus FileStatus) error {
	if !ValidateFileStatus(newStatus) {
		return ErrInvalidFileStatus
	}
	d.status = newStatus
	d.updatedAt = time.Now().UTC()
	return nil
}

// MarkScanned marks document as successfully scanned
func (d *DocumentFile) MarkScanned() {
	d.status = StatusScanned
	d.updatedAt = time.Now().UTC()
}

// MarkScanFailed marks document as having failed scan
func (d *DocumentFile) MarkScanFailed() {
	d.status = StatusScanFailed
	d.updatedAt = time.Now().UTC()
}

// MarkDeleted marks document as deleted
func (d *DocumentFile) MarkDeleted() error {
	if d.IsDeleted() {
		return fmt.Errorf("document already deleted")
	}
	d.status = StatusDeleted
	d.updatedAt = time.Now().UTC()
	return nil
}

// AddVersion increments version tracking for new uploaded version
func (d *DocumentFile) AddVersion() {
	d.currentVersion++
	d.totalVersions++
	d.updatedAt = time.Now().UTC()
}
