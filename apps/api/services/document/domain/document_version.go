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
	objectKey     string
	sizeBytes     int64
	createdBy     string // User UUID
	createdAt     time.Time
}

// NewDocumentVersion creates a new document version
func NewDocumentVersion(
	fileID DocumentID,
	versionNumber int32,
	objectKey string,
	sizeBytes int64,
	createdBy string,
) (*DocumentVersion, error) {
	if fileID.IsZero() {
		return nil, ErrInvalidDocumentID
	}
	if versionNumber <= 0 {
		return nil, ErrInvalidVersionNumber
	}
	if objectKey == "" {
		return nil, fmt.Errorf("%w: object_key is required", ErrInvalidInput)
	}
	if sizeBytes <= 0 {
		return nil, fmt.Errorf("%w: size_bytes must be positive", ErrInvalidInput)
	}
	if createdBy == "" {
		return nil, fmt.Errorf("%w: created_by is required", ErrInvalidInput)
	}

	return &DocumentVersion{
		id:            NewDocumentVersionID(),
		fileID:        fileID,
		versionNumber: versionNumber,
		objectKey:     objectKey,
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
	objectKey string,
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
		objectKey:     objectKey,
		sizeBytes:     sizeBytes,
		createdBy:     createdBy,
		createdAt:     createdAt,
	}, nil
}

// Immutable accessors
func (v *DocumentVersion) ID() DocumentVersionID { return v.id }
func (v *DocumentVersion) FileID() DocumentID    { return v.fileID }
func (v *DocumentVersion) VersionNumber() int32  { return v.versionNumber }
func (v *DocumentVersion) ObjectKey() string     { return v.objectKey }
func (v *DocumentVersion) SizeBytes() int64      { return v.sizeBytes }
func (v *DocumentVersion) CreatedBy() string     { return v.createdBy }
func (v *DocumentVersion) CreatedAt() time.Time  { return v.createdAt }
