package domain

import "time"

// DomainEvent is the interface that all domain events must implement.
type DomainEvent interface {
	EventID() string
	EventType() string
	OccurredAt() time.Time
}

// AttendanceCheckedInEvent is published when an employee checks in.
type AttendanceCheckedInEvent struct {
	eventID      string
	eventType    string
	tenantID     TenantID
	employeeID   EmployeeID
	attendanceID AttendanceID
	checkInAt    time.Time
	latitude     *float64
	longitude    *float64
	occurredAt   time.Time
}

// NewAttendanceCheckedInEvent creates a new check-in event.
func NewAttendanceCheckedInEvent(
	eventID string,
	tenantID TenantID,
	employeeID EmployeeID,
	attendanceID AttendanceID,
	checkInAt time.Time,
	latitude *float64,
	longitude *float64,
) *AttendanceCheckedInEvent {
	return &AttendanceCheckedInEvent{
		eventID:      eventID,
		eventType:    "hris.attendance.record.checked_in",
		tenantID:     tenantID,
		employeeID:   employeeID,
		attendanceID: attendanceID,
		checkInAt:    checkInAt,
		latitude:     latitude,
		longitude:    longitude,
		occurredAt:   time.Now().UTC(),
	}
}

func (e *AttendanceCheckedInEvent) EventID() string {
	return e.eventID
}

func (e *AttendanceCheckedInEvent) EventType() string {
	return e.eventType
}

func (e *AttendanceCheckedInEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *AttendanceCheckedInEvent) TenantID() TenantID {
	return e.tenantID
}

func (e *AttendanceCheckedInEvent) EmployeeID() EmployeeID {
	return e.employeeID
}

func (e *AttendanceCheckedInEvent) AttendanceID() AttendanceID {
	return e.attendanceID
}

func (e *AttendanceCheckedInEvent) CheckInAt() time.Time {
	return e.checkInAt
}

func (e *AttendanceCheckedInEvent) Location() (*float64, *float64) {
	return e.latitude, e.longitude
}

// AttendanceCheckedOutEvent is published when an employee checks out.
type AttendanceCheckedOutEvent struct {
	eventID          string
	eventType        string
	tenantID         TenantID
	employeeID       EmployeeID
	attendanceID     AttendanceID
	checkOutAt       time.Time
	latitude         *float64
	longitude        *float64
	workDurationMins int
	occurredAt       time.Time
}

// NewAttendanceCheckedOutEvent creates a new check-out event.
func NewAttendanceCheckedOutEvent(
	eventID string,
	tenantID TenantID,
	employeeID EmployeeID,
	attendanceID AttendanceID,
	checkOutAt time.Time,
	latitude *float64,
	longitude *float64,
	workDurationMins int,
) *AttendanceCheckedOutEvent {
	return &AttendanceCheckedOutEvent{
		eventID:          eventID,
		eventType:        "hris.attendance.record.checked_out",
		tenantID:         tenantID,
		employeeID:       employeeID,
		attendanceID:     attendanceID,
		checkOutAt:       checkOutAt,
		latitude:         latitude,
		longitude:        longitude,
		workDurationMins: workDurationMins,
		occurredAt:       time.Now().UTC(),
	}
}

func (e *AttendanceCheckedOutEvent) EventID() string {
	return e.eventID
}

func (e *AttendanceCheckedOutEvent) EventType() string {
	return e.eventType
}

func (e *AttendanceCheckedOutEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *AttendanceCheckedOutEvent) TenantID() TenantID {
	return e.tenantID
}

func (e *AttendanceCheckedOutEvent) EmployeeID() EmployeeID {
	return e.employeeID
}

func (e *AttendanceCheckedOutEvent) AttendanceID() AttendanceID {
	return e.attendanceID
}

func (e *AttendanceCheckedOutEvent) CheckOutAt() time.Time {
	return e.checkOutAt
}

func (e *AttendanceCheckedOutEvent) Location() (*float64, *float64) {
	return e.latitude, e.longitude
}

func (e *AttendanceCheckedOutEvent) WorkDurationMins() int {
	return e.workDurationMins
}

// AttendanceOverriddenEvent is published when an admin overrides attendance.
type AttendanceOverriddenEvent struct {
	eventID      string
	eventType    string
	tenantID     TenantID
	employeeID   EmployeeID
	attendanceID AttendanceID
	status       AttendanceStatus
	checkInAt    *time.Time
	checkOutAt   *time.Time
	reason       string
	occurredAt   time.Time
}

// NewAttendanceOverriddenEvent creates a new override event.
func NewAttendanceOverriddenEvent(
	eventID string,
	tenantID TenantID,
	employeeID EmployeeID,
	attendanceID AttendanceID,
	status AttendanceStatus,
	checkInAt *time.Time,
	checkOutAt *time.Time,
	reason string,
) *AttendanceOverriddenEvent {
	return &AttendanceOverriddenEvent{
		eventID:      eventID,
		eventType:    "hris.attendance.record.overridden",
		tenantID:     tenantID,
		employeeID:   employeeID,
		attendanceID: attendanceID,
		status:       status,
		checkInAt:    checkInAt,
		checkOutAt:   checkOutAt,
		reason:       reason,
		occurredAt:   time.Now().UTC(),
	}
}

func (e *AttendanceOverriddenEvent) EventID() string {
	return e.eventID
}

func (e *AttendanceOverriddenEvent) EventType() string {
	return e.eventType
}

func (e *AttendanceOverriddenEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *AttendanceOverriddenEvent) TenantID() TenantID {
	return e.tenantID
}

func (e *AttendanceOverriddenEvent) EmployeeID() EmployeeID {
	return e.employeeID
}

func (e *AttendanceOverriddenEvent) AttendanceID() AttendanceID {
	return e.attendanceID
}

func (e *AttendanceOverriddenEvent) Status() AttendanceStatus {
	return e.status
}

func (e *AttendanceOverriddenEvent) CheckInAt() *time.Time {
	return e.checkInAt
}

func (e *AttendanceOverriddenEvent) CheckOutAt() *time.Time {
	return e.checkOutAt
}

func (e *AttendanceOverriddenEvent) Reason() string {
	return e.reason
}
