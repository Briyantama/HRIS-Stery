package domain

import (
	"testing"
	"time"
)

func TestNewAttendanceRecord(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	date := time.Now()

	record, err := NewAttendanceRecord(tenantID, employeeID, date, "user-123")
	if err != nil {
		t.Fatalf("NewAttendanceRecord failed: %v", err)
	}

	if record.TenantID() != tenantID {
		t.Errorf("expected tenant ID %s, got %s", tenantID, record.TenantID())
	}
	if record.EmployeeID() != employeeID {
		t.Errorf("expected employee ID %s, got %s", employeeID, record.EmployeeID())
	}
	if record.Status() != StatusAbsent {
		t.Errorf("expected status ABSENT, got %s", record.Status())
	}
	if record.HasCheckedIn() {
		t.Errorf("expected no check-in, but got one")
	}
}

func TestNewAttendanceRecordMissingTenant(t *testing.T) {
	employeeID := GenerateEmployeeID()
	date := time.Now()
	emptyTenantID := TenantID("")

	_, err := NewAttendanceRecord(emptyTenantID, employeeID, date, "user-123")
	if err != ErrMissingRequiredField {
		t.Errorf("expected ErrMissingRequiredField, got %v", err)
	}
}

func TestNewAttendanceRecordMissingEmployee(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	date := time.Now()
	emptyEmployeeID := EmployeeID("")

	_, err := NewAttendanceRecord(tenantID, emptyEmployeeID, date, "user-123")
	if err != ErrMissingRequiredField {
		t.Errorf("expected ErrMissingRequiredField, got %v", err)
	}
}

func TestNewAttendanceRecordInvalidDate(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	zeroDate := time.Time{}

	_, err := NewAttendanceRecord(tenantID, employeeID, zeroDate, "user-123")
	if err != ErrInvalidDate {
		t.Errorf("expected ErrInvalidDate, got %v", err)
	}
}

func TestCheckInSuccess(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	date := time.Now()
	checkInTime := time.Now()

	record, _ := NewAttendanceRecord(tenantID, employeeID, date, "user-123")
	latitude := 6.2088
	longitude := 106.8456

	err := record.CheckIn(checkInTime, &latitude, &longitude, "Office")
	if err != nil {
		t.Fatalf("CheckIn failed: %v", err)
	}

	if !record.HasCheckedIn() {
		t.Errorf("expected check-in recorded")
	}
	if record.Status() != StatusPresent {
		t.Errorf("expected status PRESENT, got %s", record.Status())
	}
	if record.CheckInAt() == nil || !record.CheckInAt().Equal(checkInTime) {
		t.Errorf("check-in time not recorded correctly")
	}
	if record.IsActiveCheckIn() != true {
		t.Errorf("expected active check-in")
	}
}

func TestCheckInTwiceSameDay(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	date := time.Now()
	checkInTime := time.Now()

	record, _ := NewAttendanceRecord(tenantID, employeeID, date, "user-123")
	_ = record.CheckIn(checkInTime, nil, nil, "")

	// Try to check in again
	err := record.CheckIn(checkInTime.Add(1*time.Hour), nil, nil, "")
	if err != ErrCannotCheckInTwiceSameDay {
		t.Errorf("expected ErrCannotCheckInTwiceSameDay, got %v", err)
	}
}

func TestCheckOutBeforeCheckIn(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	date := time.Now()
	checkOutTime := time.Now()

	record, _ := NewAttendanceRecord(tenantID, employeeID, date, "user-123")

	err := record.CheckOut(checkOutTime, nil, nil, "")
	if err != ErrCannotCheckOutBeforeCheckIn {
		t.Errorf("expected ErrCannotCheckOutBeforeCheckIn, got %v", err)
	}
}

func TestCheckOutSuccess(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	date := time.Now()
	checkInTime := time.Now()
	checkOutTime := checkInTime.Add(8 * time.Hour)

	record, _ := NewAttendanceRecord(tenantID, employeeID, date, "user-123")
	_ = record.CheckIn(checkInTime, nil, nil, "")

	latitude := 6.2088
	longitude := 106.8456
	err := record.CheckOut(checkOutTime, &latitude, &longitude, "End of shift")
	if err != nil {
		t.Fatalf("CheckOut failed: %v", err)
	}

	if !record.HasCheckedOut() {
		t.Errorf("expected check-out recorded")
	}
	if record.IsActiveCheckIn() {
		t.Errorf("expected no active check-in after checkout")
	}
	if record.WorkDurationMins() == nil || *record.WorkDurationMins() != 480 {
		t.Errorf("expected work duration 480 mins, got %v", record.WorkDurationMins())
	}
}

func TestCheckOutBeforeCheckInTime(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	date := time.Now()
	checkInTime := time.Now()
	checkOutTime := checkInTime.Add(-1 * time.Hour) // Before check-in

	record, _ := NewAttendanceRecord(tenantID, employeeID, date, "user-123")
	_ = record.CheckIn(checkInTime, nil, nil, "")

	err := record.CheckOut(checkOutTime, nil, nil, "")
	if err != ErrInvalidTimeRange {
		t.Errorf("expected ErrInvalidTimeRange, got %v", err)
	}
}

