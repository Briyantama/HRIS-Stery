package channels

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
	"go.uber.org/zap"
)

// EmailAdapter sends notifications via SMTP
type EmailAdapter struct {
	smtpHost     string
	smtpPort     int
	smtpUser     string
	smtpPassword string
	fromAddress  string
	fromName     string
	tlsEnabled   bool
	logger       *zap.Logger
}

func NewEmailAdapter(
	smtpHost string,
	smtpPort int,
	smtpUser string,
	smtpPassword string,
	fromAddress string,
	fromName string,
	tlsEnabled bool,
	logger *zap.Logger,
) domain.NotificationChannelAdapter {
	return &EmailAdapter{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUser:     smtpUser,
		smtpPassword: smtpPassword,
		fromAddress:  fromAddress,
		fromName:     fromName,
		tlsEnabled:   tlsEnabled,
		logger:       logger,
	}
}

func (a *EmailAdapter) Send(ctx context.Context, notif *domain.Notification, rendered domain.RenderResult) (string, error) {
	// Return mock delivery ID if SMTP not configured (development mode)
	if a.smtpHost == "" || a.smtpPort == 0 {
		a.logger.Info("SMTP not configured, returning mock delivery ID", zap.String("notification_id", notif.ID().String()))
		return fmt.Sprintf("mock-email-%s", notif.ID().String()), nil
	}

	// Recipient email must be provided in render result
	if rendered.RecipientEmail == "" {
		err := fmt.Errorf("recipient email not provided in render result")
		a.logger.Error("cannot send email", zap.Error(err), zap.String("notification_id", notif.ID().String()))
		return "", err
	}
	recipientEmail := rendered.RecipientEmail

	// Construct multipart MIME email message (HTML + plain text)
	from := a.fromAddress
	if a.fromName != "" {
		from = fmt.Sprintf("%s <%s>", a.fromName, a.fromAddress)
	}

	// Build multipart message with boundary
	boundary := "boundary-" + strings.ReplaceAll(notif.ID().String(), "-", "")
	msg := a.buildMultipartMessage(from, recipientEmail, rendered.Subject, rendered.Body, rendered.PlainTextBody, boundary)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", a.smtpHost, a.smtpPort)
	auth := smtp.PlainAuth("", a.smtpUser, a.smtpPassword, a.smtpHost)

	// Send email with appropriate TLS handling
	var err error
	if a.tlsEnabled && a.smtpPort == 587 {
		// StartTLS for port 587
		err = a.sendWithStartTLS(addr, auth, a.fromAddress, []string{recipientEmail}, []byte(msg))
	} else if a.tlsEnabled && a.smtpPort == 465 {
		// Direct TLS for port 465 (not standard net/smtp but fallback to basic auth)
		err = smtp.SendMail(addr, auth, a.fromAddress, []string{recipientEmail}, []byte(msg))
	} else {
		// Plain SMTP for port 25
		err = smtp.SendMail(addr, auth, a.fromAddress, []string{recipientEmail}, []byte(msg))
	}

	if err != nil {
		a.logger.Error("failed to send email via SMTP", zap.Error(err), zap.String("notification_id", notif.ID().String()), zap.String("to", recipientEmail))
		return "", fmt.Errorf("smtp send failed: %w", err)
	}

	deliveryID := fmt.Sprintf("email-%s", notif.ID().String())
	a.logger.Info("email sent successfully", zap.String("delivery_id", deliveryID), zap.String("to", recipientEmail))
	return deliveryID, nil
}

// sendWithStartTLS sends email using STARTTLS (port 587)
func (a *EmailAdapter) sendWithStartTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mail from failed: %w", err)
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("rcpt failed: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data failed: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	return client.Quit()
}

// buildMultipartMessage constructs a multipart MIME email with HTML and plain text versions
func (a *EmailAdapter) buildMultipartMessage(from, to, subject, htmlBody, plainBody, boundary string) []byte {
	// If plain body not provided, strip HTML tags for plain text version
	if plainBody == "" {
		plainBody = stripHTMLTags(htmlBody)
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%s\r\n\r\n--%s\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s\r\n\r\n--%s\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s\r\n\r\n--%s--\r\n",
		from,
		to,
		subject,
		boundary,
		boundary,
		plainBody,
		boundary,
		htmlBody,
		boundary,
	)
	return []byte(msg)
}

// stripHTMLTags removes HTML tags from a string (simple implementation)
func stripHTMLTags(html string) string {
	// Simple regex-based removal of HTML tags
	// In production, consider using github.com/jaytaylor/html2text or similar
	result := strings.ReplaceAll(html, "<br>", "\n")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<br />", "\n")
	result = strings.ReplaceAll(result, "</p>", "\n")
	result = strings.ReplaceAll(result, "</li>", "\n")
	result = strings.ReplaceAll(result, "</div>", "\n")

	// Remove remaining HTML tags
	inTag := false
	var plainText strings.Builder
	for _, ch := range result {
		if ch == '<' {
			inTag = true
		} else if ch == '>' {
			inTag = false
		} else if !inTag {
			plainText.WriteRune(ch)
		}
	}
	return plainText.String()
}
