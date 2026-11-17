package consumers

import (
	"context"

	"github.com/Briyantama/HRIS-Stery/services/notification/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
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
	return sendNotification(ctx, c.sendHandler, c.configRepo, event.TenantID, event.EmployeeID, "employee.welcome", map[string]string{
		"first_name": event.FirstName,
		"last_name":  event.LastName,
		"email":      event.Email,
	})
}

func (c *EmployeeEventConsumer) ConsumeEmployeeTerminated(ctx context.Context, event *domain.EmployeeTerminatedEvent) error {
	// Send termination notification to employee
	return sendNotification(ctx, c.sendHandler, c.configRepo, event.TenantID, event.EmployeeID, "employee.termination", map[string]string{
		"first_name": event.FirstName,
		"last_name":  event.LastName,
	})
}
