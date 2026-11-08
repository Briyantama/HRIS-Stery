package commands

import (
	"context"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/leave/domain"
	"github.com/google/uuid"
)

// ApproveLeaveRequestCommand represents the command to approve a leave request.
type ApproveLeaveRequestCommand struct {
	TenantID       string
	LeaveRequestID string
	ApprovedByID   string
	ActorID        string
}

// ApproveLeaveRequestHandler handles the approve leave request command.
type ApproveLeaveRequestHandler struct {
	leaveRequestRepo domain.LeaveRequestRepository
	eventPublisher   domain.EventPublisher
}

// NewApproveLeaveRequestHandler creates a new approve leave request handler.
func NewApproveLeaveRequestHandler(
	leaveRequestRepo domain.LeaveRequestRepository,
	eventPublisher domain.EventPublisher,
) *ApproveLeaveRequestHandler {
	return &ApproveLeaveRequestHandler{
		leaveRequestRepo: leaveRequestRepo,
		eventPublisher:   eventPublisher,
	}
}

// ApproveLeaveRequestResult is the result of a successful approval.
type ApproveLeaveRequestResult struct {
	LeaveRequestID string
	Status         string
}

// Handle processes the approve leave request command.
func (h *ApproveLeaveRequestHandler) Handle(ctx context.Context, cmd ApproveLeaveRequestCommand) (*ApproveLeaveRequestResult, error) {
	// Validate inputs
	tenantID, err := domain.NewTenantID(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	leaveRequestID, err := domain.NewLeaveRequestID(cmd.LeaveRequestID)
	if err != nil {
		return nil, fmt.Errorf("invalid leave request id: %w", err)
	}

	approvedByID, err := domain.NewEmployeeID(cmd.ApprovedByID)
	if err != nil {
		return nil, fmt.Errorf("invalid approver id: %w", err)
	}

	// Fetch the leave request
	leaveRequest, err := h.leaveRequestRepo.GetByID(ctx, tenantID, leaveRequestID)
	if err != nil {
		return nil, fmt.Errorf("fetching leave request: %w", err)
	}

	if leaveRequest == nil {
		return nil, fmt.Errorf("leave request: %w", domain.ErrNotFound)
	}

	// Approve the request (enforces state machine invariants and no self-approval)
	if err := leaveRequest.Approve(approvedByID); err != nil {
		return nil, fmt.Errorf("approving leave request: %w", err)
	}

	// Move the days from pending to used; the status write only succeeds while the request is still
	// PENDING and the balance change is relative, both in one transaction.
	days := float64(leaveRequest.DaysCount())
	delta := domain.BalanceDelta{Used: days, Pending: -days}
	if err := h.leaveRequestRepo.TransitionAndUpdateBalanceAtomically(ctx, leaveRequest, domain.StatusPending, delta); err != nil {
		return nil, fmt.Errorf("approving leave request: %w", err)
	}

	// Publish domain event
	eventID := uuid.New().String()
	event := domain.NewLeaveApprovedEvent(
		eventID,
		tenantID,
		leaveRequest.EmployeeID(),
		leaveRequestID,
		approvedByID,
	)

	if err := h.eventPublisher.Publish(ctx, event, tenantID, cmd.ActorID); err != nil {
		return nil, fmt.Errorf("publishing event: %w", err)
	}

	return &ApproveLeaveRequestResult{
		LeaveRequestID: leaveRequestID.String(),
		Status:         string(leaveRequest.Status()),
	}, nil
}
