package domain

import (
	"context"
)

type ListFilters struct {
	Status     NotificationStatus
	UnreadOnly bool
	Limit      int
	Offset     int
}

type NotificationRepository interface {
	Create(ctx context.Context, notif *Notification) error
	GetByID(ctx context.Context, tenantID TenantID, id NotificationID) (*Notification, error)
	ListByRecipient(ctx context.Context, tenantID TenantID, recipientID RecipientID, filters ListFilters) ([]*Notification, error)
	Update(ctx context.Context, notif *Notification) error
}

type ChannelConfigRepository interface {
	Create(ctx context.Context, config *NotificationChannelConfig) error
	GetByTenant(ctx context.Context, tenantID TenantID) (*NotificationChannelConfig, error)
	Update(ctx context.Context, config *NotificationChannelConfig) error
}

type RenderResult struct {
	Subject string
	Body    string
}

type NotificationChannelAdapter interface {
	Send(ctx context.Context, notif *Notification, rendered RenderResult) (string, error)
}
