package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/document/domain"
	"github.com/google/uuid"
)

// ListDocumentsQuery captures list request
type ListDocumentsQuery struct {
	TenantID   string
	Viewer     domain.Viewer // the caller; non-HR viewers get only their own documents
	EntityType string        // optional filter, requires EntityID
	EntityID   string        // optional filter, requires EntityType
	Search     string        // optional file name search
	PageSize   int32
	Offset     int32
}

// DocumentDTO is the query result DTO
type DocumentDTO struct {
	FileID         string `json:"file_id"`
	TenantID       string `json:"tenant_id"`
	UploadedBy     string `json:"uploaded_by"`
	EntityType     string `json:"entity_type"`
	EntityID       string `json:"entity_id"`
	FileName       string `json:"file_name"`
	MimeType       string `json:"mime_type"`
	SizeBytes      int64  `json:"size_bytes"`
	Status         string `json:"status"`
	Classification string `json:"classification"`
	CurrentVersion int32  `json:"current_version"`
	TotalVersions  int32  `json:"total_versions"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// ListDocumentsResult is the query result
type ListDocumentsResult struct {
	Documents  []DocumentDTO
	TotalCount int32
}

// ListDocumentsHandler implements the list query
type ListDocumentsHandler struct {
	docRepo domain.DocumentRepository
}

func NewListDocumentsHandler(docRepo domain.DocumentRepository) *ListDocumentsHandler {
	return &ListDocumentsHandler{docRepo: docRepo}
}

func toDTO(doc *domain.DocumentFile) DocumentDTO {
	return DocumentDTO{
		FileID:         doc.ID().String(),
		TenantID:       doc.TenantID().String(),
		UploadedBy:     doc.UploadedBy(),
		EntityType:     string(doc.EntityType()),
		EntityID:       doc.EntityID(),
		FileName:       doc.FileName(),
		MimeType:       doc.MimeType(),
		SizeBytes:      doc.SizeBytes(),
		Status:         string(doc.Status()),
		Classification: string(doc.Classification()),
		CurrentVersion: doc.CurrentVersion(),
		TotalVersions:  doc.TotalVersions(),
		CreatedAt:      doc.CreatedAt().UTC().Format(time.RFC3339),
		UpdatedAt:      doc.UpdatedAt().UTC().Format(time.RFC3339),
	}
}

// Handle executes the list query. The caller clamps PageSize and Offset; a search term wins over
// the entity filter, and one of the two is required.
func (h *ListDocumentsHandler) Handle(ctx context.Context, query ListDocumentsQuery) (*ListDocumentsResult, error) {
	tenantID, err := domain.NewTenantID(query.TenantID)
	if err != nil {
		return nil, err
	}

	var docs []*domain.DocumentFile
	var total int
	switch {
	case query.Search != "":
		docs, total, err = h.docRepo.SearchDocuments(ctx, tenantID, query.Viewer, query.Search, int(query.PageSize), int(query.Offset))
	case query.EntityType != "" && query.EntityID != "":
		if !domain.ValidateEntityType(domain.EntityType(query.EntityType)) {
			return nil, fmt.Errorf("%w: %w", domain.ErrInvalidInput, domain.ErrInvalidEntityType)
		}
		if _, perr := uuid.Parse(query.EntityID); perr != nil {
			return nil, fmt.Errorf("%w: entity_id must be a uuid", domain.ErrInvalidInput)
		}
		docs, total, err = h.docRepo.GetDocumentsByEntity(ctx, tenantID, query.Viewer, domain.EntityType(query.EntityType), query.EntityID, int(query.PageSize), int(query.Offset))
	default:
		return nil, fmt.Errorf("%w: entity filter or search query required", domain.ErrInvalidInput)
	}
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}

	dtos := make([]DocumentDTO, len(docs))
	for i, doc := range docs {
		dtos[i] = toDTO(doc)
	}
	return &ListDocumentsResult{Documents: dtos, TotalCount: int32(total)}, nil
}
