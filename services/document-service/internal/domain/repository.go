package domain

import "context"

// DocumentRepository defines the port for document storage
type DocumentRepository interface {
	// CreateDocument stores a new document file
	CreateDocument(ctx context.Context, doc *DocumentFile) error

	// GetDocumentByID retrieves a document by ID (tenant-scoped)
	GetDocumentByID(ctx context.Context, tenantID TenantID, docID DocumentID) (*DocumentFile, error)

	// GetDocumentsByEntity retrieves documents for an entity (pagination)
	GetDocumentsByEntity(ctx context.Context, tenantID TenantID, entityType EntityType, entityID string, limit, offset int) ([]*DocumentFile, int, error)

	// SearchDocuments searches documents by file name
	SearchDocuments(ctx context.Context, tenantID TenantID, query string, limit, offset int) ([]*DocumentFile, int, error)

	// UpdateDocumentStatus updates file status (uploaded, scanned, deleted, etc.)
	UpdateDocumentStatus(ctx context.Context, tenantID TenantID, docID DocumentID, status FileStatus) error

	// DeleteDocument soft-deletes a document
	DeleteDocument(ctx context.Context, tenantID TenantID, docID DocumentID) error
}

// DocumentVersionRepository defines the port for version storage
type DocumentVersionRepository interface {
	// CreateVersion stores a new document version
	CreateVersion(ctx context.Context, version *DocumentVersion) error

	// GetVersionByID retrieves a specific version
	GetVersionByID(ctx context.Context, versionID DocumentVersionID) (*DocumentVersion, error)

	// GetVersionsByDocumentID retrieves all versions of a document
	GetVersionsByDocumentID(ctx context.Context, fileID DocumentID) ([]*DocumentVersion, error)

	// GetVersionByNumber retrieves a specific version by version number
	GetVersionByNumber(ctx context.Context, fileID DocumentID, versionNumber int32) (*DocumentVersion, error)
}
