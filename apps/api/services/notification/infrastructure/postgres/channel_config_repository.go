package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
	"github.com/Briyantama/stery-kit/go/pgtenant"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChannelConfigRepository struct {
	pool *pgxpool.Pool
}

func NewChannelConfigRepository(pool *pgxpool.Pool) domain.ChannelConfigRepository {
	return &ChannelConfigRepository{pool: pool}
}

// emailConfigJSON is the stored shape of notification_channel_configs.email_config.
type emailConfigJSON struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	FromAddress  string `json:"from_address"`
	FromName     string `json:"from_name"`
}

func encode(config *domain.NotificationChannelConfig) (channels []string, email []byte, err error) {
	channels = make([]string, len(config.EnabledChannels()))
	for i, ch := range config.EnabledChannels() {
		channels[i] = ch.String()
	}
	if c := config.GetEmailConfig(); c != nil {
		email, err = json.Marshal(emailConfigJSON(*c))
		if err != nil {
			return nil, nil, fmt.Errorf("marshal email config: %w", err)
		}
	}
	return channels, email, nil
}

func (r *ChannelConfigRepository) Create(ctx context.Context, config *domain.NotificationChannelConfig) error {
	channels, email, err := encode(config)
	if err != nil {
		return fmt.Errorf("create channel config: %w", err)
	}
	return pgtenant.WithTenantTx(ctx, r.pool, config.TenantID().String(), "", func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO notification.notification_channel_configs (tenant_id, channels, email_config, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)`,
			config.TenantID().String(), channels, email, config.CreatedAt(), config.UpdatedAt())
		if err != nil {
			return fmt.Errorf("create channel config: %w", err)
		}
		return nil
	})
}

// GetByTenant returns (nil, nil) when the tenant has no configuration.
func (r *ChannelConfigRepository) GetByTenant(ctx context.Context, tenantID domain.TenantID) (*domain.NotificationChannelConfig, error) {
	var config *domain.NotificationChannelConfig
	err := pgtenant.WithTenantTx(ctx, r.pool, tenantID.String(), "", func(tx pgx.Tx) error {
		var (
			channelStrs        []string
			emailRaw           []byte
			createdAt, updated time.Time
		)
		err := tx.QueryRow(ctx, `
			SELECT channels, email_config, created_at, updated_at
			FROM notification.notification_channel_configs WHERE tenant_id = $1`, tenantID.String()).
			Scan(&channelStrs, &emailRaw, &createdAt, &updated)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("get channel config: %w", err)
		}

		channels := make([]domain.ChannelType, len(channelStrs))
		for i, s := range channelStrs {
			ch, err := domain.ChannelTypeFromString(s)
			if err != nil {
				return fmt.Errorf("get channel config: %w", err)
			}
			channels[i] = ch
		}
		var email *domain.EmailConfig
		if emailRaw != nil {
			var j emailConfigJSON
			if err := json.Unmarshal(emailRaw, &j); err != nil {
				return fmt.Errorf("get channel config: decode email config: %w", err)
			}
			c := domain.EmailConfig(j)
			email = &c
		}
		config = domain.RehydrateNotificationChannelConfig(tenantID, channels, email, createdAt, updated)
		return nil
	})
	return config, err
}

func (r *ChannelConfigRepository) Update(ctx context.Context, config *domain.NotificationChannelConfig) error {
	channels, email, err := encode(config)
	if err != nil {
		return fmt.Errorf("update channel config: %w", err)
	}
	return pgtenant.WithTenantTx(ctx, r.pool, config.TenantID().String(), "", func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE notification.notification_channel_configs
			SET channels = $1, email_config = $2, updated_at = $3 WHERE tenant_id = $4`,
			channels, email, time.Now().UTC(), config.TenantID().String())
		if err != nil {
			return fmt.Errorf("update channel config: %w", err)
		}
		return nil
	})
}
