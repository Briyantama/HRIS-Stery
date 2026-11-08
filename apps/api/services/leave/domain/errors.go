package domain

import "errors"

// Domain errors.
var (
	ErrInvalidLeaveRequestID        = errors.New("invalid leave request id")
	ErrInvalidLeaveTypeID           = errors.New("invalid leave type id")
	ErrInvalidLeaveBalanceID        = errors.New("invalid leave balance id")
	ErrInvalidTenantID              = errors.New("invalid tenant id")
	ErrInvalidEmployeeID            = errors.New("invalid employee id")
	ErrInvalidDate                  = errors.New("invalid date")
	ErrInvalidDaysCount             = errors.New("invalid days count")
	ErrNotFound                     = errors.New("not found")
	ErrAlreadyExists                = errors.New("already exists")
	ErrCannotApproveAlreadyRejected = errors.New("cannot approve already rejected leave request")
	ErrCannotRejectAlreadyApproved  = errors.New("cannot reject already approved leave request")
	ErrCannotCancelRejected         = errors.New("cannot cancel rejected leave request")
	ErrCannotCancelPending          = errors.New("cannot cancel pending leave request (reject first)")
	ErrInsufficientLeaveBalance     = errors.New("insufficient leave balance")
	ErrInvalidStatusTransition      = errors.New("invalid status transition")
	ErrLeaveOverlapExists           = errors.New("leave overlaps with existing approved leave")
	ErrInvalidLeaveStatus           = errors.New("invalid leave status")
	ErrMissingRequiredField         = errors.New("missing required field")
)
