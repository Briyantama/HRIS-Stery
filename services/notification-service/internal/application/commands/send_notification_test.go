package commands

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockNotificationRepository implements domain.NotificationRepository for testing
type MockNotificationRepository struct {
	createdNotifications []*domain.Notification
}

func (m *MockNotificationRepository) Create(ctx context.Context, notif *domain.Notification) error {
	m.createdNotifications = append(m.createdNotifications, notif)
	return nil
}

func (m *MockNotificationRepository) GetByID(ctx context.Context, tenantID domain.TenantID, id domain.NotificationID) (*domain.Notification, error) {
	return nil, nil
}

func (m *MockNotificationRepository) ListByRecipient(ctx context.Context, tenantID domain.TenantID, recipientID domain.RecipientID, filters domain.ListFilters) ([]*domain.Notification, error) {
	return nil, nil
}

func (m *MockNotificationRepository) Update(ctx context.Context, notif *domain.Notification) error {
	return nil
}

func TestSendNotificationHandler_ValidCommand(t *testing.T) {
	mockRepo := &MockNotificationRepository{}
	handler := NewSendNotificationHandler(mockRepo)

	tenantID := uuid.New().String()
	recipientID := uuid.New().String()

	cmd := SendNotificationCommand{
		TenantID:    tenantID,
		RecipientID: recipientID,
		TemplateKey: "leave.requested",
		Variables: map[string]string{
			"employee_name": "John Doe",
		},
		Channels: []domain.ChannelType{domain.ChannelTypeEmail, domain.ChannelTypeInApp},
		Subject:  "Leave Request",
		Body:     "Your leave request has been submitted",
	}

	result, err := handler.Handle(context.Background(), cmd)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.NotificationIDs))
	assert.Equal(t, 2, len(mockRepo.createdNotifications))
}

func TestSendNotificationHandler_InvalidRecipientID(t *testing.T) {
	mockRepo := &MockNotificationRepository{}
	handler := NewSendNotificationHandler(mockRepo)

	cmd := SendNotificationCommand{
		TenantID:    uuid.New().String(),
		RecipientID: "invalid-uuid",
		TemplateKey: "leave.requested",
		Channels:    []domain.ChannelType{domain.ChannelTypeEmail},
	}

	_, err := handler.Handle(context.Background(), cmd)

	assert.Error(t, err)
}

func TestSendNotificationHandler_NoChannels(t *testing.T) {
	mockRepo := &MockNotificationRepository{}
	handler := NewSendNotificationHandler(mockRepo)

	cmd := SendNotificationCommand{
		TenantID:    uuid.New().String(),
		RecipientID: uuid.New().String(),
		TemplateKey: "leave.requested",
		Channels:    []domain.ChannelType{},
	}

	result, err := handler.Handle(context.Background(), cmd)

	require.NoError(t, err)
	assert.Equal(t, 0, len(result.NotificationIDs))
}
