package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/attendance-service/internal/domain"
)

// GetAttendanceQuery retrieves a single attendance record by ID.
type GetAttendanceQuery struct {
	TenantID     string
	AttendanceID string
}

// GetAttendanceHandler handles the get attendance query.
type GetAttendanceHandler struct {
	attendanceRepo domain.AttendanceRepository
}

// NewGetAttendanceHandler creates a new get attendance handler.
func NewGetAttendanceHandler(attendanceRepo domain.AttendanceRepository) *GetAttendanceHandler {
	return &GetAttendanceHandler{
		attendanceRepo: attendanceRepo,
	}
}

// AttendanceDTO is the data transfer object for attendance records.
type AttendanceDTO struct {
	ID               string  `json:"id"`
	TenantID         string  `json:"tenant_id"`
	EmployeeID       string  `json:"employee_id"`
	Date             string  `json:"date"`
	CheckInAt        *string `json:"check_in_at"`
	CheckOutAt       *string `json:"check_out_at"`
	CheckInLatitude  *float64 `json:"check_in_latitude"`
	CheckInLongitude *float64 `json:"check_in_longitude"`
	CheckOutLatitude  *float64 `json:"check_out_latitude"`
	CheckOutLongitude *float64 `json:"check_out_longitude"`
	Status           string  `json:"status"`
	WorkDurationMins *int    `json:"work_duration_mins"`
	Notes            string  `json:"notes"`
	CreatedBy        string  `json:"created_by"`
	IsOverride       bool    `json:"is_override"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// Handle retrieves an attendance record.
func (h *GetAttendanceHandler) Handle(ctx context.Context, query GetAttendanceQuery) (*AttendanceDTO, error) {
	tenantID, err := domain.NewTenantID(query.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	attendanceID, err := domain.NewAttendanceID(query.AttendanceID)
	if err != nil {
		return nil, fmt.Errorf("invalid attendance id: %w", err)
	}

	attendance, err := h.attendanceRepo.GetByID(ctx, tenantID, attendanceID)
	if err != nil {
		return nil, fmt.Errorf("fetching attendance: %w", err)
	}

	if attendance == nil {
		return nil, fmt.Errorf("attendance record not found")
	}

	return domainToDTO(attendance), nil
}

// Helper function to convert domain object to DTO
func domainToDTO(record *domain.AttendanceRecord) *AttendanceDTO {
	var checkInAtStr *string
	if record.CheckInAt() != nil {
		s := record.CheckInAt().Format("2006-01-02T15:04:05Z07:00")
		checkInAtStr = &s
	}

	var checkOutAtStr *string
	if record.CheckOutAt() != nil {
		s := record.CheckOutAt().Format("2006-01-02T15:04:05Z07:00")
		checkOutAtStr = &s
	}

	lat, lon := record.CheckInLocation()
	outLat, outLon := record.CheckOutLocation()

	return &AttendanceDTO{
		ID:                record.ID().String(),
		TenantID:          record.TenantID().String(),
		EmployeeID:        record.EmployeeID().String(),
		Date:              record.Date().Format("2006-01-02"),
		CheckInAt:         checkInAtStr,
		CheckOutAt:        checkOutAtStr,
		CheckInLatitude:   lat,
		CheckInLongitude:  lon,
		CheckOutLatitude:  outLat,
		CheckOutLongitude: outLon,
		Status:            string(record.Status()),
		WorkDurationMins:  record.WorkDurationMins(),
		Notes:             record.Notes(),
		CreatedBy:         record.CreatedBy(),
		IsOverride:        record.IsOverride(),
		CreatedAt:         record.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         record.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
