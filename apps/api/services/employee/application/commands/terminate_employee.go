package commands

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/employee/application"
	"github.com/Briyantama/HRIS-Stery/services/employee/domain"
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
		return nil, fmt.Errorf("%w: employee_id: %w", domain.ErrInvalid, err)
	}

	// Fetch employee
	employee, err := h.employeeRepo.GetByID(ctx, tenantID, employeeID)
	if err != nil {
		return nil, fmt.Errorf("load employee: %w", err)
	}

	// Validate termination date
	if cmd.TerminationDate.IsZero() {
		cmd.TerminationDate = time.Now().UTC()
	}

	// Terminate employee
	if err := employee.Terminate(cmd.TerminationDate); err != nil {
		return nil, fmt.Errorf("terminate employee: %w: %w", domain.ErrConflict, err)
	}

	// Persist
	if err := h.employeeRepo.Update(ctx, employee); err != nil {
		return nil, fmt.Errorf("persist termination: %w", err)
	}

	// Publish event
	// The row is committed: a lost event must not fail the request, but it must be visible.
	if err := h.eventPublisher.PublishSync(ctx, domain.NewEmployeeTerminatedEvent(employee, cmd.ActorID)); err != nil {
		slog.ErrorContext(ctx, "publish employee terminated event", "tenant_id", cmd.TenantID, "employee_id", cmd.EmployeeID, "error", err)
	}

	return &TerminateEmployeeResult{
		EmployeeID:      employee.ID().String(),
		Email:           employee.Email(),
		TerminationDate: *employee.TerminationDate(),
	}, nil
}
