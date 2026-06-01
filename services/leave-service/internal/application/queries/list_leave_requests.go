package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/hris-stery/hris-stery/services/leave-service/internal/domain"
)

// ListLeaveRequestsQuery retrieves leave requests for a tenant with filters.
type ListLeaveRequestsQuery struct {
	TenantID     string
	EmployeeID   *string
	Status       *string
	ApprovedByID *string
	Year         *int
	Offset       int
	Limit        int
}

// ListLeaveRequestsHandler handles the list leave requests query.
type ListLeaveRequestsHandler struct {
	leaveRequestRepo domain.LeaveRequestRepository
}

// NewListLeaveRequestsHandler creates a new list leave requests handler.
func NewListLeaveRequestsHandler(leaveRequestRepo domain.LeaveRequestRepository) *ListLeaveRequestsHandler {
	return &ListLeaveRequestsHandler{
		leaveRequestRepo: leaveRequestRepo,
	}
}

// ListLeaveRequestsResult contains the list of leave requests.
type ListLeaveRequestsResult struct {
	Records []*LeaveRequestDTO `json:"records"`
	Total   int                `json:"total"`
	Offset  int                `json:"offset"`
	Limit   int                `json:"limit"`
}

// Handle retrieves leave requests for a tenant.
func (h *ListLeaveRequestsHandler) Handle(ctx context.Context, query ListLeaveRequestsQuery) (*ListLeaveRequestsResult, error) {
	tenantID, err := domain.NewTenantID(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	// Build filters
	filters := domain.LeaveFilters{
		Offset: query.Offset,
		Limit:  query.Limit,
	}

	if query.EmployeeID != nil {
		empID, err := domain.NewEmployeeID(*query.EmployeeID)
		if err != nil {
			return nil, fmt.Errorf("invalid employee id: %w", err)
		}
		filters.EmployeeID = &empID
	}

	if query.ApprovedByID != nil {
		appID, err := domain.NewEmployeeID(*query.ApprovedByID)
		if err != nil {
			return nil, fmt.Errorf("invalid approver id: %w", err)
		}
		filters.ApprovedByID = &appID
	}

	if query.Status != nil {
		status := domain.LeaveStatus(*query.Status)
		if !status.IsValid() {
			return nil, fmt.Errorf("invalid status: %s", *query.Status)
		}
		filters.Status = &status
	}

	if query.Year != nil {
		now := time.Now()
		year := *query.Year
		filters.DateFrom = &time.Time{}
		*filters.DateFrom = time.Date(year, time.January, 1, 0, 0, 0, 0, now.Location())
		filters.DateTo = &time.Time{}
		*filters.DateTo = time.Date(year, time.December, 31, 23, 59, 59, 999999999, now.Location())
	}

	// Fetch records from repository
	records, err := h.leaveRequestRepo.ListByTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("fetching leave requests: %w", err)
	}

	// Convert to DTOs
	dtos := make([]*LeaveRequestDTO, len(records))
	for i, record := range records {
		dtos[i] = domainToDTO(record)
	}

	return &ListLeaveRequestsResult{
		Records: dtos,
		Total:   len(dtos),
		Offset:  query.Offset,
		Limit:   query.Limit,
	}, nil
}
