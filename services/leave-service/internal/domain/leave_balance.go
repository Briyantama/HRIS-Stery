package domain

import (
	"time"
)

// LeaveBalance tracks leave entitlements per employee per leave type per year.
type LeaveBalance struct {
	id             LeaveBalanceID
	tenantID       TenantID
	employeeID     EmployeeID
	leaveTypeID    LeaveTypeID
	year           int
	entitledDays   float64
	usedDays       float64
	pendingDays    float64
	createdAt      time.Time
	updatedAt      time.Time
}

// NewLeaveBalance creates a new leave balance.
func NewLeaveBalance(
	tenantID TenantID,
	employeeID EmployeeID,
	leaveTypeID LeaveTypeID,
	year int,
	entitledDays float64,
) (*LeaveBalance, error) {
	if tenantID.IsZero() {
		return nil, ErrMissingRequiredField
	}
	if employeeID.IsZero() {
		return nil, ErrMissingRequiredField
	}
	if leaveTypeID.IsZero() {
		return nil, ErrMissingRequiredField
	}
	if year <= 0 {
		return nil, ErrInvalidDate
	}
	if entitledDays < 0 {
		return nil, ErrInvalidDaysCount
	}

	now := time.Now().UTC()
	return &LeaveBalance{
		id:           GenerateLeaveBalanceID(),
		tenantID:     tenantID,
		employeeID:   employeeID,
		leaveTypeID:  leaveTypeID,
		year:         year,
		entitledDays: entitledDays,
		usedDays:     0,
		pendingDays:  0,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// RehydrateLeaveBalance reconstructs a leave balance from persisted data.
func RehydrateLeaveBalance(
	id LeaveBalanceID,
	tenantID TenantID,
	employeeID EmployeeID,
	leaveTypeID LeaveTypeID,
	year int,
	entitledDays float64,
	usedDays float64,
	pendingDays float64,
	createdAt time.Time,
	updatedAt time.Time,
) *LeaveBalance {
	return &LeaveBalance{
		id:           id,
		tenantID:     tenantID,
		employeeID:   employeeID,
		leaveTypeID:  leaveTypeID,
		year:         year,
		entitledDays: entitledDays,
		usedDays:     usedDays,
		pendingDays:  pendingDays,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// AddPending adds days to pending balance.
// Invariant: pending + used cannot exceed entitled.
func (lb *LeaveBalance) AddPending(days float64) error {
	if days <= 0 {
		return ErrInvalidDaysCount
	}

	newPending := lb.pendingDays + days
	if newPending+lb.usedDays > lb.entitledDays {
		return ErrInsufficientLeaveBalance
	}

	lb.pendingDays = newPending
	lb.updatedAt = time.Now().UTC()
	return nil
}

// RemovePending removes days from pending balance.
func (lb *LeaveBalance) RemovePending(days float64) error {
	if days <= 0 {
		return ErrInvalidDaysCount
	}

	if days > lb.pendingDays {
		return ErrInvalidDaysCount
	}

	lb.pendingDays -= days
	lb.updatedAt = time.Now().UTC()
	return nil
}

// AddUsed adds days to used balance (when a request is approved).
// Invariant: used + pending cannot exceed entitled.
func (lb *LeaveBalance) AddUsed(days float64) error {
	if days <= 0 {
		return ErrInvalidDaysCount
	}

	newUsed := lb.usedDays + days
	if newUsed+lb.pendingDays > lb.entitledDays {
		return ErrInsufficientLeaveBalance
	}

	lb.usedDays = newUsed
	lb.updatedAt = time.Now().UTC()
	return nil
}

// SubtractUsed removes days from used balance (when an approved request is cancelled).
func (lb *LeaveBalance) SubtractUsed(days float64) error {
	if days <= 0 {
		return ErrInvalidDaysCount
	}

	if days > lb.usedDays {
		return ErrInvalidDaysCount
	}

	lb.usedDays -= days
	lb.updatedAt = time.Now().UTC()
	return nil
}

// RemainingDays calculates remaining leave balance.
func (lb *LeaveBalance) RemainingDays() float64 {
	return lb.entitledDays - lb.usedDays - lb.pendingDays
}

// Getters (read-only access)

func (lb *LeaveBalance) ID() LeaveBalanceID {
	return lb.id
}

func (lb *LeaveBalance) TenantID() TenantID {
	return lb.tenantID
}

func (lb *LeaveBalance) EmployeeID() EmployeeID {
	return lb.employeeID
}

func (lb *LeaveBalance) LeaveTypeID() LeaveTypeID {
	return lb.leaveTypeID
}

func (lb *LeaveBalance) Year() int {
	return lb.year
}

func (lb *LeaveBalance) EntitledDays() float64 {
	return lb.entitledDays
}

func (lb *LeaveBalance) UsedDays() float64 {
	return lb.usedDays
}

func (lb *LeaveBalance) PendingDays() float64 {
	return lb.pendingDays
}

func (lb *LeaveBalance) CreatedAt() time.Time {
	return lb.createdAt
}

func (lb *LeaveBalance) UpdatedAt() time.Time {
	return lb.updatedAt
}
