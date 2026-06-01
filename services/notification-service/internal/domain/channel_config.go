package domain

import (
	"time"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromAddress  string
	FromName     string
}

type NotificationChannelConfig struct {
	tenantID        TenantID
	enabledChannels []ChannelType
	emailConfig     *EmailConfig
	createdAt       time.Time
	updatedAt       time.Time
}

func NewNotificationChannelConfig(
	tenantID TenantID,
	enabledChannels []ChannelType,
	emailConfig *EmailConfig,
) (*NotificationChannelConfig, error) {
	if tenantID.IsZero() {
		return nil, ErrInvalidTenantID
	}
	if len(enabledChannels) == 0 {
		return nil, ErrNoChannelsEnabled
	}

	for _, ch := range enabledChannels {
		if !ch.IsValid() {
			return nil, ErrInvalidChannel
		}
	}

	// If email is enabled, validate email config
	for _, ch := range enabledChannels {
		if ch == ChannelTypeEmail {
			if emailConfig == nil || emailConfig.SMTPHost == "" || emailConfig.FromAddress == "" {
				return nil, ErrInvalidEmailConfig
			}
		}
	}

	now := time.Now().UTC()
	return &NotificationChannelConfig{
		tenantID:        tenantID,
		enabledChannels: enabledChannels,
		emailConfig:     emailConfig,
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

func RehydrateNotificationChannelConfig(
	tenantID TenantID,
	enabledChannels []ChannelType,
	emailConfig *EmailConfig,
	createdAt time.Time,
	updatedAt time.Time,
) *NotificationChannelConfig {
	return &NotificationChannelConfig{
		tenantID:        tenantID,
		enabledChannels: enabledChannels,
		emailConfig:     emailConfig,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

func (ncc *NotificationChannelConfig) IsChannelEnabled(channel ChannelType) bool {
	for _, ch := range ncc.enabledChannels {
		if ch == channel {
			return true
		}
	}
	return false
}

func (ncc *NotificationChannelConfig) GetEmailConfig() *EmailConfig {
	return ncc.emailConfig
}

func (ncc *NotificationChannelConfig) TenantID() TenantID {
	return ncc.tenantID
}

func (ncc *NotificationChannelConfig) EnabledChannels() []ChannelType {
	return ncc.enabledChannels
}

func (ncc *NotificationChannelConfig) CreatedAt() time.Time {
	return ncc.createdAt
}

func (ncc *NotificationChannelConfig) UpdatedAt() time.Time {
	return ncc.updatedAt
}
