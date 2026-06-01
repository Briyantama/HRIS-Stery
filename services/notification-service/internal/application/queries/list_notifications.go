package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
)

type ListNotificationsQuery struct {
	TenantID    string
	RecipientID string
	Status      string
	UnreadOnly  bool
	Limit       int
	Offset      int
}

type NotificationDTO struct {
	ID          string
	RecipientID string
	Channel     string
	Status      string
	TemplateKey string
	Subject     string
	Body        string
	CreatedAt   string
	ReadAt      *string
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

	// Parse status if provided
	var status domain.NotificationStatus
	if query.Status != "" {
		parsed, err := domain.NotificationStatusFromString(query.Status)
		if err != nil {
			status = domain.NotificationStatusUnknown
		} else {
			status = parsed
		}
	}

	// Set defaults
	limit := query.Limit
	if limit == 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	filters := domain.ListFilters{
		Status:     status,
		UnreadOnly: query.UnreadOnly,
		Limit:      limit,
		Offset:     offset,
	}

	// Fetch notifications
	notifs, err := h.notificationRepo.ListByRecipient(ctx, tenantID, recipientID, filters)
	if err != nil {
		return nil, fmt.Errorf("list notifications: fetch: %w", err)
	}

	// Convert to DTOs
	dtos := make([]*NotificationDTO, len(notifs))
	for i, notif := range notifs {
		readAtStr := ""
		if notif.ReadAt() != nil {
			readAtStr = notif.ReadAt().String()
		}

		dtos[i] = &NotificationDTO{
			ID:          notif.ID().String(),
			RecipientID: notif.RecipientID().String(),
			Channel:     notif.Channel().String(),
			Status:      notif.Status().String(),
			TemplateKey: notif.TemplateKey(),
			Subject:     notif.Subject(),
			Body:        notif.Body(),
			CreatedAt:   notif.CreatedAt().String(),
			ReadAt:      &readAtStr,
		}
	}

	return &ListNotificationsResult{
		Notifications: dtos,
		Total:         len(notifs),
	}, nil
}
