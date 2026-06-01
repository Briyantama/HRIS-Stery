package commands

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/application"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
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
	ContractType string
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
	// Validate required fields
	if cmd.Email == "" || cmd.FullName == "" {
		return nil, fmt.Errorf("email and full_name are required")
	}
	if cmd.DepartmentID == "" || cmd.PositionID == "" {
		return nil, fmt.Errorf("department_id and position_id are required")
	}

	// Parse tenant ID
	tenantID, err := domain.NewTenantID(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// Parse and validate department ID
	departmentID, err := domain.NewDepartmentID(cmd.DepartmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid department_id: %w", err)
	}

	// Parse and validate position ID
	positionID, err := domain.NewPositionID(cmd.PositionID)
	if err != nil {
		return nil, fmt.Errorf("invalid position_id: %w", err)
	}

	// Verify department exists
	_, err = h.departmentRepo.GetByID(ctx, tenantID, departmentID)
	if err != nil {
		return nil, fmt.Errorf("department not found: %w", err)
	}

	// Verify position exists
	_, err = h.positionRepo.GetByID(ctx, tenantID, positionID)
	if err != nil {
		return nil, fmt.Errorf("position not found: %w", err)
	}

	// Verify email is unique within tenant
	existing, _ := h.employeeRepo.GetByTenantAndEmail(ctx, tenantID, cmd.Email)
	if existing != nil {
		return nil, fmt.Errorf("email already exists for this tenant")
	}

	// Verify manager exists if specified
	var managerID *domain.EmployeeID
	if cmd.ManagerID != "" {
		mid, err := domain.NewEmployeeID(cmd.ManagerID)
		if err != nil {
			return nil, fmt.Errorf("invalid manager_id: %w", err)
		}
		_, err = h.employeeRepo.GetByID(ctx, tenantID, mid)
		if err != nil {
			return nil, fmt.Errorf("manager not found: %w", err)
		}
		managerID = &mid
	}

	// Create employee domain object
	employee, err := domain.NewEmployee(tenantID, cmd.Email, cmd.FullName, departmentID, positionID)
	if err != nil {
		return nil, fmt.Errorf("create employee: %w", err)
	}

	// Set optional fields
	if cmd.Phone != "" {
		_ = employee.SetPhoneNumber(cmd.Phone)
	}
	if managerID != nil {
		_ = employee.SetManager(managerID)
	}

	// Persist
	if err := h.employeeRepo.Create(ctx, employee); err != nil {
		return nil, fmt.Errorf("persist employee: %w", err)
	}

	// Publish event
	event := domain.NewEmployeeCreatedEvent(employee, cmd.ActorID)
	_ = h.eventPublisher.PublishAsync(ctx, event)

	return &CreateEmployeeResult{
		EmployeeID:   employee.ID().String(),
		Email:        employee.Email(),
		FullName:     employee.FullName(),
		DepartmentID: employee.DepartmentID().String(),
		PositionID:   employee.PositionID().String(),
	}, nil
}
