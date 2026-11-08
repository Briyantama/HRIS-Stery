package domain

import "fmt"

var (
	ErrInvalidTenantID         = fmt.Errorf("invalid tenant id")
	ErrInvalidNotificationID   = fmt.Errorf("invalid notification id")
	ErrInvalidRecipientID      = fmt.Errorf("invalid recipient id")
	ErrInvalidChannel          = fmt.Errorf("invalid channel")
	ErrInvalidTemplateKey      = fmt.Errorf("invalid template key")
	ErrMissingTemplateVariable = fmt.Errorf("missing template variable")
	ErrUnsupportedChannel      = fmt.Errorf("unsupported channel type")
	ErrInvalidStatusTransition = fmt.Errorf("invalid status transition")
	ErrNoChannelsEnabled       = fmt.Errorf("no notification channels enabled")
	ErrInvalidEmailConfig      = fmt.Errorf("invalid email configuration")
	ErrNotificationNotFound    = fmt.Errorf("notification not found")
	ErrChannelConfigNotFound   = fmt.Errorf("channel configuration not found")
)