func TestCheckOutTwice(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	date := time.Now()
	checkInTime := time.Now()
	checkOutTime := checkInTime.Add(8 * time.Hour)

	record, _ := NewAttendanceRecord(tenantID, employeeID, date, "user-123")
	_ = record.CheckIn(checkInTime, nil, nil, "")
	_ = record.CheckOut(checkOutTime, nil, nil, "")

	// Try to check out again
	err := record.CheckOut(checkOutTime.Add(1*time.Hour), nil, nil, "")
	if err != ErrAlreadyCheckedOut {
		t.Errorf("expected ErrAlreadyCheckedOut, got %v", err)
	}
}

func TestOverrideToPresent(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	date := time.Now()

	record, _ := NewAttendanceRecord(tenantID, employeeID, date, "user-123")

	checkInTime := time.Now()
	checkOutTime := checkInTime.Add(8 * time.Hour)
	err := record.Override(&checkInTime, &checkOutTime, StatusPresent, "Manual override")
	if err != nil {
		t.Fatalf("Override failed: %v", err)
	}

	if record.Status() != StatusPresent {
		t.Errorf("expected status PRESENT, got %s", record.Status())
	}
	if !record.IsOverride() {
		t.Errorf("expected override flag to be set")
	}
	if record.WorkDurationMins() == nil || *record.WorkDurationMins() != 480 {
		t.Errorf("expected work duration 480 mins, got %v", record.WorkDurationMins())
	}
}

func TestOverrideToAbsent(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	date := time.Now()

	record, _ := NewAttendanceRecord(tenantID, employeeID, date, "user-123")

	err := record.Override(nil, nil, StatusAbsent, "Employee did not show up")
	if err != nil {
		t.Fatalf("Override failed: %v", err)
	}

	if record.Status() != StatusAbsent {
		t.Errorf("expected status ABSENT, got %s", record.Status())
	}
	if record.CheckInAt() != nil {
		t.Errorf("expected no check-in after override to absent")
	}
}

func TestOverrideWithInvalidTimeRange(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	date := time.Now()

	record, _ := NewAttendanceRecord(tenantID, employeeID, date, "user-123")

	checkInTime := time.Now()
	checkOutTime := checkInTime.Add(-1 * time.Hour) // Before check-in

	err := record.Override(&checkInTime, &checkOutTime, StatusPresent, "Invalid times")
	if err != ErrInvalidTimeRange {
		t.Errorf("expected ErrInvalidTimeRange, got %v", err)
	}
}

func TestOverrideWithInvalidStatus(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	date := time.Now()

	record, _ := NewAttendanceRecord(tenantID, employeeID, date, "user-123")

	err := record.Override(nil, nil, AttendanceStatus("INVALID"), "Invalid status")
	if err != ErrInvalidDate {
		t.Errorf("expected ErrInvalidDate for invalid status, got %v", err)
	}
}

func TestRehydrateAttendanceRecord(t *testing.T) {
	attendanceID := GenerateAttendanceID()
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	date := time.Now()
	checkInTime := time.Now()
	checkOutTime := checkInTime.Add(8 * time.Hour)
	durationMins := 480

	record := RehydrateAttendanceRecord(
		attendanceID,
		tenantID,
		employeeID,
		date,
		&checkInTime,
		&checkOutTime,
		nil, nil, nil, nil,
		StatusPresent,
		&durationMins,
		"Test note",
		"user-123",
		false,
		time.Now(),
		time.Now(),
	)

	if record.ID() != attendanceID {
		t.Errorf("expected ID %s, got %s", attendanceID, record.ID())
	}
	if record.Status() != StatusPresent {
		t.Errorf("expected status PRESENT, got %s", record.Status())
	}
	if record.WorkDurationMins() == nil || *record.WorkDurationMins() != 480 {
		t.Errorf("expected duration 480, got %v", record.WorkDurationMins())
	}
}

func TestAttendanceStatusIsValid(t *testing.T) {
	tests := []struct {
		status AttendanceStatus
		valid  bool
	}{
		{StatusPresent, true},
		{StatusLate, true},
		{StatusAbsent, true},
		{StatusHalfDay, true},
		{StatusOnLeave, true},
		{AttendanceStatus("INVALID"), false},
		{AttendanceStatus(""), false},
	}

	for _, tt := range tests {
		if tt.status.IsValid() != tt.valid {
			t.Errorf("status %s: expected valid=%v", tt.status, tt.valid)
		}
	}
}

func TestNotesAppendOnCheckOut(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	date := time.Now()
	checkInTime := time.Now()
	checkOutTime := checkInTime.Add(8 * time.Hour)

	record, _ := NewAttendanceRecord(tenantID, employeeID, date, "user-123")
	_ = record.CheckIn(checkInTime, nil, nil, "Morning note")
	_ = record.CheckOut(checkOutTime, nil, nil, "Evening note")

	expectedNotes := "Morning note; Evening note"
	if record.Notes() != expectedNotes {
		t.Errorf("expected notes %q, got %q", expectedNotes, record.Notes())
	}
}
