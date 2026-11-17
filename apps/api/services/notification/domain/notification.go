package domain

import (
	"fmt"
	"time"
)

type NotificationStatus int

const (
	NotificationStatusUnknown NotificationStatus = iota
	NotificationStatusPending
	NotificationStatusSent
	NotificationStatusFailed
	NotificationStatusRead
)

func (ns NotificationStatus) String() string {
	switch ns {
	case NotificationStatusPending:
		return "PENDING"
	case NotificationStatusSent:
		return "SENT"
	case NotificationStatusFailed:
		return "FAILED"
	case NotificationStatusRead:
		return "READ"
	default:
		return "UNKNOWN"
	}
}

func NotificationStatusFromString(s string) (NotificationStatus, error) {
	switch s {
	case "PENDING":
		return NotificationStatusPending, nil
	case "SENT":
		return NotificationStatusSent, nil
	case "FAILED":
		return NotificationStatusFailed, nil
	case "READ":
		return NotificationStatusRead, nil
	default:
		return NotificationStatusUnknown, fmt.Errorf("invalid status: %s", s)
	}
}

type Notification struct {
	id            NotificationID
	tenantID      TenantID
	recipientID   RecipientID
	channel       ChannelType
	templateKey   string
	variables     map[string]string
	subject       string
	body          string
	status        NotificationStatus
	deliveryID    string
	deliveryError string
	scheduledAt   *time.Time
	sentAt        *time.Time
	readAt        *time.Time
	createdAt     time.Time
	updatedAt     time.Time
}

func NewNotification(
	tenantID TenantID,
	recipientID RecipientID,
	channel ChannelType,
	templateKey string,
	variables map[string]string,
	subject string,
	body string,
) (*Notification, error) {
	if tenantID.IsZero() {
		return nil, ErrInvalidTenantID
	}
	if recipientID.IsZero() {
		return nil, ErrInvalidRecipientID
	}
	if !channel.IsValid() {
		return nil, ErrInvalidChannel
	}
	if templateKey == "" {
		return nil, ErrInvalidTemplateKey
	}

	now := time.Now().UTC()
	return &Notification{
		id:          GenerateNotificationID(),
		tenantID:    tenantID,
		recipientID: recipientID,
		channel:     channel,
		templateKey: templateKey,
		variables:   variables,
		subject:     subject,
		body:        body,
		status:      NotificationStatusPending,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func RehydrateNotification(
	id NotificationID,
	tenantID TenantID,
	recipientID RecipientID,
	channel ChannelType,
	templateKey string,
	variables map[string]string,
	subject string,
	body string,
	status NotificationStatus,
	deliveryID string,
	deliveryError string,
	scheduledAt *time.Time,
	sentAt *time.Time,
	readAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *Notification {
	return &Notification{
		id:            id,
		tenantID:      tenantID,
		recipientID:   recipientID,
		channel:       channel,
		templateKey:   templateKey,
		variables:     variables,
		subject:       subject,
		body:          body,
		status:        status,
		deliveryID:    deliveryID,
		deliveryError: deliveryError,
		scheduledAt:   scheduledAt,
		sentAt:        sentAt,
		readAt:        readAt,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func (n *Notification) MarkSent(deliveryID string) error {
	if n.status != NotificationStatusPending {
		return fmt.Errorf("%w: cannot transition from %s to SENT", ErrInvalidStatusTransition, n.status.String())
	}
	n.status = NotificationStatusSent
	n.deliveryID = deliveryID
	n.sentAt = &[]time.Time{time.Now().UTC()}[0]
	n.updatedAt = time.Now().UTC()
	return nil
}

func (n *Notification) MarkFailed(reason string) error {
	if n.status != NotificationStatusPending {
		return fmt.Errorf("%w: cannot transition from %s to FAILED", ErrInvalidStatusTransition, n.status.String())
	}
	n.status = NotificationStatusFailed
	n.deliveryError = reason
	n.updatedAt = time.Now().UTC()
	return nil
}

func (n *Notification) MarkRead() error {
	if n.status != NotificationStatusSent {
		return fmt.Errorf("%w: can only mark SENT notifications as read, current status: %s", ErrInvalidStatusTransition, n.status.String())
	}
	n.status = NotificationStatusRead
	n.readAt = &[]time.Time{time.Now().UTC()}[0]
	n.updatedAt = time.Now().UTC()
	return nil
}

func (n *Notification) ID() NotificationID {
	return n.id
}

func (n *Notification) TenantID() TenantID {
	return n.tenantID
}

func (n *Notification) RecipientID() RecipientID {
	return n.recipientID
}

func (n *Notification) Channel() ChannelType {
	return n.channel
}

func (n *Notification) Status() NotificationStatus {
	return n.status
}

func (n *Notification) TemplateKey() string {
	return n.templateKey
}

func (n *Notification) Variables() map[string]string {
	return n.variables
}

func (n *Notification) Subject() string {
	return n.subject
}

func (n *Notification) Body() string {
	return n.body
}

func (n *Notification) DeliveryID() string {
	return n.deliveryID
}

func (n *Notification) DeliveryError() string {
	return n.deliveryError
}

func (n *Notification) CreatedAt() time.Time {
	return n.createdAt
}

func (n *Notification) UpdatedAt() time.Time {
	return n.updatedAt
}

func (n *Notification) SentAt() *time.Time {
	return n.sentAt
}

func (n *Notification) ReadAt() *time.Time {
	return n.readAt
}
