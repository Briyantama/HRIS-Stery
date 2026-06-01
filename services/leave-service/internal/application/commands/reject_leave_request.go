package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/domain"
)

// RejectLeaveRequestCommand represents the command to reject a leave request.
type RejectLeaveRequestCommand struct {
	TenantID       string
	LeaveRequestID string
	Reason         string
	ActorID        string
}

// RejectLeaveRequestHandler handles the reject leave request command.
type RejectLeaveRequestHandler struct {
	leaveRequestRepo domain.LeaveRequestRepository
	leaveBalanceRepo domain.LeaveBalanceRepository
	eventPublisher   domain.EventPublisher
}

// NewRejectLeaveRequestHandler creates a new reject leave request handler.
func NewRejectLeaveRequestHandler(
	leaveRequestRepo domain.LeaveRequestRepository,
	leaveBalanceRepo domain.LeaveBalanceRepository,
	eventPublisher domain.EventPublisher,
) *RejectLeaveRequestHandler {
	return &RejectLeaveRequestHandler{
		leaveRequestRepo: leaveRequestRepo,
		leaveBalanceRepo: leaveBalanceRepo,
		eventPublisher:   eventPublisher,
	}
}

// RejectLeaveRequestResult is the result of a successful rejection.
type RejectLeaveRequestResult struct {
	LeaveRequestID string
	Status         string
}

// Handle processes the reject leave request command.
func (h *RejectLeaveRequestHandler) Handle(ctx context.Context, cmd RejectLeaveRequestCommand) (*RejectLeaveRequestResult, error) {
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

	// Reject the request (enforces state machine invariants)
	if err := leaveRequest.Reject(cmd.Reason); err != nil {
		return nil, fmt.Errorf("rejecting leave request: %w", err)
	}

	// Update balance: remove from pending
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

	// Remove from pending
	daysFloat := float64(leaveRequest.DaysCount())
	if err := balance.RemovePending(daysFloat); err != nil {
		return nil, fmt.Errorf("removing pending days: %w", err)
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
	event := domain.NewLeaveRejectedEvent(
		eventID,
		tenantID,
		leaveRequest.EmployeeID(),
		leaveRequestID,
		cmd.Reason,
	)

	if err := h.eventPublisher.Publish(ctx, event, tenantID, cmd.ActorID); err != nil {
		return nil, fmt.Errorf("publishing event: %w", err)
	}

	return &RejectLeaveRequestResult{
		LeaveRequestID: leaveRequestID.String(),
		Status:         string(leaveRequest.Status()),
	}, nil
}
