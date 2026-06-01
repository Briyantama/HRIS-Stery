package consumers

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSendNotificationHandler for testing
type MockSendNotificationHandler struct {
	handledCommands []commands.SendNotificationCommand
}

func (m *MockSendNotificationHandler) Handle(ctx context.Context, cmd commands.SendNotificationCommand) (*commands.SendNotificationResult, error) {
	m.handledCommands = append(m.handledCommands, cmd)
	return &commands.SendNotificationResult{
		NotificationIDs: []string{uuid.New().String()},
	}, nil
}

// MockChannelConfigRepository for testing
type MockChannelConfigRepository struct {
	configs map[string]*domain.NotificationChannelConfig
}

func (m *MockChannelConfigRepository) Create(ctx context.Context, config *domain.NotificationChannelConfig) error {
	m.configs[config.TenantID().String()] = config
	return nil
}

func (m *MockChannelConfigRepository) GetByTenant(ctx context.Context, tenantID domain.TenantID) (*domain.NotificationChannelConfig, error) {
	return m.configs[tenantID.String()], nil
}

func (m *MockChannelConfigRepository) Update(ctx context.Context, config *domain.NotificationChannelConfig) error {
	m.configs[config.TenantID().String()] = config
	return nil
}

func TestLeaveEventConsumer_ConsumeLeaveRequested(t *testing.T) {
	// Setup
	mockSendHandler := &MockSendNotificationHandler{}
	mockConfigRepo := &MockChannelConfigRepository{
		configs: make(map[string]*domain.NotificationChannelConfig),
	}

	tenantID := domain.MustNewTenantID(uuid.New().String())
	emailConfig := &domain.EmailConfig{
		SMTPHost:    "smtp.example.com",
		SMTPPort:    587,
		FromAddress: "noreply@example.com",
		FromName:    "HRIS",
	}
	config, _ := domain.NewNotificationChannelConfig(tenantID, []domain.ChannelType{domain.ChannelTypeEmail}, emailConfig)
	mockConfigRepo.Create(context.Background(), config)

	consumer := NewLeaveEventConsumer(mockSendHandler, mockConfigRepo)

	// Create event
	event := &domain.LeaveRequestedEvent{
		EventID:    uuid.New().String(),
		TenantID:   tenantID,
		ActorID:    uuid.New().String(),
		OccurredAt: time.Now(),
		EmployeeID: uuid.New().String(),
		LeaveType:  "Annual Leave",
		StartDate:  "2026-06-15",
		EndDate:    "2026-06-20",
		ApproverId: uuid.New().String(),
	}

	// Act
	err := consumer.ConsumeLeaveRequested(context.Background(), event)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, len(mockSendHandler.handledCommands))
	assert.Equal(t, "leave.requested", mockSendHandler.handledCommands[0].TemplateKey)
	assert.Equal(t, event.ApproverId, mockSendHandler.handledCommands[0].RecipientID)
}

func TestLeaveEventConsumer_ConsumeLeaveApproved(t *testing.T) {
	// Setup
	mockSendHandler := &MockSendNotificationHandler{}
	mockConfigRepo := &MockChannelConfigRepository{
		configs: make(map[string]*domain.NotificationChannelConfig),
	}

	tenantID := domain.MustNewTenantID(uuid.New().String())
	emailConfig := &domain.EmailConfig{
		SMTPHost:    "smtp.example.com",
		SMTPPort:    587,
		FromAddress: "noreply@example.com",
		FromName:    "HRIS",
	}
	config, _ := domain.NewNotificationChannelConfig(tenantID, []domain.ChannelType{domain.ChannelTypeInApp}, emailConfig)
	mockConfigRepo.Create(context.Background(), config)

	consumer := NewLeaveEventConsumer(mockSendHandler, mockConfigRepo)

	// Create event
	employeeID := uuid.New().String()
	event := &domain.LeaveApprovedEvent{
		EventID:    uuid.New().String(),
		TenantID:   tenantID,
		ActorID:    uuid.New().String(),
		OccurredAt: time.Now(),
		EmployeeID: employeeID,
		LeaveType:  "Annual Leave",
		StartDate:  "2026-06-15",
		EndDate:    "2026-06-20",
		ApproverId: uuid.New().String(),
	}

	// Act
	err := consumer.ConsumeLeaveApproved(context.Background(), event)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, len(mockSendHandler.handledCommands))
	assert.Equal(t, "leave.approved", mockSendHandler.handledCommands[0].TemplateKey)
	assert.Equal(t, employeeID, mockSendHandler.handledCommands[0].RecipientID)
}

func TestLeaveEventConsumer_ConsumeLeaveRejected(t *testing.T) {
	// Setup
	mockSendHandler := &MockSendNotificationHandler{}
	mockConfigRepo := &MockChannelConfigRepository{
		configs: make(map[string]*domain.NotificationChannelConfig),
	}

	tenantID := domain.MustNewTenantID(uuid.New().String())
	emailConfig := &domain.EmailConfig{
		SMTPHost:    "smtp.example.com",
		SMTPPort:    587,
		FromAddress: "noreply@example.com",
		FromName:    "HRIS",
	}
	config, _ := domain.NewNotificationChannelConfig(tenantID, []domain.ChannelType{domain.ChannelTypeEmail}, emailConfig)
	mockConfigRepo.Create(context.Background(), config)

	consumer := NewLeaveEventConsumer(mockSendHandler, mockConfigRepo)

	// Create event
	employeeID := uuid.New().String()
	event := &domain.LeaveRejectedEvent{
		EventID:    uuid.New().String(),
		TenantID:   tenantID,
		ActorID:    uuid.New().String(),
		OccurredAt: time.Now(),
		EmployeeID: employeeID,
		LeaveType:  "Annual Leave",
		StartDate:  "2026-06-15",
		EndDate:    "2026-06-20",
		Reason:     "Insufficient headcount",
	}

	// Act
	err := consumer.ConsumeLeaveRejected(context.Background(), event)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, len(mockSendHandler.handledCommands))
	assert.Equal(t, "leave.rejected", mockSendHandler.handledCommands[0].TemplateKey)
	assert.Equal(t, employeeID, mockSendHandler.handledCommands[0].RecipientID)
}

func TestLeaveEventConsumer_NoConfigSkipsNotification(t *testing.T) {
	// Setup - no config in repo
	mockSendHandler := &MockSendNotificationHandler{}
	mockConfigRepo := &MockChannelConfigRepository{
		configs: make(map[string]*domain.NotificationChannelConfig),
	}

	consumer := NewLeaveEventConsumer(mockSendHandler, mockConfigRepo)

	tenantID := domain.MustNewTenantID(uuid.New().String())
	event := &domain.LeaveRequestedEvent{
		EventID:    uuid.New().String(),
		TenantID:   tenantID,
		ActorID:    uuid.New().String(),
		OccurredAt: time.Now(),
		EmployeeID: uuid.New().String(),
		LeaveType:  "Annual Leave",
		StartDate:  "2026-06-15",
		EndDate:    "2026-06-20",
		ApproverId: uuid.New().String(),
	}

	// Act
	err := consumer.ConsumeLeaveRequested(context.Background(), event)

	// Assert - should not error, but should not send
	require.NoError(t, err)
	assert.Equal(t, 0, len(mockSendHandler.handledCommands))
}
