package consumers

import (
	"context"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/notification/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
)

// sendNotification sends one notification over the tenant's enabled channels.
// It does nothing when the tenant has no channel config or no enabled channel.
func sendNotification(ctx context.Context, sendHandler commands.SendNotificationHandler, configRepo domain.ChannelConfigRepository, tenantID domain.TenantID, recipientID, templateKey string, variables map[string]string) error {
	config, err := configRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("send notification: get config: %w", err)
	}
	if config == nil {
		return nil
	}
	channels := config.EnabledChannels()
	if len(channels) == 0 {
		return nil
	}

	subject, body := templateContent(templateKey, variables)
	_, err = sendHandler.Handle(ctx, commands.SendNotificationCommand{
		TenantID:    tenantID.String(),
		RecipientID: recipientID,
		TemplateKey: templateKey,
		Variables:   variables,
		Channels:    channels,
		Subject:     subject,
		Body:        body,
	})
	return err
}

// templateContent maps a template key to a plain subject and body.
func templateContent(templateKey string, v map[string]string) (string, string) {
	switch templateKey {
	case "auth.activation":
		return "Account Activation",
			fmt.Sprintf("Welcome %s %s! Your account has been created. Please verify your email: %s",
				v["first_name"], v["last_name"], v["email"])
	case "employee.welcome":
		return "Welcome to HRIS Stery",
			fmt.Sprintf("Welcome %s %s! Your account has been created. You can log in with your email: %s",
				v["first_name"], v["last_name"], v["email"])
	case "employee.termination":
		return "Employment Termination Notice",
			fmt.Sprintf("Employee %s %s has been terminated from the system",
				v["first_name"], v["last_name"])
	case "leave.requested":
		return "Leave Request Submitted",
			fmt.Sprintf("Leave request for %s to %s has been submitted", v["start_date"], v["end_date"])
	case "leave.approved":
		return "Leave Approved",
			fmt.Sprintf("Your leave from %s to %s has been approved", v["start_date"], v["end_date"])
	case "leave.rejected":
		return "Leave Rejected",
			fmt.Sprintf("Your leave from %s to %s has been rejected. Reason: %s", v["start_date"], v["end_date"], v["reason"])
	default:
		return "Notification", "You have a new notification"
	}
}
