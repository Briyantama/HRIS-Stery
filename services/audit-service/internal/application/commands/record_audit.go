package commands

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/audit-service/internal/domain"
)

type RecordAuditCommand struct {
	TenantID     string
	ActorID      string
	Action       domain.AuditAction
	ResourceType domain.ResourceType
	ResourceID   string
	Description  string
	Success      bool
	ErrorMessage string
	Changes      map[string]string
}

type RecordAuditResult struct {
	EntryID string
}

type RecordAuditHandler interface {
	Handle(ctx context.Context, cmd RecordAuditCommand) (*RecordAuditResult, error)
}

type RecordAuditHandlerImpl struct {
	repo domain.AuditRepository
}

func NewRecordAuditHandler(repo domain.AuditRepository) RecordAuditHandler {
	return &RecordAuditHandlerImpl{repo: repo}
}

func (h *RecordAuditHandlerImpl) Handle(ctx context.Context, cmd RecordAuditCommand) (*RecordAuditResult, error) {
	tenantID := domain.MustNewTenantID(cmd.TenantID)

	entry, err := domain.NewAuditEntry(
		tenantID,
		cmd.ActorID,
		cmd.Action,
		cmd.ResourceType,
		cmd.ResourceID,
		cmd.Description,
		cmd.Success,
		cmd.ErrorMessage,
		cmd.Changes,
	)
	if err != nil {
		return nil, fmt.Errorf("create audit entry: %w", err)
	}

	if err := h.repo.Record(ctx, entry); err != nil {
		return nil, fmt.Errorf("record audit: %w", err)
	}

	return &RecordAuditResult{EntryID: entry.ID().String()}, nil
}
