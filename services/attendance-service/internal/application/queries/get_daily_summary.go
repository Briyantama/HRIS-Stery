package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/hris-stery/hris-stery/services/attendance-service/internal/domain"
)

// GetDailySummaryQuery retrieves a summary of attendance for a specific date.
type GetDailySummaryQuery struct {
	TenantID string
	Date     time.Time
}

// GetDailySummaryHandler handles the get daily summary query.
type GetDailySummaryHandler struct {
	attendanceRepo domain.AttendanceRepository
}

// NewGetDailySummaryHandler creates a new get daily summary handler.
func NewGetDailySummaryHandler(attendanceRepo domain.AttendanceRepository) *GetDailySummaryHandler {
	return &GetDailySummaryHandler{
		attendanceRepo: attendanceRepo,
	}
}

// DailySummaryResult contains attendance summary for a specific date.
type DailySummaryResult struct {
	Date            string `json:"date"`
	TenantID        string `json:"tenant_id"`
	TotalEmployees  int    `json:"total_employees"`
	PresentCount    int    `json:"present_count"`
	AbsentCount     int    `json:"absent_count"`
	LateCount       int    `json:"late_count"`
	HalfDayCount    int    `json:"half_day_count"`
	OnLeaveCount    int    `json:"on_leave_count"`
	AverageWorkMins int    `json:"average_work_mins"`
}

// Handle retrieves a daily attendance summary.
func (h *GetDailySummaryHandler) Handle(ctx context.Context, query GetDailySummaryQuery) (*DailySummaryResult, error) {
	tenantID, err := domain.NewTenantID(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	if query.Date.IsZero() {
		return nil, fmt.Errorf("invalid date: %w", domain.ErrInvalidDate)
	}

	// Fetch all attendance records for this date within the tenant
	filters := domain.AttendanceFilters{
		DateFrom: &query.Date,
		DateTo:   &query.Date,
		Offset:   0,
		Limit:    10000, // Reasonable upper bound for daily attendance
	}

	records, err := h.attendanceRepo.ListByTenant(ctx, tenantID, filters)
	if err != nil {
		return nil, fmt.Errorf("fetching attendance records: %w", err)
	}

	// Summarize the records
	summary := &DailySummaryResult{
		Date:     query.Date.Format("2006-01-02"),
		TenantID: tenantID.String(),
	}

	totalWorkMins := 0
	recordsWithDuration := 0

	for _, record := range records {
		summary.TotalEmployees++

		switch record.Status() {
		case domain.StatusPresent:
			summary.PresentCount++
		case domain.StatusAbsent:
			summary.AbsentCount++
		case domain.StatusLate:
			summary.LateCount++
		case domain.StatusHalfDay:
			summary.HalfDayCount++
		case domain.StatusOnLeave:
			summary.OnLeaveCount++
		}

		if record.WorkDurationMins() != nil {
			totalWorkMins += *record.WorkDurationMins()
			recordsWithDuration++
		}
	}

	if recordsWithDuration > 0 {
		summary.AverageWorkMins = totalWorkMins / recordsWithDuration
	}

	return summary, nil
}
