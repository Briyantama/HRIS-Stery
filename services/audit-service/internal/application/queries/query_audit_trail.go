package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/hris-stery/hris-stery/services/audit-service/internal/domain"
)

type AuditQueryFilters struct {
	TenantID     string
	ActorID      string
	Action       domain.AuditAction
	ResourceType domain.ResourceType
	ResourceID   string
	Limit        int
	Offset       int
}

type AuditEntryDTO struct {
	ID           string
	TenantID     string
	ActorID      string
	Action       domain.AuditAction
	ResourceType domain.ResourceType
	ResourceID   string
	Description  string
	Success      bool
	ErrorMessage string
	Changes      map[string]string
	CreatedAt    time.Time
}

type QueryAuditTrailResult struct {
	Entries    []AuditEntryDTO
	TotalCount int
}

type QueryAuditTrailHandler interface {
	Handle(ctx context.Context, filters AuditQueryFilters) (*QueryAuditTrailResult, error)
}

type QueryAuditTrailHandlerImpl struct {
	repo domain.AuditRepository
}

func NewQueryAuditTrailHandler(repo domain.AuditRepository) QueryAuditTrailHandler {
	return &QueryAuditTrailHandlerImpl{repo: repo}
}

func (h *QueryAuditTrailHandlerImpl) Handle(ctx context.Context, filters AuditQueryFilters) (*QueryAuditTrailResult, error) {
	tenantID := domain.MustNewTenantID(filters.TenantID)

	queryFilters := domain.QueryFilters{
		ActorID:      filters.ActorID,
		Action:       filters.Action,
		ResourceType: filters.ResourceType,
		ResourceID:   filters.ResourceID,
		Limit:        filters.Limit,
		Offset:       filters.Offset,
	}

	entries, total, err := h.repo.Query(ctx, tenantID, queryFilters)
	if err != nil {
		return nil, fmt.Errorf("query audit trail: %w", err)
	}

	dtos := make([]AuditEntryDTO, len(entries))
	for i, e := range entries {
		dtos[i] = AuditEntryDTO{
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
		}
	}

	return &QueryAuditTrailResult{Entries: dtos, TotalCount: total}, nil
}
