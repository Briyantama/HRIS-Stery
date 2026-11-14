package consumers

import (
	"context"

	"github.com/Briyantama/HRIS-Stery/services/notification/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
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
	return sendNotification(ctx, c.sendHandler, c.configRepo, event.TenantID, event.ApproverId, "leave.requested", map[string]string{
		"employee_id": event.EmployeeID,
		"leave_type":  event.LeaveType,
		"start_date":  event.StartDate,
		"end_date":    event.EndDate,
	})
}

func (c *LeaveEventConsumer) ConsumeLeaveApproved(ctx context.Context, event *domain.LeaveApprovedEvent) error {
	return sendNotification(ctx, c.sendHandler, c.configRepo, event.TenantID, event.EmployeeID, "leave.approved", map[string]string{
		"leave_type": event.LeaveType,
		"start_date": event.StartDate,
		"end_date":   event.EndDate,
	})
}

func (c *LeaveEventConsumer) ConsumeLeaveRejected(ctx context.Context, event *domain.LeaveRejectedEvent) error {
	return sendNotification(ctx, c.sendHandler, c.configRepo, event.TenantID, event.EmployeeID, "leave.rejected", map[string]string{
		"leave_type": event.LeaveType,
		"start_date": event.StartDate,
		"end_date":   event.EndDate,
		"reason":     event.Reason,
	})
}
