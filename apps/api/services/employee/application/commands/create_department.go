package commands

import (
	"context"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/employee/domain"
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
	if cmd.Name == "" {
		return nil, fmt.Errorf("%w: department name is required", domain.ErrInvalid)
	}

	tenantID, err := domain.NewTenantID(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("%w: tenant_id: %w", domain.ErrInvalid, err)
	}

	department, err := domain.NewDepartment(tenantID, cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalid, err)
	}
	if cmd.Description != "" {
		_ = department.SetDescription(cmd.Description)
	}

	if err := h.departmentRepo.Create(ctx, department); err != nil {
		return nil, fmt.Errorf("persist department: %w", err)
	}

	return &CreateDepartmentResult{
		DepartmentID: department.ID().String(),
		Name:         department.Name(),
		Description:  department.Description(),
	}, nil
}
