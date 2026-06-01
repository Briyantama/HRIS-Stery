package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/leave-service/internal/domain"
)

// ListLeaveTypesQuery retrieves all leave types for a tenant.
type ListLeaveTypesQuery struct {
	TenantID string
}

// ListLeaveTypesHandler handles the list leave types query.
type ListLeaveTypesHandler struct {
	leaveTypeRepo domain.LeaveTypeRepository
}

// NewListLeaveTypesHandler creates a new list leave types handler.
func NewListLeaveTypesHandler(leaveTypeRepo domain.LeaveTypeRepository) *ListLeaveTypesHandler {
	return &ListLeaveTypesHandler{
		leaveTypeRepo: leaveTypeRepo,
	}
}

// LeaveTypeDTO represents a leave type for the API.
type LeaveTypeDTO struct {
	ID               string  `json:"id"`
	Code             string  `json:"code"`
	Name             string  `json:"name"`
	MaxDaysPerYear   float64 `json:"max_days_per_year"`
	RequiresDocument bool    `json:"requires_document"`
	IsPaid           bool    `json:"is_paid"`
	IsActive         bool    `json:"is_active"`
}

// ListLeaveTypesResult contains the list of leave types.
type ListLeaveTypesResult struct {
	LeaveTypes []*LeaveTypeDTO `json:"leave_types"`
	Total      int             `json:"total"`
}

// Handle retrieves leave types for a tenant.
func (h *ListLeaveTypesHandler) Handle(ctx context.Context, query ListLeaveTypesQuery) (*ListLeaveTypesResult, error) {
	tenantID, err := domain.NewTenantID(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	leaveTypes, err := h.leaveTypeRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("fetching leave types: %w", err)
	}

	// Convert to DTOs
	dtos := make([]*LeaveTypeDTO, len(leaveTypes))
	for i, lt := range leaveTypes {
		dtos[i] = &LeaveTypeDTO{
			ID:               lt.ID().String(),
			Code:             lt.Code(),
			Name:             lt.Name(),
			MaxDaysPerYear:   lt.MaxDaysPerYear(),
			RequiresDocument: lt.RequiresDocument(),
			IsPaid:           lt.IsPaid(),
			IsActive:         lt.IsActive(),
		}
	}

	return &ListLeaveTypesResult{
		LeaveTypes: dtos,
		Total:      len(dtos),
	}, nil
}
