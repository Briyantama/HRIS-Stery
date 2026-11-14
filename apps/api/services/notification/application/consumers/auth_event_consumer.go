package consumers

import (
	"context"

	"github.com/Briyantama/HRIS-Stery/services/notification/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
)

type AuthEventConsumer struct {
	sendHandler commands.SendNotificationHandler
	configRepo  domain.ChannelConfigRepository
}

func NewAuthEventConsumer(
	sendHandler commands.SendNotificationHandler,
	configRepo domain.ChannelConfigRepository,
) *AuthEventConsumer {
	return &AuthEventConsumer{
		sendHandler: sendHandler,
		configRepo:  configRepo,
	}
}

func (c *AuthEventConsumer) ConsumeUserRegistered(ctx context.Context, event *domain.UserRegisteredEvent) error {
	return sendNotification(ctx, c.sendHandler, c.configRepo, event.TenantID, event.UserID, "auth.activation", map[string]string{
		"first_name": event.FirstName,
		"last_name":  event.LastName,
		"email":      event.Email,
	})
}
