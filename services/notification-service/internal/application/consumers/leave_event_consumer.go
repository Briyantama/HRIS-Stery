package consumers

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/notification-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
)

type LeaveEventConsumer struct {
	sendHandler commands.SendNotificationHandler
	configRepo  domain.ChannelConfigRepository
}

func NewLeaveEventConsumer(
	sendHandler commands.SendNotificationHandler,
	configRepo domain.ChannelConfigRepository,
) *LeaveEventConsumer {
	return &LeaveEventConsumer{
		sendHandler: sendHandler,
		configRepo:  configRepo,
	}
}

func (c *LeaveEventConsumer) ConsumeLeaveRequested(ctx context.Context, event *domain.LeaveRequestedEvent) error {
	return c.sendNotification(ctx, event.TenantID, event.ApproverId, "leave.requested", map[string]string{
		"employee_id": event.EmployeeID,
		"leave_type":  event.LeaveType,
		"start_date":  event.StartDate,
		"end_date":    event.EndDate,
	})
}

func (c *LeaveEventConsumer) ConsumeLeaveApproved(ctx context.Context, event *domain.LeaveApprovedEvent) error {
	return c.sendNotification(ctx, event.TenantID, event.EmployeeID, "leave.approved", map[string]string{
		"leave_type": event.LeaveType,
		"start_date": event.StartDate,
		"end_date":   event.EndDate,
	})
}

func (c *LeaveEventConsumer) ConsumeLeaveRejected(ctx context.Context, event *domain.LeaveRejectedEvent) error {
	return c.sendNotification(ctx, event.TenantID, event.EmployeeID, "leave.rejected", map[string]string{
		"leave_type": event.LeaveType,
		"start_date": event.StartDate,
		"end_date":   event.EndDate,
		"reason":     event.Reason,
	})
}

func (c *LeaveEventConsumer) sendNotification(ctx context.Context, tenantID domain.TenantID, recipientID string, templateKey string, variables map[string]string) error {
	// Get channel config for tenant
	config, err := c.configRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("send leave notification: get config: %w", err)
	}

	if config == nil {
		// Skip if no config (tenant not configured for notifications)
		return nil
	}

	// Get enabled channels
	channels := config.EnabledChannels()
	if len(channels) == 0 {
		return nil
	}

	// Map template key to subject/body (simplified - in production, use template engine)
	subject, body := c.getTemplateContent(templateKey, variables)

	// Send notification
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

func (c *LeaveEventConsumer) getTemplateContent(templateKey string, variables map[string]string) (string, string) {
	// Simplified template mapping - in production, use proper template engine
	switch templateKey {
	case "leave.requested":
		return "Leave Request Submitted",
			fmt.Sprintf("Leave request for %s to %s has been submitted", variables["start_date"], variables["end_date"])
	case "leave.approved":
		return "Leave Approved",
			fmt.Sprintf("Your leave from %s to %s has been approved", variables["start_date"], variables["end_date"])
	case "leave.rejected":
		return "Leave Rejected",
			fmt.Sprintf("Your leave from %s to %s has been rejected. Reason: %s", variables["start_date"], variables["end_date"], variables["reason"])
	default:
		return "Notification", "You have a new notification"
	}
}
