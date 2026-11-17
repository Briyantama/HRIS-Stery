package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/attendance/domain"
)

// ListAttendanceQuery retrieves attendance records for a tenant with filters.
type ListAttendanceQuery struct {
	TenantID   string
	EmployeeID *string
	DateFrom   *time.Time
	DateTo     *time.Time
	Status     *string
	Offset     int
	Limit      int
}

// ListAttendanceHandler handles the list attendance query.
type ListAttendanceHandler struct {
	attendanceRepo domain.AttendanceRepository
}

// NewListAttendanceHandler creates a new list attendance handler.
func NewListAttendanceHandler(attendanceRepo domain.AttendanceRepository) *ListAttendanceHandler {
	return &ListAttendanceHandler{
		attendanceRepo: attendanceRepo,
	}
}

// ListAttendanceResult contains the list of attendance records.
type ListAttendanceResult struct {
	Records []*AttendanceDTO `json:"records"`
	Total   int              `json:"total"`
	Offset  int              `json:"offset"`
	Limit   int              `json:"limit"`
}

// Handle retrieves attendance records for a tenant.
func (h *ListAttendanceHandler) Handle(ctx context.Context, query ListAttendanceQuery) (*ListAttendanceResult, error) {
	tenantID, err := domain.NewTenantID(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	// Build filters
	filters := domain.AttendanceFilters{
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

	if query.DateFrom != nil {
		filters.DateFrom = query.DateFrom
	}

	if query.DateTo != nil {
		filters.DateTo = query.DateTo
	}

	if query.Status != nil {
		status := domain.AttendanceStatus(*query.Status)
		if !status.IsValid() {
			return nil, fmt.Errorf("invalid status %q: %w", *query.Status, domain.ErrInvalidStatus)
		}
		filters.Status = &status
	}

	// Fetch records from repository
	records, err := h.attendanceRepo.ListByTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("fetching attendance records: %w", err)
	}

	// Convert to DTOs
	dtos := make([]*AttendanceDTO, len(records))
	for i, record := range records {
		dtos[i] = domainToDTO(record)
	}

	total, err := h.attendanceRepo.CountByTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("counting attendance records: %w", err)
	}

	return &ListAttendanceResult{
		Records: dtos,
		Total:   total,
		Offset:  query.Offset,
		Limit:   query.Limit,
	}, nil
}
