package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/document-service/internal/domain"
	"go.uber.org/zap"
)

// ListDocumentsQuery captures list request
type ListDocumentsQuery struct {
	TenantID   string
	EntityType string // optional filter
	EntityID   string // optional filter
	Search     string // optional search
	PageSize   int32
	Offset     int32
}

// DocumentDTO is the query result DTO
type DocumentDTO struct {
	FileID         string
	TenantID       string
	UploadedBy     string
	EntityType     string
	EntityID       string
	FileName       string
	MimeType       string
	SizeBytes      int64
	Status         string
	Classification string
	CurrentVersion int32
	TotalVersions  int32
	CreatedAt      string
	UpdatedAt      string
}

// ListDocumentsResult is the query result
type ListDocumentsResult struct {
	Documents  []DocumentDTO
	TotalCount int32
	PageSize   int32
	Offset     int32
}

// ListDocumentsHandler implements the list query
type ListDocumentsHandler struct {
	docRepo domain.DocumentRepository
	logger  *zap.Logger
}

func NewListDocumentsHandler(
	docRepo domain.DocumentRepository,
	logger *zap.Logger,
) *ListDocumentsHandler {
	return &ListDocumentsHandler{
		docRepo: docRepo,
		logger:  logger,
	}
}

// Handle executes the list query
func (h *ListDocumentsHandler) Handle(ctx context.Context, query ListDocumentsQuery) (*ListDocumentsResult, error) {
	tenantID := domain.MustNewTenantID(query.TenantID)
	entityType := domain.EntityType(query.EntityType)

	// Determine page size limits
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	var docs []*domain.DocumentFile
	var total int
	var err error

	// Retrieve documents
	if query.Search != "" {
		docs, total, err = h.docRepo.SearchDocuments(ctx, tenantID, query.Search, int(pageSize), int(offset))
	} else if query.EntityType != "" && query.EntityID != "" {
		docs, total, err = h.docRepo.GetDocumentsByEntity(ctx, tenantID, entityType, query.EntityID, int(pageSize), int(offset))
	} else {
		h.logger.Warn("list documents requires entity filter or search", zap.String("tenant_id", query.TenantID))
		return nil, fmt.Errorf("entity filter or search query required")
	}

	if err != nil {
		h.logger.Error("failed to list documents", zap.Error(err))
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	// Map to DTOs
	dtos := make([]DocumentDTO, len(docs))
	for i, doc := range docs {
		dtos[i] = DocumentDTO{
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
			CreatedAt:      doc.CreatedAt().Format("2006-01-02T15:04:05Z"),
			UpdatedAt:      doc.UpdatedAt().Format("2006-01-02T15:04:05Z"),
		}
	}

	return &ListDocumentsResult{
		Documents:  dtos,
		TotalCount: int32(total),
		PageSize:   pageSize,
		Offset:     offset,
	}, nil
}
