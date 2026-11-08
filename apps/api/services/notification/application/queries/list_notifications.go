package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
)

type ListNotificationsQuery struct {
	TenantID    string
	RecipientID string
	UnreadOnly  bool
	Limit       int
	Offset      int
}

type NotificationDTO struct {
	ID          string  `json:"id"`
	RecipientID string  `json:"recipient_id"`
	Channel     string  `json:"channel"`
	Status      string  `json:"status"`
	TemplateKey string  `json:"template_key"`
	Subject     string  `json:"subject"`
	Body        string  `json:"body"`
	CreatedAt   string  `json:"created_at"`
	ReadAt      *string `json:"read_at"`
}

type ListNotificationsResult struct {
	Notifications []*NotificationDTO
	Total         int
}

type ListNotificationsHandler interface {
	Handle(ctx context.Context, query ListNotificationsQuery) (*ListNotificationsResult, error)
}

type ListNotificationsHandlerImpl struct {
	notificationRepo domain.NotificationRepository
}

func NewListNotificationsHandler(notificationRepo domain.NotificationRepository) ListNotificationsHandler {
	return &ListNotificationsHandlerImpl{
		notificationRepo: notificationRepo,
	}
}

func (h *ListNotificationsHandlerImpl) Handle(ctx context.Context, query ListNotificationsQuery) (*ListNotificationsResult, error) {
	tenantID := domain.MustNewTenantID(query.TenantID)
	recipientID, err := domain.NewRecipientID(query.RecipientID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	filters := domain.ListFilters{
		UnreadOnly: query.UnreadOnly,
		Limit:      query.Limit,
		Offset:     offset,
	}

	// Fetch notifications
	notifs, err := h.notificationRepo.ListByRecipient(ctx, tenantID, recipientID, filters)
	if err != nil {
		return nil, fmt.Errorf("list notifications: fetch: %w", err)
	}

	total, err := h.notificationRepo.CountByRecipient(ctx, tenantID, recipientID, domain.ListFilters{UnreadOnly: query.UnreadOnly})
	if err != nil {
		return nil, fmt.Errorf("list notifications: count: %w", err)
	}

	// Convert to DTOs
	dtos := make([]*NotificationDTO, len(notifs))
	for i, notif := range notifs {
		var readAt *string
		if t := notif.ReadAt(); t != nil {
			s := t.UTC().Format(time.RFC3339)
			readAt = &s
		}

		dtos[i] = &NotificationDTO{
			ID:          notif.ID().String(),
			RecipientID: notif.RecipientID().String(),
			Channel:     notif.Channel().String(),
			Status:      notif.Status().String(),
			TemplateKey: notif.TemplateKey(),
			Subject:     notif.Subject(),
			Body:        notif.Body(),
			CreatedAt:   notif.CreatedAt().UTC().Format(time.RFC3339),
			ReadAt:      readAt,
		}
	}

	return &ListNotificationsResult{
		Notifications: dtos,
		Total:         total,
	}, nil
}
