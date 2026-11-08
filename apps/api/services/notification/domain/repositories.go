package domain

import (
	"context"
)

type ListFilters struct {
	UnreadOnly bool
	Limit      int
	Offset     int
}

type NotificationRepository interface {
	Create(ctx context.Context, notif *Notification) error
	GetByID(ctx context.Context, tenantID TenantID, id NotificationID) (*Notification, error)
	ListByRecipient(ctx context.Context, tenantID TenantID, recipientID RecipientID, filters ListFilters) ([]*Notification, error)
	CountByRecipient(ctx context.Context, tenantID TenantID, recipientID RecipientID, filters ListFilters) (int, error)
	Update(ctx context.Context, notif *Notification) error
}

type ChannelConfigRepository interface {
	Create(ctx context.Context, config *NotificationChannelConfig) error
	GetByTenant(ctx context.Context, tenantID TenantID) (*NotificationChannelConfig, error)
	Update(ctx context.Context, config *NotificationChannelConfig) error
}
