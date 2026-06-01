package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/leave-service/internal/domain"
)

// GetLeaveBalanceQuery retrieves leave balances for an employee in a specific year.
type GetLeaveBalanceQuery struct {
	TenantID   string
	EmployeeID string
	Year       int
}

// GetLeaveBalanceHandler handles the get leave balance query.
type GetLeaveBalanceHandler struct {
	leaveBalanceRepo domain.LeaveBalanceRepository
}

// NewGetLeaveBalanceHandler creates a new get leave balance handler.
func NewGetLeaveBalanceHandler(leaveBalanceRepo domain.LeaveBalanceRepository) *GetLeaveBalanceHandler {
	return &GetLeaveBalanceHandler{
		leaveBalanceRepo: leaveBalanceRepo,
	}
}

// LeaveBalanceDetailDTO represents a single leave type's balance.
type LeaveBalanceDetailDTO struct {
	LeaveTypeID   string  `json:"leave_type_id"`
	EntitledDays  float64 `json:"entitled_days"`
	UsedDays      float64 `json:"used_days"`
	PendingDays   float64 `json:"pending_days"`
	RemainingDays float64 `json:"remaining_days"`
}

// LeaveBalanceResult contains the aggregate leave balance for an employee.
type LeaveBalanceResult struct {
	EmployeeID string                   `json:"employee_id"`
	TenantID   string                   `json:"tenant_id"`
	Year       int                      `json:"year"`
	ByType     []*LeaveBalanceDetailDTO `json:"by_type"`
}

// Handle retrieves leave balance for an employee in a year.
func (h *GetLeaveBalanceHandler) Handle(ctx context.Context, query GetLeaveBalanceQuery) (*LeaveBalanceResult, error) {
	tenantID, err := domain.NewTenantID(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	employeeID, err := domain.NewEmployeeID(query.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee id: %w", err)
	}

	// Fetch all balances for this employee in the year
	balances, err := h.leaveBalanceRepo.ListByEmployee(ctx, tenantID, employeeID, query.Year)
	if err != nil {
		return nil, fmt.Errorf("fetching leave balances: %w", err)
	}

	// Convert to DTOs
	details := make([]*LeaveBalanceDetailDTO, len(balances))
	for i, balance := range balances {
		details[i] = &LeaveBalanceDetailDTO{
			LeaveTypeID:   balance.LeaveTypeID().String(),
			EntitledDays:  balance.EntitledDays(),
			UsedDays:      balance.UsedDays(),
			PendingDays:   balance.PendingDays(),
			RemainingDays: balance.RemainingDays(),
		}
	}

	return &LeaveBalanceResult{
		EmployeeID: employeeID.String(),
		TenantID:   tenantID.String(),
		Year:       query.Year,
		ByType:     details,
	}, nil
}
