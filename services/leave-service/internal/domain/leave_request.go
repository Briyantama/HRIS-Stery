package domain

import (
	"time"
)

// LeaveStatus represents the leave request status.
type LeaveStatus string

const (
	StatusPending   LeaveStatus = "PENDING"
	StatusApproved  LeaveStatus = "APPROVED"
	StatusRejected  LeaveStatus = "REJECTED"
	StatusCancelled LeaveStatus = "CANCELLED"
)

// IsValid returns true if the status is a valid LeaveStatus.
func (s LeaveStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusRejected, StatusCancelled:
		return true
	default:
		return false
	}
}

// LeaveRequest is the aggregate root for leave requests.
type LeaveRequest struct {
	id              LeaveRequestID
	tenantID        TenantID
	employeeID      EmployeeID
	leaveTypeID     LeaveTypeID
	startDate       time.Time
	endDate         time.Time
	daysCount       int
	status          LeaveStatus
	reason          string
	rejectionReason string
	approvedByID    *EmployeeID
	approvedAt      *time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

// NewLeaveRequest creates a new leave request.
func NewLeaveRequest(
	tenantID TenantID,
	employeeID EmployeeID,
	leaveTypeID LeaveTypeID,
	startDate time.Time,
	endDate time.Time,
	daysCount int,
	reason string,
	createdBy string,
) (*LeaveRequest, error) {
	// Validate inputs
	if tenantID.IsZero() {
		return nil, ErrMissingRequiredField
	}
	if employeeID.IsZero() {
		return nil, ErrMissingRequiredField
	}
	if leaveTypeID.IsZero() {
		return nil, ErrMissingRequiredField
	}
	if startDate.IsZero() || endDate.IsZero() {
		return nil, ErrInvalidDate
	}
	if endDate.Before(startDate) {
		return nil, ErrInvalidDate
	}
	if daysCount <= 0 {
		return nil, ErrInvalidDaysCount
	}

	now := time.Now().UTC()
	return &LeaveRequest{
		id:         GenerateLeaveRequestID(),
		tenantID:   tenantID,
		employeeID: employeeID,
		leaveTypeID: leaveTypeID,
		startDate:  startDate,
		endDate:    endDate,
		daysCount:  daysCount,
		status:     StatusPending,
		reason:     reason,
		createdAt:  now,
		updatedAt:  now,
	}, nil
}

// RehydrateLeaveRequest reconstructs a leave request from persisted data.
func RehydrateLeaveRequest(
	id LeaveRequestID,
	tenantID TenantID,
	employeeID EmployeeID,
	leaveTypeID LeaveTypeID,
	startDate time.Time,
	endDate time.Time,
	daysCount int,
	status LeaveStatus,
	reason string,
	rejectionReason string,
	approvedByID *EmployeeID,
	approvedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *LeaveRequest {
	return &LeaveRequest{
		id:              id,
		tenantID:        tenantID,
		employeeID:      employeeID,
		leaveTypeID:     leaveTypeID,
		startDate:       startDate,
		endDate:         endDate,
		daysCount:       daysCount,
		status:          status,
		reason:          reason,
		rejectionReason: rejectionReason,
		approvedByID:    approvedByID,
		approvedAt:      approvedAt,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

// Approve approves the leave request.
// Invariant: Can only approve if status is PENDING.
func (lr *LeaveRequest) Approve(approverId EmployeeID) error {
	if lr.status != StatusPending {
		return ErrCannotApproveAlreadyRejected
	}

	lr.status = StatusApproved
	lr.approvedByID = &approverId
	now := time.Now().UTC()
	lr.approvedAt = &now
	lr.updatedAt = now

	return nil
}

// Reject rejects the leave request.
// Invariant: Can only reject if status is PENDING.
func (lr *LeaveRequest) Reject(reason string) error {
	if lr.status != StatusPending {
		return ErrCannotRejectAlreadyApproved
	}

	lr.status = StatusRejected
	lr.rejectionReason = reason
	lr.updatedAt = time.Now().UTC()

	return nil
}

// Cancel cancels the leave request.
// Invariant: Can only cancel if status is APPROVED.
func (lr *LeaveRequest) Cancel(reason string) error {
	if lr.status == StatusRejected {
		return ErrCannotCancelRejected
	}
	if lr.status == StatusCancelled {
		return ErrInvalidStatusTransition
	}

	lr.status = StatusCancelled
	lr.rejectionReason = reason
	lr.updatedAt = time.Now().UTC()

	return nil
}

// Getters (read-only access)

func (lr *LeaveRequest) ID() LeaveRequestID {
	return lr.id
}

func (lr *LeaveRequest) TenantID() TenantID {
	return lr.tenantID
}

func (lr *LeaveRequest) EmployeeID() EmployeeID {
	return lr.employeeID
}

func (lr *LeaveRequest) LeaveTypeID() LeaveTypeID {
	return lr.leaveTypeID
}

func (lr *LeaveRequest) StartDate() time.Time {
	return lr.startDate
}

func (lr *LeaveRequest) EndDate() time.Time {
	return lr.endDate
}

func (lr *LeaveRequest) DaysCount() int {
	return lr.daysCount
}

func (lr *LeaveRequest) Status() LeaveStatus {
	return lr.status
}

func (lr *LeaveRequest) Reason() string {
	return lr.reason
}

func (lr *LeaveRequest) RejectionReason() string {
	return lr.rejectionReason
}

func (lr *LeaveRequest) ApprovedByID() *EmployeeID {
	return lr.approvedByID
}

func (lr *LeaveRequest) ApprovedAt() *time.Time {
	return lr.approvedAt
}

func (lr *LeaveRequest) CreatedAt() time.Time {
	return lr.createdAt
}

func (lr *LeaveRequest) UpdatedAt() time.Time {
	return lr.updatedAt
}

// Helper methods

// IsApproved returns true if the leave request is approved.
func (lr *LeaveRequest) IsApproved() bool {
	return lr.status == StatusApproved
}

// IsRejected returns true if the leave request is rejected.
func (lr *LeaveRequest) IsRejected() bool {
	return lr.status == StatusRejected
}

// IsCancelled returns true if the leave request is cancelled.
func (lr *LeaveRequest) IsCancelled() bool {
	return lr.status == StatusCancelled
}

// IsPending returns true if the leave request is pending.
func (lr *LeaveRequest) IsPending() bool {
	return lr.status == StatusPending
}
