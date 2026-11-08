package commands

import (
	"context"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/leave/domain"
	"github.com/google/uuid"
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
	eventPublisher   domain.EventPublisher
}

// NewRejectLeaveRequestHandler creates a new reject leave request handler.
func NewRejectLeaveRequestHandler(
	leaveRequestRepo domain.LeaveRequestRepository,
	eventPublisher domain.EventPublisher,
) *RejectLeaveRequestHandler {
	return &RejectLeaveRequestHandler{
		leaveRequestRepo: leaveRequestRepo,
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

	rejecterID, err := domain.NewEmployeeID(cmd.ActorID)
	if err != nil {
		return nil, fmt.Errorf("invalid actor id: %w", err)
	}

	// Fetch the leave request
	leaveRequest, err := h.leaveRequestRepo.GetByID(ctx, tenantID, leaveRequestID)
	if err != nil {
		return nil, fmt.Errorf("fetching leave request: %w", err)
	}

	if leaveRequest == nil {
		return nil, fmt.Errorf("leave request: %w", domain.ErrNotFound)
	}

	// Reject the request (enforces state machine invariants and no self-rejection)
	if err := leaveRequest.Reject(rejecterID, cmd.Reason); err != nil {
		return nil, fmt.Errorf("rejecting leave request: %w", err)
	}

	// Release the pending days; the status write only succeeds while the request is still PENDING and
	// the balance change is relative, both in one transaction.
	delta := domain.BalanceDelta{Pending: -float64(leaveRequest.DaysCount())}
	if err := h.leaveRequestRepo.TransitionAndUpdateBalanceAtomically(ctx, leaveRequest, domain.StatusPending, delta); err != nil {
		return nil, fmt.Errorf("rejecting leave request: %w", err)
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
