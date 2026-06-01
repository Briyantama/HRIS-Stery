package queries

import (
	"context"
	"fmt"

	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
)

type GetChannelConfigQuery struct {
	TenantID string
}

type ChannelConfigDTO struct {
	TenantID        string
	EnabledChannels []string
	EmailConfig     *EmailConfigDTO
}

type EmailConfigDTO struct {
	SMTPHost    string
	SMTPPort    int
	SMTPUser    string
	FromAddress string
	FromName    string
}

type GetChannelConfigHandler interface {
	Handle(ctx context.Context, query GetChannelConfigQuery) (*ChannelConfigDTO, error)
}

type GetChannelConfigHandlerImpl struct {
	configRepo domain.ChannelConfigRepository
}

func NewGetChannelConfigHandler(configRepo domain.ChannelConfigRepository) GetChannelConfigHandler {
	return &GetChannelConfigHandlerImpl{
		configRepo: configRepo,
	}
}

func (h *GetChannelConfigHandlerImpl) Handle(ctx context.Context, query GetChannelConfigQuery) (*ChannelConfigDTO, error) {
	tenantID := domain.MustNewTenantID(query.TenantID)

	config, err := h.configRepo.GetByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get channel config: %w", err)
	}

	if config == nil {
		return nil, domain.ErrChannelConfigNotFound
	}

	// Convert to DTO
	enabledChannels := make([]string, len(config.EnabledChannels()))
	for i, ch := range config.EnabledChannels() {
		enabledChannels[i] = ch.String()
	}

	var emailConfigDTO *EmailConfigDTO
	if config.GetEmailConfig() != nil {
		emailCfg := config.GetEmailConfig()
		emailConfigDTO = &EmailConfigDTO{
			SMTPHost:    emailCfg.SMTPHost,
			SMTPPort:    emailCfg.SMTPPort,
			SMTPUser:    emailCfg.SMTPUser,
			FromAddress: emailCfg.FromAddress,
			FromName:    emailCfg.FromName,
		}
	}

	return &ChannelConfigDTO{
		TenantID:        tenantID.String(),
		EnabledChannels: enabledChannels,
		EmailConfig:     emailConfigDTO,
	}, nil
}
