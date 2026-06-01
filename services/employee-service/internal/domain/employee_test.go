package domain

import (
	"testing"
	"time"
)

func TestNewEmployee(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	deptID := GenerateDepartmentID()
	posID := GeneratePositionID()

	emp, err := NewEmployee(tenantID, "john@example.com", "John Doe", deptID, posID)
	if err != nil {
		t.Fatalf("NewEmployee failed: %v", err)
	}

	if emp.Email() != "john@example.com" {
		t.Errorf("expected email john@example.com, got %s", emp.Email())
	}
	if emp.FullName() != "John Doe" {
		t.Errorf("expected full name John Doe, got %s", emp.FullName())
	}
	if !emp.IsActive() {
		t.Errorf("expected employee to be active")
	}
	if emp.IsTerminated() {
		t.Errorf("expected employee to not be terminated")
	}
	if emp.Status() != StatusActive {
		t.Errorf("expected status ACTIVE, got %s", emp.Status())
	}
}

func TestEmployeeValidation(t *testing.T) {
	tests := []struct {
		name        string
		tenantID    TenantID
		email       string
		fullName    string
		deptID      DepartmentID
		posID       PositionID
		expectError bool
	}{
		{
			name:        "valid employee",
			tenantID:    MustNewTenantID("550e8400-e29b-41d4-a716-446655440000"),
			email:       "john@example.com",
			fullName:    "John Doe",
			deptID:      GenerateDepartmentID(),
			posID:       GeneratePositionID(),
			expectError: false,
		},
		{
			name:        "missing email",
			tenantID:    MustNewTenantID("550e8400-e29b-41d4-a716-446655440000"),
			email:       "",
			fullName:    "John Doe",
			deptID:      GenerateDepartmentID(),
			posID:       GeneratePositionID(),
			expectError: true,
		},
		{
			name:        "invalid email",
			tenantID:    MustNewTenantID("550e8400-e29b-41d4-a716-446655440000"),
			email:       "not-an-email",
			fullName:    "John Doe",
			deptID:      GenerateDepartmentID(),
			posID:       GeneratePositionID(),
			expectError: true,
		},
		{
			name:        "missing full name",
			tenantID:    MustNewTenantID("550e8400-e29b-41d4-a716-446655440000"),
			email:       "john@example.com",
			fullName:    "",
			deptID:      GenerateDepartmentID(),
			posID:       GeneratePositionID(),
			expectError: true,
		},
		{
			name:        "missing department",
			tenantID:    MustNewTenantID("550e8400-e29b-41d4-a716-446655440000"),
			email:       "john@example.com",
			fullName:    "John Doe",
			deptID:      DepartmentID(""),
			posID:       GeneratePositionID(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEmployee(tt.tenantID, tt.email, tt.fullName, tt.deptID, tt.posID)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error=%v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestEmployeeTerminate(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	deptID := GenerateDepartmentID()
	posID := GeneratePositionID()

	emp, _ := NewEmployee(tenantID, "john@example.com", "John Doe", deptID, posID)

	if emp.IsTerminated() {
		t.Errorf("expected employee to not be terminated initially")
	}

	now := time.Now().UTC()
	if err := emp.Terminate(now); err != nil {
		t.Fatalf("Terminate failed: %v", err)
	}

	if !emp.IsTerminated() {
		t.Errorf("expected employee to be terminated")
	}
	if emp.Status() != StatusTerminated {
		t.Errorf("expected status TERMINATED, got %s", emp.Status())
	}
	if emp.TerminationDate() == nil {
		t.Errorf("expected termination date to be set")
	}

	// Cannot terminate twice
	if err := emp.Terminate(now); err == nil {
		t.Errorf("expected error when terminating already terminated employee")
	}
}

func TestEmployeeStatusTransition(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	deptID := GenerateDepartmentID()
	posID := GeneratePositionID()

	emp, _ := NewEmployee(tenantID, "john@example.com", "John Doe", deptID, posID)

	// Can go to ON_LEAVE from ACTIVE
	if err := emp.UpdateStatus(StatusOnLeave); err != nil {
		t.Fatalf("UpdateStatus to ON_LEAVE failed: %v", err)
	}
	if emp.Status() != StatusOnLeave {
		t.Errorf("expected status ON_LEAVE, got %s", emp.Status())
	}

	// Can go back to ACTIVE from ON_LEAVE
	if err := emp.UpdateStatus(StatusActive); err != nil {
		t.Fatalf("UpdateStatus to ACTIVE failed: %v", err)
	}
	if emp.Status() != StatusActive {
		t.Errorf("expected status ACTIVE, got %s", emp.Status())
	}

	// Can go to TERMINATED from ACTIVE
	if err := emp.UpdateStatus(StatusTerminated); err != nil {
		t.Fatalf("UpdateStatus to TERMINATED failed: %v", err)
	}
	if emp.Status() != StatusTerminated {
		t.Errorf("expected status TERMINATED, got %s", emp.Status())
	}

	// Cannot go back to ACTIVE from TERMINATED
	if err := emp.UpdateStatus(StatusActive); err == nil {
		t.Errorf("expected error when transitioning from TERMINATED to ACTIVE")
	}
}

func TestEmployeeUpdateDepartment(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	deptID1 := GenerateDepartmentID()
	deptID2 := GenerateDepartmentID()
	posID := GeneratePositionID()

	emp, _ := NewEmployee(tenantID, "john@example.com", "John Doe", deptID1, posID)

	if err := emp.UpdateDepartment(deptID2); err != nil {
		t.Fatalf("UpdateDepartment failed: %v", err)
	}
	if !emp.DepartmentID().Equals(deptID2) {
		t.Errorf("expected department to be updated")
	}

	// Terminate and try to update
	_ = emp.Terminate(time.Now().UTC())
	if err := emp.UpdateDepartment(deptID1); err == nil {
		t.Errorf("expected error when updating department of terminated employee")
	}
}
