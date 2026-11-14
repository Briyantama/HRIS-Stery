package domain

import "context"

// DocumentRepository defines the port for document storage
type DocumentRepository interface {
	// CreateDocument stores a new document file together with its version 1 row
	CreateDocument(ctx context.Context, doc *DocumentFile) error

	// GetDocumentByID retrieves a document by ID (tenant-scoped); ErrDocumentNotFound when absent
	GetDocumentByID(ctx context.Context, tenantID TenantID, docID DocumentID) (*DocumentFile, error)

	// GetDocumentsByEntity retrieves the documents of an entity that the viewer may see (pagination)
	GetDocumentsByEntity(ctx context.Context, tenantID TenantID, viewer Viewer, entityType EntityType, entityID string, limit, offset int) ([]*DocumentFile, int, error)

	// SearchDocuments searches documents the viewer may see by file name
	SearchDocuments(ctx context.Context, tenantID TenantID, viewer Viewer, query string, limit, offset int) ([]*DocumentFile, int, error)

	// UpdateDocumentStatus updates file status (uploaded, scanned, deleted, etc.)
	UpdateDocumentStatus(ctx context.Context, tenantID TenantID, docID DocumentID, status FileStatus) error

	// SetVersions persists the current and total version counters
	SetVersions(ctx context.Context, tenantID TenantID, docID DocumentID, current, total int32) error

	// DeleteDocument soft-deletes a document
	DeleteDocument(ctx context.Context, tenantID TenantID, docID DocumentID) error
}

// DocumentVersionRepository defines the port for version storage
type DocumentVersionRepository interface {
	// CreateVersion stores a new document version
	CreateVersion(ctx context.Context, version *DocumentVersion) error

	// GetVersionsByDocumentID retrieves all versions of a document, newest first
	GetVersionsByDocumentID(ctx context.Context, fileID DocumentID) ([]*DocumentVersion, error)

	// GetVersionByNumber retrieves one version; ErrVersionNotFound when absent
	GetVersionByNumber(ctx context.Context, fileID DocumentID, versionNumber int32) (*DocumentVersion, error)
}
