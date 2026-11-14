package commands

import (
	"context"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/leave/domain"
	"uuid"
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
	eventPublisher   domain.EventPublisher
}

// NewCancelLeaveRequestHandler creates a new cancel leave request handler.
func NewCancelLeaveRequestHandler(
	leaveRequestRepo domain.LeaveRequestRepository,
	eventPublisher domain.EventPublisher,
) *CancelLeaveRequestHandler {
	return &CancelLeaveRequestHandler{
		leaveRequestRepo: leaveRequestRepo,
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
		return nil, fmt.Errorf("leave request: %w", domain.ErrNotFound)
	}

	// Store original status for balance update logic
	originalStatus := leaveRequest.Status()

	// Cancel the request (enforces state machine invariants)
	if err := leaveRequest.Cancel(cmd.Reason); err != nil {
		return nil, fmt.Errorf("cancelling leave request: %w", err)
	}

	// Give the days back from where they were held; the status write only succeeds while the request
	// still has the status we read and the balance change is relative, both in one transaction.
	days := float64(leaveRequest.DaysCount())
	delta := domain.BalanceDelta{Pending: -days}
	if originalStatus == domain.StatusApproved {
		delta = domain.BalanceDelta{Used: -days}
	}
	if err := h.leaveRequestRepo.TransitionAndUpdateBalanceAtomically(ctx, leaveRequest, originalStatus, delta); err != nil {
		return nil, fmt.Errorf("cancelling leave request: %w", err)
	}

	// Publish domain event
	eventID := uuid.NewV7().String()
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
