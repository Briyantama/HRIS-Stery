package commands

import (
	"context"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/document/domain"
)

// CreateDocumentVersionCommand captures version creation request
type CreateDocumentVersionCommand struct {
	TenantID   string
	DocumentID string
	ObjectKey  string
	SizeBytes  int64
	CreatedBy  string
}

// CreateDocumentVersionResult is the result of version creation
type CreateDocumentVersionResult struct {
	VersionID     string `json:"version_id"`
	VersionNumber int32  `json:"version_number"`
	DocumentID    string `json:"file_id"`
	CreatedAt     string `json:"created_at"`
}

// CreateDocumentVersionHandler implements the version creation use case
type CreateDocumentVersionHandler struct {
	docRepo     domain.DocumentRepository
	versionRepo domain.DocumentVersionRepository
}

func NewCreateDocumentVersionHandler(docRepo domain.DocumentRepository, versionRepo domain.DocumentVersionRepository) *CreateDocumentVersionHandler {
	return &CreateDocumentVersionHandler{docRepo: docRepo, versionRepo: versionRepo}
}

// Handle executes the create version command
func (h *CreateDocumentVersionHandler) Handle(ctx context.Context, cmd CreateDocumentVersionCommand) (*CreateDocumentVersionResult, error) {
	tenantID := domain.MustNewTenantID(cmd.TenantID)
	docID, err := domain.NewDocumentIDFromString(cmd.DocumentID)
	if err != nil {
		return nil, domain.ErrDocumentNotFound
	}
	doc, err := h.docRepo.GetDocumentByID(ctx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if doc.IsDeleted() {
		return nil, domain.ErrDocumentDeleted
	}

	nextVersionNumber := doc.CurrentVersion() + 1
	version, err := domain.NewDocumentVersion(docID, nextVersionNumber, cmd.ObjectKey, cmd.SizeBytes, cmd.CreatedBy)
	if err != nil {
		return nil, err
	}
	if err = h.versionRepo.CreateVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("create version: %w", err)
	}
	doc.AddVersion()
	if err = h.docRepo.SetVersions(ctx, tenantID, docID, doc.CurrentVersion(), doc.TotalVersions()); err != nil {
		return nil, fmt.Errorf("update version counters: %w", err)
	}

	return &CreateDocumentVersionResult{
		VersionID:     version.ID().String(),
		VersionNumber: nextVersionNumber,
		DocumentID:    cmd.DocumentID,
		CreatedAt:     version.CreatedAt().Format("2006-01-02T15:04:05Z"),
	}, nil
}
