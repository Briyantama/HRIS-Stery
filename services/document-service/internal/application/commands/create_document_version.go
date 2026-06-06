package commands

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/document-service/internal/domain"
	"go.uber.org/zap"
)

// CreateDocumentVersionCommand captures version creation request
type CreateDocumentVersionCommand struct {
	TenantID   string
	DocumentID string
	MinIOKey   string
	SizeBytes  int64
	CreatedBy  string
}

// CreateDocumentVersionResult is the result of version creation
type CreateDocumentVersionResult struct {
	VersionID     string
	VersionNumber int32
	DocumentID    string
	CreatedAt     string
}

// CreateDocumentVersionHandler implements the version creation use case
type CreateDocumentVersionHandler struct {
	docRepo     domain.DocumentRepository
	versionRepo domain.DocumentVersionRepository
	logger      *zap.Logger
}

func NewCreateDocumentVersionHandler(
	docRepo domain.DocumentRepository,
	versionRepo domain.DocumentVersionRepository,
	logger *zap.Logger,
) *CreateDocumentVersionHandler {
	return &CreateDocumentVersionHandler{
		docRepo:     docRepo,
		versionRepo: versionRepo,
		logger:      logger,
	}
}

// Handle executes the create version command
func (h *CreateDocumentVersionHandler) Handle(ctx context.Context, cmd CreateDocumentVersionCommand) (*CreateDocumentVersionResult, error) {
	tenantID := domain.MustNewTenantID(cmd.TenantID)
	docID, err := domain.NewDocumentIDFromString(cmd.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("invalid document id: %w", err)
	}

	// Retrieve document
	doc, err := h.docRepo.GetDocumentByID(ctx, tenantID, docID)
	if err != nil {
		h.logger.Error("failed to get document", zap.Error(err))
		return nil, domain.ErrDocumentNotFound
	}

	if doc.IsDeleted() {
		return nil, domain.ErrDocumentDeleted
	}

	// Create new version in domain
	nextVersionNumber := doc.CurrentVersion() + 1
	version, err := domain.NewDocumentVersion(
		docID,
		nextVersionNumber,
		cmd.MinIOKey,
		cmd.SizeBytes,
		cmd.CreatedBy,
	)
	if err != nil {
		h.logger.Warn("invalid version parameters", zap.Error(err))
		return nil, fmt.Errorf("invalid version parameters: %w", err)
	}

	// Persist version
	if err = h.versionRepo.CreateVersion(ctx, version); err != nil {
		h.logger.Error("failed to create version", zap.Error(err))
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// Update document's version tracking
	doc.AddVersion()
	if err = h.docRepo.UpdateDocumentStatus(ctx, tenantID, docID, doc.Status()); err != nil {
		h.logger.Error("failed to update document version count", zap.Error(err))
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	h.logger.Info("document version created",
		zap.String("document_id", cmd.DocumentID),
		zap.Int32("version_number", nextVersionNumber),
	)

	return &CreateDocumentVersionResult{
		VersionID:     version.ID().String(),
		VersionNumber: nextVersionNumber,
		DocumentID:    cmd.DocumentID,
		CreatedAt:     version.CreatedAt().Format("2006-01-02T15:04:05Z"),
	}, nil
}
