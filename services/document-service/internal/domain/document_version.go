package domain

import (
	"fmt"
	"time"
)

// DocumentVersion tracks a specific version of a document file
type DocumentVersion struct {
	id            DocumentVersionID
	fileID        DocumentID
	versionNumber int32
	minioKey      string
	sizeBytes     int64
	createdBy     string // User UUID
	createdAt     time.Time
}

// NewDocumentVersion creates a new document version
func NewDocumentVersion(
	fileID DocumentID,
	versionNumber int32,
	minioKey string,
	sizeBytes int64,
	createdBy string,
) (*DocumentVersion, error) {
	if fileID.IsZero() {
		return nil, ErrInvalidDocumentID
	}
	if versionNumber <= 0 {
		return nil, ErrInvalidVersionNumber
	}
	if minioKey == "" {
		return nil, fmt.Errorf("minio_key is required")
	}
	if sizeBytes <= 0 {
		return nil, fmt.Errorf("size_bytes must be positive")
	}
	if createdBy == "" {
		return nil, fmt.Errorf("created_by is required")
	}

	return &DocumentVersion{
		id:            NewDocumentVersionID(),
		fileID:        fileID,
		versionNumber: versionNumber,
		minioKey:      minioKey,
		sizeBytes:     sizeBytes,
		createdBy:     createdBy,
		createdAt:     time.Now().UTC(),
	}, nil
}

// RehydrateDocumentVersion reconstructs a DocumentVersion from stored data
func RehydrateDocumentVersion(
	id DocumentVersionID,
	fileID DocumentID,
	versionNumber int32,
	minioKey string,
	sizeBytes int64,
	createdBy string,
	createdAt time.Time,
) (*DocumentVersion, error) {
	if fileID.IsZero() {
		return nil, ErrInvalidDocumentID
	}
	if versionNumber <= 0 {
		return nil, ErrInvalidVersionNumber
	}

	return &DocumentVersion{
		id:            id,
		fileID:        fileID,
		versionNumber: versionNumber,
		minioKey:      minioKey,
		sizeBytes:     sizeBytes,
		createdBy:     createdBy,
		createdAt:     createdAt,
	}, nil
}

// Immutable accessors
func (v *DocumentVersion) ID() DocumentVersionID { return v.id }
func (v *DocumentVersion) FileID() DocumentID    { return v.fileID }
func (v *DocumentVersion) VersionNumber() int32  { return v.versionNumber }
func (v *DocumentVersion) MinioKey() string      { return v.minioKey }
func (v *DocumentVersion) SizeBytes() int64      { return v.sizeBytes }
func (v *DocumentVersion) CreatedBy() string     { return v.createdBy }
func (v *DocumentVersion) CreatedAt() time.Time  { return v.createdAt }
