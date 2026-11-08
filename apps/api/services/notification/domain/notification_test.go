package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNotification_ValidCreation(t *testing.T) {
	tenantID := MustNewTenantID(uuid.New().String())
	recipientID, err := NewRecipientID(uuid.New().String())
	require.NoError(t, err)

	variables := map[string]string{
		"employee_name": "John Doe",
		"leave_type":    "Annual",
	}

	notif, err := NewNotification(
		tenantID,
		recipientID,
		ChannelTypeEmail,
		"leave.requested",
		variables,
		"Leave Request Submitted",
		"Your leave request has been submitted for approval.",
	)

	require.NoError(t, err)
	assert.Equal(t, recipientID, notif.RecipientID())
	assert.Equal(t, ChannelTypeEmail, notif.Channel())
	assert.Equal(t, NotificationStatusPending, notif.Status())
	assert.Equal(t, "leave.requested", notif.TemplateKey())
	assert.Equal(t, variables, notif.Variables())
}

func TestNewNotification_InvalidTenantID(t *testing.T) {
	zeroTenant := TenantID("")
	recipientID, _ := NewRecipientID(uuid.New().String())

	_, err := NewNotification(
		zeroTenant,
		recipientID,
		ChannelTypeEmail,
		"leave.requested",
		map[string]string{},
		"Test",
		"Test",
	)

	assert.Error(t, err)
}

func TestNewNotification_InvalidRecipientID(t *testing.T) {
	tenantID := MustNewTenantID(uuid.New().String())
	zeroRecipient := RecipientID{}

	_, err := NewNotification(
		tenantID,
		zeroRecipient,
		ChannelTypeEmail,
		"leave.requested",
		map[string]string{},
		"Test",
		"Test",
	)

	assert.Error(t, err)
}

func TestNewNotification_InvalidChannel(t *testing.T) {
	tenantID := MustNewTenantID(uuid.New().String())
	recipientID, _ := NewRecipientID(uuid.New().String())

	_, err := NewNotification(
		tenantID,
		recipientID,
		ChannelTypeUnknown,
		"leave.requested",
		map[string]string{},
		"Test",
		"Test",
	)

	assert.Error(t, err)
}

func TestNotification_MarkSent(t *testing.T) {
	tenantID := MustNewTenantID(uuid.New().String())
	recipientID, _ := NewRecipientID(uuid.New().String())

	notif, _ := NewNotification(
		tenantID,
		recipientID,
		ChannelTypeEmail,
		"leave.requested",
		map[string]string{},
		"Test",
		"Test",
	)

	assert.Equal(t, NotificationStatusPending, notif.Status())

	deliveryID := uuid.New().String()
	err := notif.MarkSent(deliveryID)
	require.NoError(t, err)

	assert.Equal(t, NotificationStatusSent, notif.Status())
	assert.Equal(t, deliveryID, notif.DeliveryID())
	assert.NotNil(t, notif.SentAt())
}

func TestNotification_MarkFailed(t *testing.T) {
	tenantID := MustNewTenantID(uuid.New().String())
	recipientID, _ := NewRecipientID(uuid.New().String())

	notif, _ := NewNotification(
		tenantID,
		recipientID,
		ChannelTypeEmail,
		"leave.requested",
		map[string]string{},
		"Test",
		"Test",
	)

	assert.Equal(t, NotificationStatusPending, notif.Status())

	err := notif.MarkFailed("SMTP connection failed")
	require.NoError(t, err)

	assert.Equal(t, NotificationStatusFailed, notif.Status())
	assert.Equal(t, "SMTP connection failed", notif.DeliveryError())
}

func TestNotification_MarkSentThenRead(t *testing.T) {
	tenantID := MustNewTenantID(uuid.New().String())
	recipientID, _ := NewRecipientID(uuid.New().String())

	notif, _ := NewNotification(
		tenantID,
		recipientID,
		ChannelTypeInApp,
		"leave.approved",
		map[string]string{},
		"Leave Approved",
		"Your leave has been approved.",
	)

	err := notif.MarkSent("delivery-123")
	require.NoError(t, err)
	assert.Equal(t, NotificationStatusSent, notif.Status())

	err = notif.MarkRead()
	require.NoError(t, err)
	assert.Equal(t, NotificationStatusRead, notif.Status())
	assert.NotNil(t, notif.ReadAt())
}

func TestNotification_InvalidTransition_FailedToSent(t *testing.T) {
	tenantID := MustNewTenantID(uuid.New().String())
	recipientID, _ := NewRecipientID(uuid.New().String())

	notif, _ := NewNotification(
		tenantID,
		recipientID,
		ChannelTypeEmail,
		"test",
		map[string]string{},
		"Test",
		"Test",
	)

	// Mark as failed
	err := notif.MarkFailed("error")
	require.NoError(t, err)

	// Try to mark as sent (should fail - FAILED is terminal)
	err = notif.MarkSent("delivery-123")
	assert.Error(t, err)
	assert.Equal(t, NotificationStatusFailed, notif.Status())
}

