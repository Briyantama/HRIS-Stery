package domain

import (
	"testing"
	"time"
)

func TestNewLeaveRequest(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	leaveTypeID := GenerateLeaveTypeID()
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, 5)

	req, err := NewLeaveRequest(tenantID, employeeID, leaveTypeID, startDate, endDate, 5, "Vacation", "user-123")
	if err != nil {
		t.Fatalf("NewLeaveRequest failed: %v", err)
	}

	if req.Status() != StatusPending {
		t.Errorf("expected PENDING, got %s", req.Status())
	}
	if req.DaysCount() != 5 {
		t.Errorf("expected 5 days, got %d", req.DaysCount())
	}
	if req.IsPending() == false {
		t.Errorf("expected IsPending to be true")
	}
}

func TestApproveLeaveRequest(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	leaveTypeID := GenerateLeaveTypeID()
	approverId := GenerateEmployeeID()
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, 5)

	req, _ := NewLeaveRequest(tenantID, employeeID, leaveTypeID, startDate, endDate, 5, "Vacation", "user-123")
	err := req.Approve(approverId)
	if err != nil {
		t.Fatalf("Approve failed: %v", err)
	}

	if req.Status() != StatusApproved {
		t.Errorf("expected APPROVED, got %s", req.Status())
	}
	if req.IsApproved() == false {
		t.Errorf("expected IsApproved to be true")
	}
	if req.ApprovedByID() == nil {
		t.Errorf("expected ApprovedByID to be set")
	}
}

func TestCannotApproveTwice(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	leaveTypeID := GenerateLeaveTypeID()
	approverId := GenerateEmployeeID()
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, 5)

	req, _ := NewLeaveRequest(tenantID, employeeID, leaveTypeID, startDate, endDate, 5, "Vacation", "user-123")
	_ = req.Approve(approverId)

	err := req.Approve(approverId)
	if err != ErrCannotApproveAlreadyRejected {
		t.Errorf("expected ErrCannotApproveAlreadyRejected, got %v", err)
	}
}

func TestRejectLeaveRequest(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	leaveTypeID := GenerateLeaveTypeID()
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, 5)

	req, _ := NewLeaveRequest(tenantID, employeeID, leaveTypeID, startDate, endDate, 5, "Vacation", "user-123")
	err := req.Reject("Business constraints")
	if err != nil {
		t.Fatalf("Reject failed: %v", err)
	}

	if req.Status() != StatusRejected {
		t.Errorf("expected REJECTED, got %s", req.Status())
	}
	if req.IsRejected() == false {
		t.Errorf("expected IsRejected to be true")
	}
}

func TestCancelApprovedLeaveRequest(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	leaveTypeID := GenerateLeaveTypeID()
	approverId := GenerateEmployeeID()
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, 5)

	req, _ := NewLeaveRequest(tenantID, employeeID, leaveTypeID, startDate, endDate, 5, "Vacation", "user-123")
	_ = req.Approve(approverId)

	err := req.Cancel("Changed plans")
	if err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}

	if req.Status() != StatusCancelled {
		t.Errorf("expected CANCELLED, got %s", req.Status())
	}
	if req.IsCancelled() == false {
		t.Errorf("expected IsCancelled to be true")
	}
}

func TestCannotCancelRejected(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	leaveTypeID := GenerateLeaveTypeID()
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, 5)

	req, _ := NewLeaveRequest(tenantID, employeeID, leaveTypeID, startDate, endDate, 5, "Vacation", "user-123")
	_ = req.Reject("Business constraints")

	err := req.Cancel("Changed plans")
	if err != ErrCannotCancelRejected {
		t.Errorf("expected ErrCannotCancelRejected, got %v", err)
	}
}

func TestInvalidDateRange(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	leaveTypeID := GenerateLeaveTypeID()
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, -5) // End before start

	_, err := NewLeaveRequest(tenantID, employeeID, leaveTypeID, startDate, endDate, 5, "Vacation", "user-123")
	if err != ErrInvalidDate {
		t.Errorf("expected ErrInvalidDate, got %v", err)
	}
}

func TestLeaveBalanceAddPending(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	leaveTypeID := GenerateLeaveTypeID()

	balance, _ := NewLeaveBalance(tenantID, employeeID, leaveTypeID, 2026, 20.0)
	err := balance.AddPending(5.0)
	if err != nil {
		t.Fatalf("AddPending failed: %v", err)
	}

	if balance.PendingDays() != 5.0 {
		t.Errorf("expected 5 pending days, got %f", balance.PendingDays())
	}
	if balance.RemainingDays() != 15.0 {
		t.Errorf("expected 15 remaining days, got %f", balance.RemainingDays())
	}
}

func TestLeaveBalanceInsufficientBalance(t *testing.T) {
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	leaveTypeID := GenerateLeaveTypeID()

	balance, _ := NewLeaveBalance(tenantID, employeeID, leaveTypeID, 2026, 10.0)
	_ = balance.AddPending(8.0)

	err := balance.AddPending(5.0) // Would exceed entitled days
	if err != ErrInsufficientLeaveBalance {
		t.Errorf("expected ErrInsufficientLeaveBalance, got %v", err)
	}
}

func TestRehydrateLeaveRequest(t *testing.T) {
	requestID := GenerateLeaveRequestID()
	tenantID := MustNewTenantID("550e8400-e29b-41d4-a716-446655440000")
	employeeID := GenerateEmployeeID()
	leaveTypeID := GenerateLeaveTypeID()
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, 5)

	req := RehydrateLeaveRequest(
		requestID,
		tenantID,
		employeeID,
		leaveTypeID,
		startDate,
		endDate,
		5,
		StatusApproved,
		"Vacation",
		"",
		nil,
		nil,
		time.Now(),
		time.Now(),
	)

	if req.ID() != requestID {
		t.Errorf("expected ID %s, got %s", requestID, req.ID())
	}
	if req.Status() != StatusApproved {
		t.Errorf("expected APPROVED, got %s", req.Status())
	}
}
