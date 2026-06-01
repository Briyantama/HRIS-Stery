package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/leave-service/internal/domain"
)

// GetLeaveRequestQuery retrieves a single leave request by ID.
type GetLeaveRequestQuery struct {
	TenantID       string
	LeaveRequestID string
}

// GetLeaveRequestHandler handles the get leave request query.
type GetLeaveRequestHandler struct {
	leaveRequestRepo domain.LeaveRequestRepository
}

// NewGetLeaveRequestHandler creates a new get leave request handler.
func NewGetLeaveRequestHandler(leaveRequestRepo domain.LeaveRequestRepository) *GetLeaveRequestHandler {
	return &GetLeaveRequestHandler{
		leaveRequestRepo: leaveRequestRepo,
	}
}

// LeaveRequestDTO is the data transfer object for leave requests.
type LeaveRequestDTO struct {
	ID              string  `json:"id"`
	TenantID        string  `json:"tenant_id"`
	EmployeeID      string  `json:"employee_id"`
	LeaveTypeID     string  `json:"leave_type_id"`
	StartDate       string  `json:"start_date"`
	EndDate         string  `json:"end_date"`
	DaysCount       int     `json:"days_count"`
	Status          string  `json:"status"`
	Reason          string  `json:"reason"`
	RejectionReason string  `json:"rejection_reason,omitempty"`
	ApprovedByID    *string `json:"approved_by_id,omitempty"`
	ApprovedAt      *string `json:"approved_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// Handle retrieves a leave request.
func (h *GetLeaveRequestHandler) Handle(ctx context.Context, query GetLeaveRequestQuery) (*LeaveRequestDTO, error) {
	tenantID, err := domain.NewTenantID(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	leaveRequestID, err := domain.NewLeaveRequestID(query.LeaveRequestID)
	if err != nil {
		return nil, fmt.Errorf("invalid leave request id: %w", err)
	}

	leaveRequest, err := h.leaveRequestRepo.GetByID(ctx, tenantID, leaveRequestID)
	if err != nil {
		return nil, fmt.Errorf("fetching leave request: %w", err)
	}

	if leaveRequest == nil {
		return nil, fmt.Errorf("leave request not found")
	}

	return domainToDTO(leaveRequest), nil
}

// Helper function to convert domain object to DTO
func domainToDTO(req *domain.LeaveRequest) *LeaveRequestDTO {
	var approvedByIDStr *string
	if req.ApprovedByID() != nil {
		idStr := req.ApprovedByID().String()
		approvedByIDStr = &idStr
	}

	var approvedAtStr *string
	if req.ApprovedAt() != nil {
		atStr := req.ApprovedAt().Format("2006-01-02T15:04:05Z07:00")
		approvedAtStr = &atStr
	}

	return &LeaveRequestDTO{
		ID:              req.ID().String(),
		TenantID:        req.TenantID().String(),
		EmployeeID:      req.EmployeeID().String(),
		LeaveTypeID:     req.LeaveTypeID().String(),
		StartDate:       req.StartDate().Format("2006-01-02"),
		EndDate:         req.EndDate().Format("2006-01-02"),
		DaysCount:       req.DaysCount(),
		Status:          string(req.Status()),
		Reason:          req.Reason(),
		RejectionReason: req.RejectionReason(),
		ApprovedByID:    approvedByIDStr,
		ApprovedAt:      approvedAtStr,
		CreatedAt:       req.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       req.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