func TestNotification_InvalidTransition_PendingToRead(t *testing.T) {
	tenantID := MustNewTenantID(uuid.New().String())
	recipientID, _ := NewRecipientID(uuid.New().String())

	notif, _ := NewNotification(
		tenantID,
		recipientID,
		ChannelTypeInApp,
		"test",
		map[string]string{},
		"Test",
		"Test",
	)

	// Try to mark as read without marking as sent first (should fail)
	err := notif.MarkRead()
	assert.Error(t, err)
	assert.Equal(t, NotificationStatusPending, notif.Status())
}

func TestRehydrateNotification(t *testing.T) {
	id := GenerateNotificationID()
	tenantID := MustNewTenantID(uuid.New().String())
	recipientID, _ := NewRecipientID(uuid.New().String())
	now := time.Now().UTC()
	sentAt := now.Add(time.Minute)

	notif := RehydrateNotification(
		id,
		tenantID,
		recipientID,
		ChannelTypeEmail,
		"leave.approved",
		map[string]string{"name": "John"},
		"Approved",
		"Your leave is approved",
		NotificationStatusSent,
		"delivery-123",
		"",
		nil,
		&sentAt,
		nil,
		now,
		now,
	)

	assert.Equal(t, id, notif.ID())
	assert.Equal(t, tenantID, notif.TenantID())
	assert.Equal(t, recipientID, notif.RecipientID())
	assert.Equal(t, NotificationStatusSent, notif.Status())
	assert.Equal(t, "delivery-123", notif.DeliveryID())
}

func TestNewNotificationChannelConfig_Valid(t *testing.T) {
	tenantID := MustNewTenantID(uuid.New().String())
	emailConfig := &EmailConfig{
		SMTPHost:    "smtp.example.com",
		SMTPPort:    587,
		FromAddress: "noreply@example.com",
		FromName:    "HRIS Stery",
	}

	config, err := NewNotificationChannelConfig(
		tenantID,
		[]ChannelType{ChannelTypeEmail, ChannelTypeInApp},
		emailConfig,
	)

	require.NoError(t, err)
	assert.True(t, config.IsChannelEnabled(ChannelTypeEmail))
	assert.True(t, config.IsChannelEnabled(ChannelTypeInApp))
	assert.NotNil(t, config.GetEmailConfig())
}

func TestNewNotificationChannelConfig_NoChannelsEnabled(t *testing.T) {
	tenantID := MustNewTenantID(uuid.New().String())

	_, err := NewNotificationChannelConfig(
		tenantID,
		[]ChannelType{},
		nil,
	)

	assert.Error(t, err)
}

func TestNewNotificationChannelConfig_EmailEnabledWithoutConfig(t *testing.T) {
	tenantID := MustNewTenantID(uuid.New().String())

	_, err := NewNotificationChannelConfig(
		tenantID,
		[]ChannelType{ChannelTypeEmail},
		nil,
	)

	assert.Error(t, err)
}

func TestNotificationChannelConfig_IsChannelEnabled(t *testing.T) {
	tenantID := MustNewTenantID(uuid.New().String())
	emailConfig := &EmailConfig{
		SMTPHost:    "smtp.example.com",
		SMTPPort:    587,
		FromAddress: "noreply@example.com",
		FromName:    "HRIS Stery",
	}

	config, _ := NewNotificationChannelConfig(
		tenantID,
		[]ChannelType{ChannelTypeInApp},
		emailConfig,
	)

	assert.False(t, config.IsChannelEnabled(ChannelTypeEmail))
	assert.True(t, config.IsChannelEnabled(ChannelTypeInApp))
}

func TestChannelTypeString(t *testing.T) {
	assert.Equal(t, "EMAIL", ChannelTypeEmail.String())
	assert.Equal(t, "IN_APP", ChannelTypeInApp.String())
	assert.Equal(t, "UNKNOWN", ChannelTypeUnknown.String())
}

func TestChannelTypeFromString(t *testing.T) {
	ct, err := ChannelTypeFromString("EMAIL")
	require.NoError(t, err)
	assert.Equal(t, ChannelTypeEmail, ct)

	ct, err = ChannelTypeFromString("IN_APP")
	require.NoError(t, err)
	assert.Equal(t, ChannelTypeInApp, ct)

	_, err = ChannelTypeFromString("INVALID")
	assert.Error(t, err)
}

func TestNotificationStatusString(t *testing.T) {
	assert.Equal(t, "PENDING", NotificationStatusPending.String())
	assert.Equal(t, "SENT", NotificationStatusSent.String())
	assert.Equal(t, "FAILED", NotificationStatusFailed.String())
	assert.Equal(t, "READ", NotificationStatusRead.String())
}

func TestNotificationStatusFromString(t *testing.T) {
	status, err := NotificationStatusFromString("PENDING")
	require.NoError(t, err)
	assert.Equal(t, NotificationStatusPending, status)

	status, err = NotificationStatusFromString("SENT")
	require.NoError(t, err)
	assert.Equal(t, NotificationStatusSent, status)

	status, err = NotificationStatusFromString("READ")
	require.NoError(t, err)
	assert.Equal(t, NotificationStatusRead, status)

	_, err = NotificationStatusFromString("INVALID")
	assert.Error(t, err)
}
