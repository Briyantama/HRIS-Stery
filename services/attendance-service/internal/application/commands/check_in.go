package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/domain"
)

// CheckInCommand represents the command to check an employee in.
type CheckInCommand struct {
	TenantID   string
	EmployeeID string
	Date       time.Time
	CheckInAt  time.Time
	Latitude   *float64
	Longitude  *float64
	Notes      string
	ActorID    string
}

// CheckInHandler handles the check-in command.
type CheckInHandler struct {
	attendanceRepo domain.AttendanceRepository
	eventPublisher domain.EventPublisher
}

// NewCheckInHandler creates a new check-in handler.
func NewCheckInHandler(
	attendanceRepo domain.AttendanceRepository,
	eventPublisher domain.EventPublisher,
) *CheckInHandler {
	return &CheckInHandler{
		attendanceRepo: attendanceRepo,
		eventPublisher: eventPublisher,
	}
}

// CheckInResult is the result of a successful check-in.
type CheckInResult struct {
	AttendanceID string
	EmployeeID   string
	TenantID     string
	CheckInAt    time.Time
	Status       string
}

// Handle processes the check-in command.
func (h *CheckInHandler) Handle(ctx context.Context, cmd CheckInCommand) (*CheckInResult, error) {
	// Validate inputs
	tenantID, err := domain.NewTenantID(cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	employeeID, err := domain.NewEmployeeID(cmd.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee id: %w", err)
	}

	if cmd.Date.IsZero() {
		return nil, fmt.Errorf("invalid date: %w", domain.ErrInvalidDate)
	}

	// Check if employee already has a check-in record for this date
	existingRecord, err := h.attendanceRepo.GetByEmployeeAndDate(ctx, tenantID, employeeID, cmd.Date)
	if err != nil && err != domain.ErrNotFound {
		return nil, fmt.Errorf("checking existing record: %w", err)
	}

	var attendance *domain.AttendanceRecord
	if existingRecord != nil {
		// Employee already has a record for this date; attempt to check in
		attendance = existingRecord
		if err := attendance.CheckIn(cmd.CheckInAt, cmd.Latitude, cmd.Longitude, cmd.Notes); err != nil {
			return nil, fmt.Errorf("check in failed: %w", err)
		}
		if err := h.attendanceRepo.Update(ctx, attendance); err != nil {
			return nil, fmt.Errorf("updating attendance record: %w", err)
		}
	} else {
		// Create a new attendance record for this date
		attendance, err = domain.NewAttendanceRecord(tenantID, employeeID, cmd.Date, cmd.ActorID)
		if err != nil {
			return nil, fmt.Errorf("creating attendance record: %w", err)
		}

		if err := attendance.CheckIn(cmd.CheckInAt, cmd.Latitude, cmd.Longitude, cmd.Notes); err != nil {
			return nil, fmt.Errorf("check in failed: %w", err)
		}

		if err := h.attendanceRepo.Create(ctx, attendance); err != nil {
			return nil, fmt.Errorf("saving attendance record: %w", err)
		}
	}

	// Publish check-in event
	eventID := uuid.New().String()
	event := domain.NewAttendanceCheckedInEvent(
		eventID,
		tenantID,
		employeeID,
		attendance.ID(),
		cmd.CheckInAt,
		cmd.Latitude,
		cmd.Longitude,
	)

	if err := h.eventPublisher.Publish(ctx, event, tenantID, cmd.ActorID); err != nil {
		return nil, fmt.Errorf("publishing check-in event: %w", err)
	}

	return &CheckInResult{
		AttendanceID: attendance.ID().String(),
		EmployeeID:   employeeID.String(),
		TenantID:     tenantID.String(),
		CheckInAt:    cmd.CheckInAt,
		Status:       string(attendance.Status()),
	}, nil
}
