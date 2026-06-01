package commands

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
)

// CreateDepartmentCommand is the input DTO.
type CreateDepartmentCommand struct {
	TenantID    string
	Name        string
	Description string
	ActorID     string
}

// CreateDepartmentResult is the output DTO.
type CreateDepartmentResult struct {
	DepartmentID string
	Name         string
	Description  string
}

// CreateDepartmentHandler implements the create department use case.
type CreateDepartmentHandler struct {
	departmentRepo domain.DepartmentRepository
}

// NewCreateDepartmentHandler creates a new create department handler.
func NewCreateDepartmentHandler(departmentRepo domain.DepartmentRepository) *CreateDepartmentHandler {
	return &CreateDepartmentHandler{departmentRepo: departmentRepo}
}

// Handle executes the create department command.
func (h *CreateDepartmentHandler) Handle(ctx context.Context, cmd CreateDepartmentCommand) (*CreateDepartmentResult, error) {
	// Validate required fields
	if cmd.Name == "" {
		return nil, fmt.Errorf("department name is required")
	}

	// Parse tenant ID
	tenantID, err := domain.NewTenantID(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// Create department domain object
	department, err := domain.NewDepartment(tenantID, cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("create department: %w", err)
	}

	// Set description if provided
	if cmd.Description != "" {
		_ = department.SetDescription(cmd.Description)
	}

	// Persist
	if err := h.departmentRepo.Create(ctx, department); err != nil {
		return nil, fmt.Errorf("persist department: %w", err)
	}

	return &CreateDepartmentResult{
		DepartmentID: department.ID().String(),
		Name:         department.Name(),
		Description:  department.Description(),
	}, nil
}
