package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
)

type GetDeliveryStatusQuery struct {
	TenantID       string
	NotificationID string
}

type DeliveryStatusDTO struct {
	NotificationID string
	RecipientID    string
	Channel        string
	TemplateKey    string
	Subject        string
	Body           string
	Status         string
	DeliveryID     *string
	DeliveryError  *string
	CreatedAt      string
	SentAt         *string
	ReadAt         *string
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
		s := sentAt.String()
		sentAtStr = &s
	}

	var readAtStr *string
	if readAt != nil {
		r := readAt.String()
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
		CreatedAt:      notif.CreatedAt().String(),
		SentAt:         sentAtStr,
		ReadAt:         readAtStr,
	}, nil
}
