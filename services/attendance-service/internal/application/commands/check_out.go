package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/domain"
)

// CheckOutCommand represents the command to check an employee out.
type CheckOutCommand struct {
	TenantID   string
	EmployeeID string
	Date       time.Time
	CheckOutAt time.Time
	Latitude   *float64
	Longitude  *float64
	Notes      string
	ActorID    string
}

// CheckOutHandler handles the check-out command.
type CheckOutHandler struct {
	attendanceRepo domain.AttendanceRepository
	eventPublisher domain.EventPublisher
}

// NewCheckOutHandler creates a new check-out handler.
func NewCheckOutHandler(
	attendanceRepo domain.AttendanceRepository,
	eventPublisher domain.EventPublisher,
) *CheckOutHandler {
	return &CheckOutHandler{
		attendanceRepo: attendanceRepo,
		eventPublisher: eventPublisher,
	}
}

// CheckOutResult is the result of a successful check-out.
type CheckOutResult struct {
	AttendanceID      string
	EmployeeID        string
	TenantID          string
	CheckOutAt        time.Time
	WorkDurationMins  int
	Status            string
}

// Handle processes the check-out command.
func (h *CheckOutHandler) Handle(ctx context.Context, cmd CheckOutCommand) (*CheckOutResult, error) {
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

	// Fetch the attendance record for this employee on this date
	attendance, err := h.attendanceRepo.GetByEmployeeAndDate(ctx, tenantID, employeeID, cmd.Date)
	if err != nil {
		return nil, fmt.Errorf("fetching attendance record: %w", err)
	}

	if attendance == nil {
		return nil, fmt.Errorf("no attendance record found for employee on this date")
	}

	// Perform check-out on the domain aggregate
	if err := attendance.CheckOut(cmd.CheckOutAt, cmd.Latitude, cmd.Longitude, cmd.Notes); err != nil {
		return nil, fmt.Errorf("check out failed: %w", err)
	}

	// Persist the updated record
	if err := h.attendanceRepo.Update(ctx, attendance); err != nil {
		return nil, fmt.Errorf("updating attendance record: %w", err)
	}

	// Publish check-out event
	eventID := uuid.New().String()
	workDurationMins := 0
	if attendance.WorkDurationMins() != nil {
		workDurationMins = *attendance.WorkDurationMins()
	}

	event := domain.NewAttendanceCheckedOutEvent(
		eventID,
		tenantID,
		employeeID,
		attendance.ID(),
		cmd.CheckOutAt,
		cmd.Latitude,
		cmd.Longitude,
		workDurationMins,
	)

	if err := h.eventPublisher.Publish(ctx, event, tenantID, cmd.ActorID); err != nil {
		return nil, fmt.Errorf("publishing check-out event: %w", err)
	}

	return &CheckOutResult{
		AttendanceID:     attendance.ID().String(),
		EmployeeID:       employeeID.String(),
		TenantID:         tenantID.String(),
		CheckOutAt:       cmd.CheckOutAt,
		WorkDurationMins: workDurationMins,
		Status:           string(attendance.Status()),
	}, nil
}
