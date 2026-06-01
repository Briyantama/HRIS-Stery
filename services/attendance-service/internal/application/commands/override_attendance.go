package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hris-stery/hris-stery/services/attendance-service/internal/domain"
)

// OverrideAttendanceCommand represents the command to override an attendance record.
type OverrideAttendanceCommand struct {
	TenantID   string
	EmployeeID string
	Date       time.Time
	CheckInAt  *time.Time
	CheckOutAt *time.Time
	Status     string
	Reason     string
	ActorID    string
}

// OverrideAttendanceHandler handles the override attendance command.
type OverrideAttendanceHandler struct {
	attendanceRepo domain.AttendanceRepository
	eventPublisher domain.EventPublisher
}

// NewOverrideAttendanceHandler creates a new override attendance handler.
func NewOverrideAttendanceHandler(
	attendanceRepo domain.AttendanceRepository,
	eventPublisher domain.EventPublisher,
) *OverrideAttendanceHandler {
	return &OverrideAttendanceHandler{
		attendanceRepo: attendanceRepo,
		eventPublisher: eventPublisher,
	}
}

// OverrideAttendanceResult is the result of a successful override.
type OverrideAttendanceResult struct {
	AttendanceID string
	EmployeeID   string
	TenantID     string
	Status       string
	Reason       string
}

// Handle processes the override attendance command.
func (h *OverrideAttendanceHandler) Handle(ctx context.Context, cmd OverrideAttendanceCommand) (*OverrideAttendanceResult, error) {
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

	status := domain.AttendanceStatus(cmd.Status)
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid status: %s", cmd.Status)
	}

	// Fetch or create attendance record
	attendance, err := h.attendanceRepo.GetByEmployeeAndDate(ctx, tenantID, employeeID, cmd.Date)
	if err != nil && err != domain.ErrNotFound {
		return nil, fmt.Errorf("fetching attendance record: %w", err)
	}

	if attendance == nil {
		// Create a new attendance record if it doesn't exist
		attendance, err = domain.NewAttendanceRecord(tenantID, employeeID, cmd.Date, cmd.ActorID)
		if err != nil {
			return nil, fmt.Errorf("creating attendance record: %w", err)
		}
	}

	// Apply the override on the domain aggregate
	if err := attendance.Override(cmd.CheckInAt, cmd.CheckOutAt, status, cmd.Reason); err != nil {
		return nil, fmt.Errorf("override failed: %w", err)
	}

	// Persist the updated or created record
	if attendance.ID().IsZero() {
		// New record
		if err := h.attendanceRepo.Create(ctx, attendance); err != nil {
			return nil, fmt.Errorf("saving attendance record: %w", err)
		}
	} else {
		// Existing record update
		if err := h.attendanceRepo.Update(ctx, attendance); err != nil {
			return nil, fmt.Errorf("updating attendance record: %w", err)
		}
	}

	// Publish override event
	eventID := uuid.New().String()
	event := domain.NewAttendanceOverriddenEvent(
		eventID,
		tenantID,
		employeeID,
		attendance.ID(),
		status,
		cmd.CheckInAt,
		cmd.CheckOutAt,
		cmd.Reason,
	)

	if err := h.eventPublisher.Publish(ctx, event, tenantID, cmd.ActorID); err != nil {
		return nil, fmt.Errorf("publishing override event: %w", err)
	}

	return &OverrideAttendanceResult{
		AttendanceID: attendance.ID().String(),
		EmployeeID:   employeeID.String(),
		TenantID:     tenantID.String(),
		Status:       string(status),
		Reason:       cmd.Reason,
	}, nil
}
