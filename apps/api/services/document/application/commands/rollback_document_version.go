package commands

import (
	"context"

	"github.com/Briyantama/HRIS-Stery/services/document/domain"
)

// RollbackDocumentVersionCommand restores an earlier version as the newest one.
type RollbackDocumentVersionCommand struct {
	TenantID      string
	DocumentID    string
	VersionNumber int32
	ActorID       string
}

// RollbackDocumentVersionHandler appends a new version that points at the old version's object, so
// history is never rewritten.
type RollbackDocumentVersionHandler struct {
	docRepo     domain.DocumentRepository
	versionRepo domain.DocumentVersionRepository
	create      *CreateDocumentVersionHandler
}

func NewRollbackDocumentVersionHandler(
	docRepo domain.DocumentRepository,
	versionRepo domain.DocumentVersionRepository,
	create *CreateDocumentVersionHandler,
) *RollbackDocumentVersionHandler {
	return &RollbackDocumentVersionHandler{docRepo: docRepo, versionRepo: versionRepo, create: create}
}

// Handle executes the rollback command
func (h *RollbackDocumentVersionHandler) Handle(ctx context.Context, cmd RollbackDocumentVersionCommand) (*CreateDocumentVersionResult, error) {
	tenantID := domain.MustNewTenantID(cmd.TenantID)
	docID, err := domain.NewDocumentIDFromString(cmd.DocumentID)
	if err != nil {
		return nil, domain.ErrDocumentNotFound
	}
	// Tenant check before touching the (tenant-less) versions table.
	if _, err = h.docRepo.GetDocumentByID(ctx, tenantID, docID); err != nil {
		return nil, err
	}
	target, err := h.versionRepo.GetVersionByNumber(ctx, docID, cmd.VersionNumber)
	if err != nil {
		return nil, err
	}
	return h.create.Handle(ctx, CreateDocumentVersionCommand{
		TenantID:   cmd.TenantID,
		DocumentID: cmd.DocumentID,
		ObjectKey:  target.ObjectKey(),
		SizeBytes:  target.SizeBytes(),
		CreatedBy:  cmd.ActorID,
	})
}
