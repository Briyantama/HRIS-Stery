package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Briyantama/HRIS-Stery/services/employee/application"
	"github.com/Briyantama/HRIS-Stery/services/employee/domain"
)

// CreateEmployeeCommand is the input DTO for creating an employee.
type CreateEmployeeCommand struct {
	TenantID     string
	Email        string
	FullName     string
	Phone        string
	DepartmentID string
	PositionID   string
	ManagerID    string
	ActorID      string
}

// CreateEmployeeResult is the output DTO.
type CreateEmployeeResult struct {
	EmployeeID   string
	Email        string
	FullName     string
	DepartmentID string
	PositionID   string
}

// CreateEmployeeHandler implements the create employee use case.
type CreateEmployeeHandler struct {
	employeeRepo   domain.EmployeeRepository
	departmentRepo domain.DepartmentRepository
	positionRepo   domain.PositionRepository
	eventPublisher application.EventPublisher
}

// NewCreateEmployeeHandler creates a new create employee handler.
func NewCreateEmployeeHandler(
	employeeRepo domain.EmployeeRepository,
	departmentRepo domain.DepartmentRepository,
	positionRepo domain.PositionRepository,
	eventPublisher application.EventPublisher,
) *CreateEmployeeHandler {
	return &CreateEmployeeHandler{
		employeeRepo:   employeeRepo,
		departmentRepo: departmentRepo,
		positionRepo:   positionRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle executes the create employee command.
func (h *CreateEmployeeHandler) Handle(ctx context.Context, cmd CreateEmployeeCommand) (*CreateEmployeeResult, error) {
	if cmd.Email == "" || cmd.FullName == "" {
		return nil, fmt.Errorf("%w: email and full_name are required", domain.ErrInvalid)
	}
	if cmd.DepartmentID == "" || cmd.PositionID == "" {
		return nil, fmt.Errorf("%w: department_id and position_id are required", domain.ErrInvalid)
	}

	tenantID, err := domain.NewTenantID(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("%w: tenant_id: %w", domain.ErrInvalid, err)
	}
	departmentID, err := domain.NewDepartmentID(cmd.DepartmentID)
	if err != nil {
		return nil, fmt.Errorf("%w: department_id: %w", domain.ErrInvalid, err)
	}
	positionID, err := domain.NewPositionID(cmd.PositionID)
	if err != nil {
		return nil, fmt.Errorf("%w: position_id: %w", domain.ErrInvalid, err)
	}

	// A missing department or position is a bad reference in the request, not a missing resource.
	if _, err = h.departmentRepo.GetByID(ctx, tenantID, departmentID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("%w: department not found", domain.ErrInvalid)
		}
		return nil, fmt.Errorf("load department: %w", err)
	}
	if _, err = h.positionRepo.GetByID(ctx, tenantID, positionID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("%w: position not found", domain.ErrInvalid)
		}
		return nil, fmt.Errorf("load position: %w", err)
	}

	// Email must be unique within the tenant.
	existing, err := h.employeeRepo.GetByTenantAndEmail(ctx, tenantID, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("email already exists for this tenant: %w", domain.ErrConflict)
	}

	var managerID *domain.EmployeeID
	if cmd.ManagerID != "" {
		mid, err := domain.NewEmployeeID(cmd.ManagerID)
		if err != nil {
			return nil, fmt.Errorf("%w: manager_id: %w", domain.ErrInvalid, err)
		}
		if _, err = h.employeeRepo.GetByID(ctx, tenantID, mid); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil, fmt.Errorf("%w: manager not found", domain.ErrInvalid)
			}
			return nil, fmt.Errorf("load manager: %w", err)
		}
		managerID = &mid
	}

	employee, err := domain.NewEmployee(tenantID, cmd.Email, cmd.FullName, departmentID, positionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalid, err)
	}
	if cmd.Phone != "" {
		if err = employee.SetPhoneNumber(cmd.Phone); err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInvalid, err)
		}
	}
	if managerID != nil {
		if err = employee.SetManager(managerID); err != nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrInvalid, err)
		}
	}

	if err := h.employeeRepo.Create(ctx, employee); err != nil {
		return nil, fmt.Errorf("persist employee: %w", err)
	}

	// The row is committed: a lost event must not fail the request, but it must be visible.
	if err := h.eventPublisher.PublishSync(ctx, domain.NewEmployeeCreatedEvent(employee, cmd.ActorID)); err != nil {
		slog.ErrorContext(ctx, "publish employee created event", "tenant_id", cmd.TenantID, "employee_id", employee.ID().String(), "error", err)
	}

	return &CreateEmployeeResult{
		EmployeeID:   employee.ID().String(),
		Email:        employee.Email(),
		FullName:     employee.FullName(),
		DepartmentID: employee.DepartmentID().String(),
		PositionID:   employee.PositionID().String(),
	}, nil
}
