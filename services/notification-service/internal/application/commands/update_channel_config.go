package commands

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
)

type UpdateChannelConfigCommand struct {
	TenantID    string
	Channels    []domain.ChannelType
	EmailConfig *domain.EmailConfig
}

type UpdateChannelConfigResult struct {
	TenantID        string
	EnabledChannels []string
}

type UpdateChannelConfigHandler interface {
	Handle(ctx context.Context, cmd UpdateChannelConfigCommand) (*UpdateChannelConfigResult, error)
}

type UpdateChannelConfigHandlerImpl struct {
	configRepo domain.ChannelConfigRepository
}

func NewUpdateChannelConfigHandler(configRepo domain.ChannelConfigRepository) UpdateChannelConfigHandler {
	return &UpdateChannelConfigHandlerImpl{
		configRepo: configRepo,
	}
}

func (h *UpdateChannelConfigHandlerImpl) Handle(ctx context.Context, cmd UpdateChannelConfigCommand) (*UpdateChannelConfigResult, error) {
	tenantID := domain.MustNewTenantID(cmd.TenantID)

	// Create or update config
	config, err := domain.NewNotificationChannelConfig(
		tenantID,
		cmd.Channels,
		cmd.EmailConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("update channel config: %w", err)
	}

	// Try to get existing config
	existing, _ := h.configRepo.GetByTenant(ctx, tenantID)
	if existing != nil {
		// Update existing
		if err := h.configRepo.Update(ctx, config); err != nil {
			return nil, fmt.Errorf("update channel config: persist: %w", err)
		}
	} else {
		// Create new
		if err := h.configRepo.Create(ctx, config); err != nil {
			return nil, fmt.Errorf("update channel config: persist: %w", err)
		}
	}

	// Build result
	enabledChannelStrs := make([]string, len(cmd.Channels))
	for i, ch := range cmd.Channels {
		enabledChannelStrs[i] = ch.String()
	}

	return &UpdateChannelConfigResult{
		TenantID:        tenantID.String(),
		EnabledChannels: enabledChannelStrs,
	}, nil
}
