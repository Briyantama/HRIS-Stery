package domain

import (
	"time"
)

// AttendanceStatus represents the attendance status on a given day.
type AttendanceStatus string

const (
	StatusPresent AttendanceStatus = "PRESENT"
	StatusLate    AttendanceStatus = "LATE"
	StatusAbsent  AttendanceStatus = "ABSENT"
	StatusHalfDay AttendanceStatus = "HALF_DAY"
	StatusOnLeave AttendanceStatus = "ON_LEAVE"
)

// IsValid returns true if the status is a valid AttendanceStatus.
func (s AttendanceStatus) IsValid() bool {
	switch s {
	case StatusPresent, StatusLate, StatusAbsent, StatusHalfDay, StatusOnLeave:
		return true
	default:
		return false
	}
}

// AttendanceRecord is the aggregate root for attendance records.
// It tracks check-in and check-out times for an employee on a specific date.
type AttendanceRecord struct {
	id                AttendanceID
	tenantID          TenantID
	employeeID        EmployeeID
	date              time.Time  // ISO 8601 date
	checkInAt         *time.Time // When employee checked in
	checkOutAt        *time.Time // When employee checked out
	checkInLatitude   *float64   // Optional geolocation
	checkInLongitude  *float64
	checkOutLatitude  *float64
	checkOutLongitude *float64
	status            AttendanceStatus
	workDurationMins  *int // Computed after check-out
	notes             string
	createdBy         string // "system" for auto check-in/out, user ID for admin override
	isOverride        bool   // True if admin manually overrode the record
	createdAt         time.Time
	updatedAt         time.Time
}

// NewAttendanceRecord creates a new attendance record for an employee on a given date.
func NewAttendanceRecord(
	tenantID TenantID,
	employeeID EmployeeID,
	date time.Time,
	createdBy string,
) (*AttendanceRecord, error) {
	if tenantID.IsZero() {
		return nil, ErrMissingRequiredField
	}
	if employeeID.IsZero() {
		return nil, ErrMissingRequiredField
	}
	if date.IsZero() {
		return nil, ErrInvalidDate
	}

	now := time.Now().UTC()
	return &AttendanceRecord{
		id:         GenerateAttendanceID(),
		tenantID:   tenantID,
		employeeID: employeeID,
		date:       date,
		status:     StatusAbsent, // Default until check-in
		createdBy:  createdBy,
		createdAt:  now,
		updatedAt:  now,
	}, nil
}

// RehydrateAttendanceRecord reconstructs an attendance record from persisted data.
func RehydrateAttendanceRecord(
	id AttendanceID,
	tenantID TenantID,
	employeeID EmployeeID,
	date time.Time,
	checkInAt *time.Time,
	checkOutAt *time.Time,
	checkInLatitude *float64,
	checkInLongitude *float64,
	checkOutLatitude *float64,
	checkOutLongitude *float64,
	status AttendanceStatus,
	workDurationMins *int,
	notes string,
	createdBy string,
	isOverride bool,
	createdAt time.Time,
	updatedAt time.Time,
) *AttendanceRecord {
	return &AttendanceRecord{
		id:                id,
		tenantID:          tenantID,
		employeeID:        employeeID,
		date:              date,
		checkInAt:         checkInAt,
		checkOutAt:        checkOutAt,
		checkInLatitude:   checkInLatitude,
		checkInLongitude:  checkInLongitude,
		checkOutLatitude:  checkOutLatitude,
		checkOutLongitude: checkOutLongitude,
		status:            status,
		workDurationMins:  workDurationMins,
		notes:             notes,
		createdBy:         createdBy,
		isOverride:        isOverride,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}
}

// CheckIn records a check-in time and location.
// Invariant: Cannot check in if already checked in (no duplicate active check-in).
func (a *AttendanceRecord) CheckIn(
	checkInAt time.Time,
	latitude *float64,
	longitude *float64,
	notes string,
) error {
	if a.checkInAt != nil {
		return ErrCannotCheckInTwiceSameDay
	}

	a.checkInAt = &checkInAt
	a.checkInLatitude = latitude
	a.checkInLongitude = longitude
	if notes != "" {
		a.notes = notes
	}
	a.updatedAt = time.Now().UTC()

	// Auto-set status to PRESENT on check-in
	a.status = StatusPresent

	return nil
}

