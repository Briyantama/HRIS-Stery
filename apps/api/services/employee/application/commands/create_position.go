package commands

import (
	"context"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/employee/domain"
)

// CreatePositionCommand is the input DTO.
type CreatePositionCommand struct {
	TenantID    string
	Title       string
	Description string
	Level       string
	ActorID     string
}

// CreatePositionResult is the output DTO.
type CreatePositionResult struct {
	PositionID  string
	Title       string
	Description string
	Level       string
}

// CreatePositionHandler implements the create position use case.
type CreatePositionHandler struct {
	positionRepo domain.PositionRepository
}

// NewCreatePositionHandler creates a new create position handler.
func NewCreatePositionHandler(positionRepo domain.PositionRepository) *CreatePositionHandler {
	return &CreatePositionHandler{positionRepo: positionRepo}
}

// Handle executes the create position command.
func (h *CreatePositionHandler) Handle(ctx context.Context, cmd CreatePositionCommand) (*CreatePositionResult, error) {
	if cmd.Title == "" {
		return nil, fmt.Errorf("%w: position title is required", domain.ErrInvalid)
	}
	if cmd.Level == "" {
		return nil, fmt.Errorf("%w: position level is required", domain.ErrInvalid)
	}

	tenantID, err := domain.NewTenantID(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("%w: tenant_id: %w", domain.ErrInvalid, err)
	}

	level := domain.PositionLevel(cmd.Level)
	if !level.IsValid() {
		return nil, fmt.Errorf("%w: invalid position level", domain.ErrInvalid)
	}

	position, err := domain.NewPosition(tenantID, cmd.Title, level)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalid, err)
	}

	if err := h.positionRepo.Create(ctx, position); err != nil {
		return nil, fmt.Errorf("persist position: %w", err)
	}

	return &CreatePositionResult{
		PositionID:  position.ID().String(),
		Title:       position.Title(),
		Description: position.Description(),
		Level:       string(position.Level()),
	}, nil
}
