package commands

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/document-service/internal/domain"
	"go.uber.org/zap"
)

// DeleteDocumentCommand captures delete request
type DeleteDocumentCommand struct {
	TenantID   string
	DocumentID string
	DeletedBy  string
	Reason     string
}

// DeleteDocumentResult is the result of deletion
type DeleteDocumentResult struct {
	Success   bool
	DeletedAt string // ISO 8601 timestamp
}

// DeleteDocumentHandler implements the delete use case
type DeleteDocumentHandler struct {
	docRepo domain.DocumentRepository
	logger  *zap.Logger
}

func NewDeleteDocumentHandler(
	docRepo domain.DocumentRepository,
	logger *zap.Logger,
) *DeleteDocumentHandler {
	return &DeleteDocumentHandler{
		docRepo: docRepo,
		logger:  logger,
	}
}

// Handle executes the delete command
func (h *DeleteDocumentHandler) Handle(ctx context.Context, cmd DeleteDocumentCommand) (*DeleteDocumentResult, error) {
	tenantID := domain.MustNewTenantID(cmd.TenantID)
	docID, err := domain.NewDocumentIDFromString(cmd.DocumentID)
	if err != nil {
		h.logger.Warn("invalid document id", zap.Error(err))
		return nil, fmt.Errorf("invalid document id: %w", err)
	}

	// Retrieve document to verify ownership
	doc, err := h.docRepo.GetDocumentByID(ctx, tenantID, docID)
	if err != nil {
		h.logger.Error("failed to get document", zap.Error(err), zap.String("document_id", cmd.DocumentID))
		return nil, domain.ErrDocumentNotFound
	}

	// Check if already deleted
	if doc.IsDeleted() {
		return nil, fmt.Errorf("document already deleted")
	}

	// Mark as deleted in domain
	if err = doc.MarkDeleted(); err != nil {
		return nil, fmt.Errorf("failed to mark document as deleted: %w", err)
	}

	// Persist deletion
	if err = h.docRepo.DeleteDocument(ctx, tenantID, docID); err != nil {
		h.logger.Error("failed to delete document", zap.Error(err), zap.String("document_id", cmd.DocumentID))
		return nil, fmt.Errorf("failed to delete document: %w", err)
	}

	h.logger.Info("document deleted successfully",
		zap.String("document_id", cmd.DocumentID),
		zap.String("tenant_id", cmd.TenantID),
		zap.String("reason", cmd.Reason),
	)

	return &DeleteDocumentResult{
		Success:   true,
		DeletedAt: doc.UpdatedAt().Format("2006-01-02T15:04:05Z"),
	}, nil
}