// CheckOut records a check-out time and location.
// Invariant: Cannot check out before checking in.
func (a *AttendanceRecord) CheckOut(
	checkOutAt time.Time,
	latitude *float64,
	longitude *float64,
	notes string,
) error {
	if a.checkInAt == nil {
		return ErrCannotCheckOutBeforeCheckIn
	}
	if a.checkOutAt != nil {
		return ErrAlreadyCheckedOut
	}

	// Validate time range
	if checkOutAt.Before(*a.checkInAt) {
		return ErrInvalidTimeRange
	}

	a.checkOutAt = &checkOutAt
	a.checkOutLatitude = latitude
	a.checkOutLongitude = longitude

	// Append to notes if additional notes provided
	if notes != "" {
		if a.notes != "" {
			a.notes += "; " + notes
		} else {
			a.notes = notes
		}
	}

	// Compute work duration in minutes
	duration := checkOutAt.Sub(*a.checkInAt)
	durationMins := int(duration.Minutes())
	a.workDurationMins = &durationMins

	a.updatedAt = time.Now().UTC()

	return nil
}

// Override allows an administrator to manually override attendance for a specific date.
func (a *AttendanceRecord) Override(
	checkInAt *time.Time,
	checkOutAt *time.Time,
	status AttendanceStatus,
	reason string,
) error {
	if !status.IsValid() {
		return ErrInvalidDate
	}

	// Validate time range if both provided
	if checkInAt != nil && checkOutAt != nil && checkOutAt.Before(*checkInAt) {
		return ErrInvalidTimeRange
	}

	a.checkInAt = checkInAt
	a.checkOutAt = checkOutAt
	a.status = status
	a.isOverride = true

	// Compute work duration if both times provided
	if checkInAt != nil && checkOutAt != nil {
		duration := checkOutAt.Sub(*checkInAt)
		durationMins := int(duration.Minutes())
		a.workDurationMins = &durationMins
	}

	if reason != "" {
		a.notes = reason
	}

	a.updatedAt = time.Now().UTC()

	return nil
}

// Getters (all methods are for reading state, enforcing encapsulation)

func (a *AttendanceRecord) ID() AttendanceID {
	return a.id
}

func (a *AttendanceRecord) TenantID() TenantID {
	return a.tenantID
}

func (a *AttendanceRecord) EmployeeID() EmployeeID {
	return a.employeeID
}

func (a *AttendanceRecord) Date() time.Time {
	return a.date
}

func (a *AttendanceRecord) CheckInAt() *time.Time {
	return a.checkInAt
}

func (a *AttendanceRecord) CheckOutAt() *time.Time {
	return a.checkOutAt
}

func (a *AttendanceRecord) CheckInLocation() (*float64, *float64) {
	return a.checkInLatitude, a.checkInLongitude
}

func (a *AttendanceRecord) CheckOutLocation() (*float64, *float64) {
	return a.checkOutLatitude, a.checkOutLongitude
}

func (a *AttendanceRecord) Status() AttendanceStatus {
	return a.status
}

func (a *AttendanceRecord) WorkDurationMins() *int {
	return a.workDurationMins
}

func (a *AttendanceRecord) Notes() string {
	return a.notes
}

func (a *AttendanceRecord) CreatedBy() string {
	return a.createdBy
}

func (a *AttendanceRecord) IsOverride() bool {
	return a.isOverride
}

func (a *AttendanceRecord) CreatedAt() time.Time {
	return a.createdAt
}

func (a *AttendanceRecord) UpdatedAt() time.Time {
	return a.updatedAt
}

// HasCheckedIn returns true if the employee has checked in.
func (a *AttendanceRecord) HasCheckedIn() bool {
	return a.checkInAt != nil
}

// HasCheckedOut returns true if the employee has checked out.
func (a *AttendanceRecord) HasCheckedOut() bool {
	return a.checkOutAt != nil
}

// IsActiveCheckIn returns true if employee is currently checked in but not checked out.
func (a *AttendanceRecord) IsActiveCheckIn() bool {
	return a.checkInAt != nil && a.checkOutAt == nil && !a.isOverride
}
