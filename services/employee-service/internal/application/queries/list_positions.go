package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
)

// ListPositionsQuery is the input DTO.
type ListPositionsQuery struct {
	TenantID string
}

// PositionSummary is a brief position record.
type PositionSummary struct {
	ID          string
	Title       string
	Description string
	Level       string
	CreatedAt   int64
}

// ListPositionsResult is the output DTO.
type ListPositionsResult struct {
	Positions []*PositionSummary
	Total     int32
}

// ListPositionsHandler implements the list positions use case.
type ListPositionsHandler struct {
	positionRepo domain.PositionRepository
}

// NewListPositionsHandler creates a new list positions handler.
func NewListPositionsHandler(positionRepo domain.PositionRepository) *ListPositionsHandler {
	return &ListPositionsHandler{positionRepo: positionRepo}
}

// Handle executes the list positions query.
func (h *ListPositionsHandler) Handle(ctx context.Context, query ListPositionsQuery) (*ListPositionsResult, error) {
	// Parse tenant ID
	tenantID, err := domain.NewTenantID(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// Fetch positions
	positions, err := h.positionRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list positions: %w", err)
	}

	// Build result
	summaries := make([]*PositionSummary, len(positions))
	for i, pos := range positions {
		summaries[i] = &PositionSummary{
			ID:          pos.ID().String(),
			Title:       pos.Title(),
			Description: pos.Description(),
			Level:       string(pos.Level()),
			CreatedAt:   pos.CreatedAt().Unix(),
		}
	}

	return &ListPositionsResult{
		Positions: summaries,
		Total:     int32(len(summaries)),
	}, nil
}
