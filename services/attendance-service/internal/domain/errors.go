package domain

import "errors"

// Domain errors.
var (
	ErrInvalidAttendanceID  = errors.New("invalid attendance id")
	ErrInvalidTenantID      = errors.New("invalid tenant id")
	ErrInvalidEmployeeID    = errors.New("invalid employee id")
	ErrInvalidDate          = errors.New("invalid date")
	ErrNotFound             = errors.New("not found")
	ErrAlreadyExists        = errors.New("already exists")
	ErrCannotCheckOutBeforeCheckIn = errors.New("cannot check out before check in")
	ErrCannotCheckInTwiceSameDay = errors.New("cannot check in twice on same day")
	ErrAlreadyCheckedOut    = errors.New("already checked out")
	ErrInvalidTimeRange     = errors.New("check-in time must be before check-out time")
	ErrMissingRequiredField = errors.New("missing required field")
)
