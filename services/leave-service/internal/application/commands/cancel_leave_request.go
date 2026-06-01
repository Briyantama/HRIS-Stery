package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/domain"
)

// CancelLeaveRequestCommand represents the command to cancel a leave request.
type CancelLeaveRequestCommand struct {
	TenantID       string
	LeaveRequestID string
	Reason         string
	ActorID        string
}

// CancelLeaveRequestHandler handles the cancel leave request command.
type CancelLeaveRequestHandler struct {
	leaveRequestRepo domain.LeaveRequestRepository
	leaveBalanceRepo domain.LeaveBalanceRepository
	eventPublisher   domain.EventPublisher
}

// NewCancelLeaveRequestHandler creates a new cancel leave request handler.
func NewCancelLeaveRequestHandler(
	leaveRequestRepo domain.LeaveRequestRepository,
	leaveBalanceRepo domain.LeaveBalanceRepository,
	eventPublisher domain.EventPublisher,
) *CancelLeaveRequestHandler {
	return &CancelLeaveRequestHandler{
		leaveRequestRepo: leaveRequestRepo,
		leaveBalanceRepo: leaveBalanceRepo,
		eventPublisher:   eventPublisher,
	}
}

// CancelLeaveRequestResult is the result of a successful cancellation.
type CancelLeaveRequestResult struct {
	LeaveRequestID string
	Status         string
}

// Handle processes the cancel leave request command.
func (h *CancelLeaveRequestHandler) Handle(ctx context.Context, cmd CancelLeaveRequestCommand) (*CancelLeaveRequestResult, error) {
	// Validate inputs
	tenantID, err := domain.NewTenantID(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	leaveRequestID, err := domain.NewLeaveRequestID(cmd.LeaveRequestID)
	if err != nil {
		return nil, fmt.Errorf("invalid leave request id: %w", err)
	}

	// Fetch the leave request
	leaveRequest, err := h.leaveRequestRepo.GetByID(ctx, tenantID, leaveRequestID)
	if err != nil {
		return nil, fmt.Errorf("fetching leave request: %w", err)
	}

	if leaveRequest == nil {
		return nil, fmt.Errorf("leave request not found")
	}

	// Store original status for balance update logic
	originalStatus := leaveRequest.Status()

	// Cancel the request (enforces state machine invariants)
	if err := leaveRequest.Cancel(cmd.Reason); err != nil {
		return nil, fmt.Errorf("cancelling leave request: %w", err)
	}

	// Update balance based on original status
	year := leaveRequest.StartDate().Year()
	balance, err := h.leaveBalanceRepo.GetByEmployeeAndType(
		ctx,
		tenantID,
		leaveRequest.EmployeeID(),
		leaveRequest.LeaveTypeID(),
		year,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching balance: %w", err)
	}

	if balance == nil {
		return nil, fmt.Errorf("balance not found")
	}

	daysFloat := float64(leaveRequest.DaysCount())

	// If originally APPROVED, remove from used; if PENDING, remove from pending
	if originalStatus == domain.StatusApproved {
		if err := balance.SubtractUsed(daysFloat); err != nil {
			return nil, fmt.Errorf("subtracting used days: %w", err)
		}
	} else if originalStatus == domain.StatusPending {
		if err := balance.RemovePending(daysFloat); err != nil {
			return nil, fmt.Errorf("removing pending days: %w", err)
		}
	}

	if err := h.leaveBalanceRepo.Update(ctx, balance); err != nil {
		return nil, fmt.Errorf("saving balance: %w", err)
	}

	// Persist the updated leave request
	if err := h.leaveRequestRepo.Update(ctx, leaveRequest); err != nil {
		return nil, fmt.Errorf("saving leave request: %w", err)
	}

	// Publish domain event
	eventID := uuid.New().String()
	event := domain.NewLeaveCancelledEvent(
		eventID,
		tenantID,
		leaveRequest.EmployeeID(),
		leaveRequestID,
		cmd.Reason,
	)

	if err := h.eventPublisher.Publish(ctx, event, tenantID, cmd.ActorID); err != nil {
		return nil, fmt.Errorf("publishing event: %w", err)
	}

	return &CancelLeaveRequestResult{
		LeaveRequestID: leaveRequestID.String(),
		Status:         string(leaveRequest.Status()),
	}, nil
}
