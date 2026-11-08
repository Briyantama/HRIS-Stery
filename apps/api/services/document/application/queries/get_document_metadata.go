package queries

import (
	"context"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/document/domain"
)

// GetDocumentMetadataQuery captures metadata request
type GetDocumentMetadataQuery struct {
	TenantID        string
	DocumentID      string
	IncludeVersions bool
}

// DocumentVersionDTO is a version in query results
type DocumentVersionDTO struct {
	VersionID     string `json:"version_id"`
	VersionNumber int32  `json:"version_number"`
	SizeBytes     int64  `json:"size_bytes"`
	CreatedBy     string `json:"created_by"`
	CreatedAt     string `json:"created_at"`
}

// GetDocumentMetadataResult is the query result
type GetDocumentMetadataResult struct {
	Document DocumentDTO          `json:"file"`
	Versions []DocumentVersionDTO `json:"versions"`
}

// GetDocumentMetadataHandler implements the metadata query
type GetDocumentMetadataHandler struct {
	docRepo     domain.DocumentRepository
	versionRepo domain.DocumentVersionRepository
}

func NewGetDocumentMetadataHandler(docRepo domain.DocumentRepository, versionRepo domain.DocumentVersionRepository) *GetDocumentMetadataHandler {
	return &GetDocumentMetadataHandler{docRepo: docRepo, versionRepo: versionRepo}
}

// Handle executes the metadata query
func (h *GetDocumentMetadataHandler) Handle(ctx context.Context, query GetDocumentMetadataQuery) (*GetDocumentMetadataResult, error) {
	tenantID := domain.MustNewTenantID(query.TenantID)
	docID, err := domain.NewDocumentIDFromString(query.DocumentID)
	if err != nil {
		return nil, domain.ErrDocumentNotFound
	}
	doc, err := h.docRepo.GetDocumentByID(ctx, tenantID, docID)
	if err != nil {
		return nil, err
	}

	result := &GetDocumentMetadataResult{Document: toDTO(doc), Versions: []DocumentVersionDTO{}}
	if query.IncludeVersions {
		versions, err := h.versionRepo.GetVersionsByDocumentID(ctx, docID)
		if err != nil {
			return nil, fmt.Errorf("get versions: %w", err)
		}
		for _, v := range versions {
			result.Versions = append(result.Versions, DocumentVersionDTO{
				VersionID:     v.ID().String(),
				VersionNumber: v.VersionNumber(),
				SizeBytes:     v.SizeBytes(),
				CreatedBy:     v.CreatedBy(),
				CreatedAt:     v.CreatedAt().Format("2006-01-02T15:04:05Z"),
			})
		}
	}
	return result, nil
}
