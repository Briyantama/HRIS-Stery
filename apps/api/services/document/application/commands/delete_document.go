package commands

import (
	"context"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/document/domain"
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
	Success   bool   `json:"success"`
	DeletedAt string `json:"deleted_at"` // ISO 8601 timestamp
}

// DeleteDocumentHandler implements the soft-delete use case; stored objects are kept.
type DeleteDocumentHandler struct {
	docRepo domain.DocumentRepository
}

func NewDeleteDocumentHandler(docRepo domain.DocumentRepository) *DeleteDocumentHandler {
	return &DeleteDocumentHandler{docRepo: docRepo}
}

// Handle executes the delete command
func (h *DeleteDocumentHandler) Handle(ctx context.Context, cmd DeleteDocumentCommand) (*DeleteDocumentResult, error) {
	tenantID := domain.MustNewTenantID(cmd.TenantID)
	docID, err := domain.NewDocumentIDFromString(cmd.DocumentID)
	if err != nil {
		return nil, domain.ErrDocumentNotFound
	}
	doc, err := h.docRepo.GetDocumentByID(ctx, tenantID, docID)
	if err != nil {
		return nil, err
	}
	if err = doc.MarkDeleted(); err != nil {
		return nil, err
	}
	if err = h.docRepo.DeleteDocument(ctx, tenantID, docID); err != nil {
		return nil, fmt.Errorf("delete document: %w", err)
	}
	return &DeleteDocumentResult{Success: true, DeletedAt: doc.UpdatedAt().Format("2006-01-02T15:04:05Z")}, nil
}
