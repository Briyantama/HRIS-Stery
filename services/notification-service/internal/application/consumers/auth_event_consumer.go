package consumers

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/notification-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
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
	return c.sendNotification(ctx, event.TenantID, event.UserID, "auth.activation", map[string]string{
		"first_name": event.FirstName,
		"last_name":  event.LastName,
		"email":      event.Email,
	})
}

func (c *AuthEventConsumer) sendNotification(ctx context.Context, tenantID domain.TenantID, recipientID string, templateKey string, variables map[string]string) error {
	// Get channel config for tenant
	config, err := c.configRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("send auth notification: get config: %w", err)
	}

	if config == nil {
		// Skip if no config
		return nil
	}

	channels := config.EnabledChannels()
	if len(channels) == 0 {
		return nil
	}

	subject, body := c.getTemplateContent(templateKey, variables)

	cmd := commands.SendNotificationCommand{
		TenantID:    tenantID.String(),
		RecipientID: recipientID,
		TemplateKey: templateKey,
		Variables:   variables,
		Channels:    channels,
		Subject:     subject,
		Body:        body,
	}

	_, err = c.sendHandler.Handle(ctx, cmd)
	return err
}

func (c *AuthEventConsumer) getTemplateContent(templateKey string, variables map[string]string) (string, string) {
	switch templateKey {
	case "auth.activation":
		return "Account Activation",
			fmt.Sprintf("Welcome %s %s! Your account has been created. Please verify your email: %s",
				variables["first_name"], variables["last_name"], variables["email"])
	default:
		return "Notification", "You have a new notification"
	}
}
