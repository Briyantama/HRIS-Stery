package queries

import (
	"context"

	"github.com/Briyantama/HRIS-Stery/services/audit/domain"
)

// GetAuditEntryHandler loads one entry of a tenant; domain.ErrAuditEntryNotFound when absent.
type GetAuditEntryHandler interface {
	Handle(ctx context.Context, tenantID, entryID string) (*AuditEntryDTO, error)
}

type GetAuditEntryHandlerImpl struct {
	repo domain.AuditRepository
}

func NewGetAuditEntryHandler(repo domain.AuditRepository) GetAuditEntryHandler {
	return &GetAuditEntryHandlerImpl{repo: repo}
}

func (h *GetAuditEntryHandlerImpl) Handle(ctx context.Context, tenantID, entryID string) (*AuditEntryDTO, error) {
	id, err := domain.NewAuditEntryIDFromString(entryID)
	if err != nil {
		return nil, domain.ErrAuditEntryNotFound // a malformed id cannot exist
	}
	e, err := h.repo.GetByID(ctx, domain.MustNewTenantID(tenantID), id)
	if err != nil {
		return nil, err
	}
	return &AuditEntryDTO{
		ID:           e.ID().String(),
		TenantID:     e.TenantID().String(),
		ActorID:      e.ActorID(),
		Action:       e.Action(),
		ResourceType: e.ResourceType(),
		ResourceID:   e.ResourceID(),
		Description:  e.Description(),
		Success:      e.Success(),
		ErrorMessage: e.ErrorMessage(),
		Changes:      e.Changes(),
		CreatedAt:    e.CreatedAt(),
	}, nil
}
