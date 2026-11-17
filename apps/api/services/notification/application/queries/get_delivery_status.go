package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
)

type GetDeliveryStatusQuery struct {
	TenantID       string
	NotificationID string
}

type DeliveryStatusDTO struct {
	NotificationID string  `json:"notification_id"`
	RecipientID    string  `json:"recipient_id"`
	Channel        string  `json:"channel"`
	TemplateKey    string  `json:"template_key"`
	Subject        string  `json:"subject"`
	Body           string  `json:"body"`
	Status         string  `json:"status"`
	DeliveryID     *string `json:"delivery_id"`
	DeliveryError  *string `json:"delivery_error"`
	CreatedAt      string  `json:"created_at"`
	SentAt         *string `json:"sent_at"`
	ReadAt         *string `json:"read_at"`
}

type GetDeliveryStatusHandler interface {
	Handle(ctx context.Context, query GetDeliveryStatusQuery) (*DeliveryStatusDTO, error)
}

type GetDeliveryStatusHandlerImpl struct {
	notificationRepo domain.NotificationRepository
}

func NewGetDeliveryStatusHandler(notificationRepo domain.NotificationRepository) GetDeliveryStatusHandler {
	return &GetDeliveryStatusHandlerImpl{
		notificationRepo: notificationRepo,
	}
}

func (h *GetDeliveryStatusHandlerImpl) Handle(ctx context.Context, query GetDeliveryStatusQuery) (*DeliveryStatusDTO, error) {
	tenantID := domain.MustNewTenantID(query.TenantID)
	notifID, err := domain.NewNotificationID(query.NotificationID)
	if err != nil {
		return nil, fmt.Errorf("get delivery status: invalid notification_id: %w", err)
	}

	notif, err := h.notificationRepo.GetByID(ctx, tenantID, notifID)
	if err != nil {
		return nil, fmt.Errorf("get delivery status: fetch: %w", err)
	}

	if notif == nil {
		return nil, domain.ErrNotificationNotFound
	}

	// Convert to DTO
	deliveryID := notif.DeliveryID()
	deliveryError := notif.DeliveryError()
	sentAt := notif.SentAt()
	readAt := notif.ReadAt()

	var deliveryIDStr *string
	if deliveryID != "" {
		deliveryIDStr = &deliveryID
	}

	var deliveryErrorStr *string
	if deliveryError != "" {
		deliveryErrorStr = &deliveryError
	}

	var sentAtStr *string
	if sentAt != nil {
		s := sentAt.UTC().Format(time.RFC3339)
		sentAtStr = &s
	}

	var readAtStr *string
	if readAt != nil {
		r := readAt.UTC().Format(time.RFC3339)
		readAtStr = &r
	}

	return &DeliveryStatusDTO{
		NotificationID: notif.ID().String(),
		RecipientID:    notif.RecipientID().String(),
		Channel:        notif.Channel().String(),
		TemplateKey:    notif.TemplateKey(),
		Subject:        notif.Subject(),
		Body:           notif.Body(),
		Status:         notif.Status().String(),
		DeliveryID:     deliveryIDStr,
		DeliveryError:  deliveryErrorStr,
		CreatedAt:      notif.CreatedAt().UTC().Format(time.RFC3339),
		SentAt:         sentAtStr,
		ReadAt:         readAtStr,
	}, nil
}
