package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/domain"
)

// CreateLeaveRequestCommand represents the command to create a new leave request.
type CreateLeaveRequestCommand struct {
	TenantID    string
	EmployeeID  string
	LeaveTypeID string
	StartDate   time.Time
	EndDate     time.Time
	DaysCount   int
	Reason      string
	ActorID     string
}

// CreateLeaveRequestHandler handles the create leave request command.
type CreateLeaveRequestHandler struct {
	leaveRequestRepo  domain.LeaveRequestRepository
	leaveBalanceRepo  domain.LeaveBalanceRepository
	eventPublisher    domain.EventPublisher
	employeeValidator EmployeeValidator
}

// EmployeeValidator abstracts employee-service validation.
type EmployeeValidator interface {
	// ValidateEmployeeExists checks if an employee exists in the given tenant.
	ValidateEmployeeExists(ctx context.Context, tenantID, employeeID string) (bool, error)
}

// NewCreateLeaveRequestHandler creates a new create leave request handler.
func NewCreateLeaveRequestHandler(
	leaveRequestRepo domain.LeaveRequestRepository,
	leaveBalanceRepo domain.LeaveBalanceRepository,
	eventPublisher domain.EventPublisher,
	employeeValidator EmployeeValidator,
) *CreateLeaveRequestHandler {
	return &CreateLeaveRequestHandler{
		leaveRequestRepo:  leaveRequestRepo,
		leaveBalanceRepo:  leaveBalanceRepo,
		eventPublisher:    eventPublisher,
		employeeValidator: employeeValidator,
	}
}

// CreateLeaveRequestResult is the result of a successful leave request creation.
type CreateLeaveRequestResult struct {
	LeaveRequestID string
	EmployeeID     string
	TenantID       string
	Status         string
	DaysCount      int
}

// Handle processes the create leave request command.
func (h *CreateLeaveRequestHandler) Handle(ctx context.Context, cmd CreateLeaveRequestCommand) (*CreateLeaveRequestResult, error) {
	// Validate inputs
	tenantID, err := domain.NewTenantID(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	employeeID, err := domain.NewEmployeeID(cmd.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee id: %w", err)
	}

	leaveTypeID, err := domain.NewLeaveTypeID(cmd.LeaveTypeID)
	if err != nil {
		return nil, fmt.Errorf("invalid leave type id: %w", err)
	}

	// Validate employee exists (via employee-service, not DB access)
	exists, err := h.employeeValidator.ValidateEmployeeExists(ctx, cmd.TenantID, cmd.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("validating employee: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("employee not found")
	}

	// Validate leave balance sufficiency
	year := cmd.StartDate.Year()
	balance, err := h.leaveBalanceRepo.GetByEmployeeAndType(ctx, tenantID, employeeID, leaveTypeID, year)
	if err != nil && err != domain.ErrNotFound {
		return nil, fmt.Errorf("checking balance: %w", err)
	}

	// If balance doesn't exist, employee has no entitlement for this leave type
	if balance == nil {
		return nil, fmt.Errorf("no leave entitlement for this leave type")
	}

	// Check if balance is sufficient
	if float64(cmd.DaysCount) > balance.RemainingDays() {
		return nil, fmt.Errorf("insufficient leave balance: requested %d, available %.1f", cmd.DaysCount, balance.RemainingDays())
	}

	// Create leave request aggregate
	leaveRequest, err := domain.NewLeaveRequest(
		tenantID,
		employeeID,
		leaveTypeID,
		cmd.StartDate,
		cmd.EndDate,
		cmd.DaysCount,
		cmd.Reason,
		cmd.ActorID,
	)
	if err != nil {
		return nil, fmt.Errorf("creating leave request: %w", err)
	}

	// Update balance: add to pending
	if err := balance.AddPending(float64(cmd.DaysCount)); err != nil {
		return nil, fmt.Errorf("updating balance: %w", err)
	}

	// Persist leave request and update balance atomically in a single transaction.
	// If either operation fails, both roll back.
	if err := h.leaveRequestRepo.CreateAndUpdateBalanceAtomically(ctx, leaveRequest, balance, h.leaveBalanceRepo); err != nil {
		return nil, fmt.Errorf("creating leave request: %w", err)
	}

	// Publish domain event
	eventID := uuid.New().String()
	event := domain.NewLeaveRequestedEvent(
		eventID,
		tenantID,
		employeeID,
		leaveRequest.ID(),
		leaveTypeID,
		cmd.StartDate,
		cmd.EndDate,
		cmd.DaysCount,
		cmd.Reason,
	)

	if err := h.eventPublisher.Publish(ctx, event, tenantID, cmd.ActorID); err != nil {
		return nil, fmt.Errorf("publishing event: %w", err)
	}

	return &CreateLeaveRequestResult{
		LeaveRequestID: leaveRequest.ID().String(),
		EmployeeID:     employeeID.String(),
		TenantID:       tenantID.String(),
		Status:         string(leaveRequest.Status()),
		DaysCount:      cmd.DaysCount,
	}, nil
}
