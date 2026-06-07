package commands

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/domain"
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
	leaveBalanceRepo domain.LeaveBalanceRepository
	eventPublisher   domain.EventPublisher
}

// NewApproveLeaveRequestHandler creates a new approve leave request handler.
func NewApproveLeaveRequestHandler(
	leaveRequestRepo domain.LeaveRequestRepository,
	leaveBalanceRepo domain.LeaveBalanceRepository,
	eventPublisher domain.EventPublisher,
) *ApproveLeaveRequestHandler {
	return &ApproveLeaveRequestHandler{
		leaveRequestRepo: leaveRequestRepo,
		leaveBalanceRepo: leaveBalanceRepo,
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
		return nil, fmt.Errorf("leave request not found")
	}

	// Approve the request (enforces state machine invariants)
	if err := leaveRequest.Approve(approvedByID); err != nil {
		return nil, fmt.Errorf("approving leave request: %w", err)
	}

	// Update balance: move from pending to used
	year := leaveRequest.StartDate().Year()
	daysFloat := float64(leaveRequest.DaysCount())

	// Get balance first (separate read transaction is acceptable)
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

	// Update balance in memory: remove from pending and add to used
	if err := balance.RemovePending(daysFloat); err != nil {
		return nil, fmt.Errorf("removing pending days: %w", err)
	}

	if err := balance.AddUsed(daysFloat); err != nil {
		return nil, fmt.Errorf("adding used days: %w", err)
	}

	// Approve request and save balance atomically in a single transaction.
	// If either operation fails, both roll back.
	if err := h.leaveRequestRepo.ApproveAndUpdateBalanceAtomically(ctx, leaveRequest, balance, h.leaveBalanceRepo); err != nil {
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
