package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/application"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
)

// TerminateEmployeeCommand is the input DTO for terminating an employee.
type TerminateEmployeeCommand struct {
	TenantID        string
	EmployeeID      string
	TerminationDate time.Time
	ActorID         string
}

// TerminateEmployeeResult is the output DTO.
type TerminateEmployeeResult struct {
	EmployeeID      string
	Email           string
	TerminationDate time.Time
}

// TerminateEmployeeHandler implements the terminate employee use case.
type TerminateEmployeeHandler struct {
	employeeRepo   domain.EmployeeRepository
	eventPublisher application.EventPublisher
}

// NewTerminateEmployeeHandler creates a new terminate employee handler.
func NewTerminateEmployeeHandler(
	employeeRepo domain.EmployeeRepository,
	eventPublisher application.EventPublisher,
) *TerminateEmployeeHandler {
	return &TerminateEmployeeHandler{
		employeeRepo:   employeeRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle executes the terminate employee command.
func (h *TerminateEmployeeHandler) Handle(ctx context.Context, cmd TerminateEmployeeCommand) (*TerminateEmployeeResult, error) {
	// Parse tenant ID
	tenantID, err := domain.NewTenantID(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// Parse employee ID
	employeeID, err := domain.NewEmployeeID(cmd.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee_id: %w", err)
	}

	// Fetch employee
	employee, err := h.employeeRepo.GetByID(ctx, tenantID, employeeID)
	if err != nil {
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	// Validate termination date
	if cmd.TerminationDate.IsZero() {
		cmd.TerminationDate = time.Now().UTC()
	}

	// Terminate employee
	if err := employee.Terminate(cmd.TerminationDate); err != nil {
		return nil, fmt.Errorf("terminate employee: %w", err)
	}

	// Persist
	if err := h.employeeRepo.Update(ctx, employee); err != nil {
		return nil, fmt.Errorf("persist termination: %w", err)
	}

	// Publish event
	event := domain.NewEmployeeTerminatedEvent(employee, cmd.ActorID)
	_ = h.eventPublisher.PublishAsync(ctx, event)

	return &TerminateEmployeeResult{
		EmployeeID:      employee.ID().String(),
		Email:           employee.Email(),
		TerminationDate: *employee.TerminationDate(),
	}, nil
}
