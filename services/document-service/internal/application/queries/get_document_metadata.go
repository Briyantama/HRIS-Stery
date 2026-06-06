package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/document-service/internal/domain"
	"go.uber.org/zap"
)

// GetDocumentMetadataQuery captures metadata request
type GetDocumentMetadataQuery struct {
	TenantID        string
	DocumentID      string
	IncludeVersions bool
}

// DocumentVersionDTO is a version in query results
type DocumentVersionDTO struct {
	VersionID     string
	VersionNumber int32
	SizeBytes     int64
	CreatedBy     string
	CreatedAt     string
}

// GetDocumentMetadataResult is the query result
type GetDocumentMetadataResult struct {
	Document DocumentDTO
	Versions []DocumentVersionDTO
}

// GetDocumentMetadataHandler implements the metadata query
type GetDocumentMetadataHandler struct {
	docRepo     domain.DocumentRepository
	versionRepo domain.DocumentVersionRepository
	logger      *zap.Logger
}

func NewGetDocumentMetadataHandler(
	docRepo domain.DocumentRepository,
	versionRepo domain.DocumentVersionRepository,
	logger *zap.Logger,
) *GetDocumentMetadataHandler {
	return &GetDocumentMetadataHandler{
		docRepo:     docRepo,
		versionRepo: versionRepo,
		logger:      logger,
	}
}

// Handle executes the metadata query
func (h *GetDocumentMetadataHandler) Handle(ctx context.Context, query GetDocumentMetadataQuery) (*GetDocumentMetadataResult, error) {
	tenantID := domain.MustNewTenantID(query.TenantID)
	docID, err := domain.NewDocumentIDFromString(query.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("invalid document id: %w", err)
	}

	// Retrieve document
	doc, err := h.docRepo.GetDocumentByID(ctx, tenantID, docID)
	if err != nil {
		h.logger.Error("failed to get document", zap.Error(err))
		return nil, domain.ErrDocumentNotFound
	}

	// Map to DTO
	result := &GetDocumentMetadataResult{
		Document: DocumentDTO{
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
		},
	}

	// Retrieve versions if requested
	if query.IncludeVersions {
		versions, err := h.versionRepo.GetVersionsByDocumentID(ctx, docID)
		if err != nil {
			h.logger.Error("failed to get versions", zap.Error(err))
			return nil, fmt.Errorf("failed to get versions: %w", err)
		}

		result.Versions = make([]DocumentVersionDTO, len(versions))
		for i, v := range versions {
			result.Versions[i] = DocumentVersionDTO{
				VersionID:     v.ID().String(),
				VersionNumber: v.VersionNumber(),
				SizeBytes:     v.SizeBytes(),
				CreatedBy:     v.CreatedBy(),
				CreatedAt:     v.CreatedAt().Format("2006-01-02T15:04:05Z"),
			}
		}
	}

	return result, nil
}
