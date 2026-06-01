package commands

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
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
	// Validate required fields
	if cmd.Title == "" {
		return nil, fmt.Errorf("position title is required")
	}
	if cmd.Level == "" {
		return nil, fmt.Errorf("position level is required")
	}

	// Parse tenant ID
	tenantID, err := domain.NewTenantID(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// Parse and validate level
	level := domain.PositionLevel(cmd.Level)
	if !level.IsValid() {
		return nil, fmt.Errorf("invalid position level: %s", cmd.Level)
	}

	// Create position domain object
	position, err := domain.NewPosition(tenantID, cmd.Title, level)
	if err != nil {
		return nil, fmt.Errorf("create position: %w", err)
	}

	// Persist
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
