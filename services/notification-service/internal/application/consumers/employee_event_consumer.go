package consumers

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/notification-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
)

type EmployeeEventConsumer struct {
	sendHandler commands.SendNotificationHandler
	configRepo  domain.ChannelConfigRepository
}

func NewEmployeeEventConsumer(
	sendHandler commands.SendNotificationHandler,
	configRepo domain.ChannelConfigRepository,
) *EmployeeEventConsumer {
	return &EmployeeEventConsumer{
		sendHandler: sendHandler,
		configRepo:  configRepo,
	}
}

func (c *EmployeeEventConsumer) ConsumeEmployeeCreated(ctx context.Context, event *domain.EmployeeCreatedEvent) error {
	return c.sendNotification(ctx, event.TenantID, event.EmployeeID, "employee.welcome", map[string]string{
		"first_name": event.FirstName,
		"last_name":  event.LastName,
		"email":      event.Email,
	})
}

func (c *EmployeeEventConsumer) ConsumeEmployeeTerminated(ctx context.Context, event *domain.EmployeeTerminatedEvent) error {
	// Send termination notification to employee
	return c.sendNotification(ctx, event.TenantID, event.EmployeeID, "employee.termination", map[string]string{
		"first_name": event.FirstName,
		"last_name":  event.LastName,
	})
}

func (c *EmployeeEventConsumer) sendNotification(ctx context.Context, tenantID domain.TenantID, recipientID string, templateKey string, variables map[string]string) error {
	// Get channel config for tenant
	config, err := c.configRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("send employee notification: get config: %w", err)
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

func (c *EmployeeEventConsumer) getTemplateContent(templateKey string, variables map[string]string) (string, string) {
	switch templateKey {
	case "employee.welcome":
		return "Welcome to HRIS Stery",
			fmt.Sprintf("Welcome %s %s! Your account has been created. You can log in with your email: %s",
				variables["first_name"], variables["last_name"], variables["email"])
	case "employee.termination":
		return "Employment Termination Notice",
			fmt.Sprintf("Employee %s %s has been terminated from the system",
				variables["first_name"], variables["last_name"])
	default:
		return "Notification", "You have a new notification"
	}
}
