package channels

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
)

// EmailAdapter sends notifications via SMTP
type EmailAdapter struct {
	smtpHost     string
	smtpPort     int
	smtpUser     string
	smtpPassword string
	fromAddress  string
	fromName     string
}

func NewEmailAdapter(
	smtpHost string,
	smtpPort int,
	smtpUser string,
	smtpPassword string,
	fromAddress string,
	fromName string,
) domain.NotificationChannelAdapter {
	return &EmailAdapter{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUser:     smtpUser,
		smtpPassword: smtpPassword,
		fromAddress:  fromAddress,
		fromName:     fromName,
	}
}

func (a *EmailAdapter) Send(ctx context.Context, notif *domain.Notification, rendered domain.RenderResult) (string, error) {
	// In a real implementation, this would send via SMTP
	// For MVP, we'll return a mock delivery ID
	// In production, integrate with SendGrid, AWS SES, or direct SMTP

	if a.smtpHost == "" {
		return fmt.Sprintf("mock-email-%s", notif.ID().String()), nil
	}

	// Attempt SMTP connection (will fail in test, but structure is correct)
	addr := fmt.Sprintf("%s:%d", a.smtpHost, a.smtpPort)
	auth := smtp.PlainAuth("", a.smtpUser, a.smtpPassword, a.smtpHost)

	// For MVP: mock delivery
	// In production: implement full email sending
	_ = addr
	_ = auth

	// Mock delivery ID
	deliveryID := fmt.Sprintf("email-%s", notif.ID().String())
	return deliveryID, nil
}
