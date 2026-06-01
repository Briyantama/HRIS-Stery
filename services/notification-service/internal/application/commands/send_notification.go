package commands

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
)

type SendNotificationCommand struct {
	TenantID    string
	RecipientID string
	TemplateKey string
	Variables   map[string]string
	Channels    []domain.ChannelType
	Subject     string
	Body        string
}

type SendNotificationResult struct {
	NotificationIDs []string
}

type SendNotificationHandler interface {
	Handle(ctx context.Context, cmd SendNotificationCommand) (*SendNotificationResult, error)
}

type SendNotificationHandlerImpl struct {
	notificationRepo domain.NotificationRepository
}

func NewSendNotificationHandler(notificationRepo domain.NotificationRepository) SendNotificationHandler {
	return &SendNotificationHandlerImpl{
		notificationRepo: notificationRepo,
	}
}

func (h *SendNotificationHandlerImpl) Handle(ctx context.Context, cmd SendNotificationCommand) (*SendNotificationResult, error) {
	tenantID := domain.MustNewTenantID(cmd.TenantID)
	recipientID, err := domain.NewRecipientID(cmd.RecipientID)
	if err != nil {
		return nil, fmt.Errorf("send notification: %w", err)
	}

	result := &SendNotificationResult{
		NotificationIDs: []string{},
	}

	// Create a notification for each channel
	for _, channel := range cmd.Channels {
		notif, err := domain.NewNotification(
			tenantID,
			recipientID,
			channel,
			cmd.TemplateKey,
			cmd.Variables,
			cmd.Subject,
			cmd.Body,
		)
		if err != nil {
			return nil, fmt.Errorf("send notification: create aggregate: %w", err)
		}

		if err := h.notificationRepo.Create(ctx, notif); err != nil {
			return nil, fmt.Errorf("send notification: persist: %w", err)
		}

		result.NotificationIDs = append(result.NotificationIDs, notif.ID().String())
	}

	return result, nil
}
