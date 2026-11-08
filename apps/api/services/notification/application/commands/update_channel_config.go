package commands

import (
	"context"
	"fmt"
	"slices"

	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
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

	// The command addresses some channels; every other enabled channel and the stored email_config
	// are kept unless the command replaces them.
	existing, err := h.configRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("update channel config: load: %w", err)
	}
	channels := slices.Clone(cmd.Channels)
	emailConfig := cmd.EmailConfig
	if existing != nil {
		for _, ch := range existing.EnabledChannels() {
			if !slices.Contains(channels, ch) {
				channels = append(channels, ch)
			}
		}
		if emailConfig == nil {
			emailConfig = existing.GetEmailConfig()
		}
	}
	config, err := domain.NewNotificationChannelConfig(tenantID, channels, emailConfig)
	if err != nil {
		return nil, fmt.Errorf("update channel config: %w", err)
	}

	if existing != nil {
		err = h.configRepo.Update(ctx, config)
	} else {
		err = h.configRepo.Create(ctx, config)
	}
	if err != nil {
		return nil, fmt.Errorf("update channel config: persist: %w", err)
	}

	enabledChannelStrs := make([]string, len(channels))
	for i, ch := range channels {
		enabledChannelStrs[i] = ch.String()
	}
	return &UpdateChannelConfigResult{
		TenantID:        tenantID.String(),
		EnabledChannels: enabledChannelStrs,
	}, nil
}
