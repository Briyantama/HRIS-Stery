package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sharedpostgres "github.com/hris-stery/hris-stery/services/_shared/postgres"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChannelConfigRepository struct {
	pool *pgxpool.Pool
}

func NewChannelConfigRepository(pool *pgxpool.Pool) domain.ChannelConfigRepository {
	return &ChannelConfigRepository{
		pool: pool,
	}
}

func (r *ChannelConfigRepository) Create(ctx context.Context, config *domain.NotificationChannelConfig) error {
	return sharedpostgres.WithTenantTx(ctx, r.pool, sharedpostgres.TenantID(config.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		channels := make([]string, len(config.EnabledChannels()))
		for i, ch := range config.EnabledChannels() {
			channels[i] = ch.String()
		}

		var emailConfigJSON []byte
		if config.GetEmailConfig() != nil {
			emailCfg := config.GetEmailConfig()
			data := map[string]interface{}{
				"smtp_host":     emailCfg.SMTPHost,
				"smtp_port":     emailCfg.SMTPPort,
				"smtp_user":     emailCfg.SMTPUser,
				"smtp_password": emailCfg.SMTPPassword,
				"from_address":  emailCfg.FromAddress,
				"from_name":     emailCfg.FromName,
			}
			emailConfigJSON, _ = json.Marshal(data)
		}

		query := `
			INSERT INTO notification.notification_channel_configs
			(tenant_id, channels, email_config, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`

		_, err := tx.Exec(ctx, query,
			config.TenantID().String(),
			channels,
			emailConfigJSON,
			config.CreatedAt(),
			config.UpdatedAt(),
		)

		return fmt.Errorf("create channel config: %w", err)
	})
}

func (r *ChannelConfigRepository) GetByTenant(ctx context.Context, tenantID domain.TenantID) (*domain.NotificationChannelConfig, error) {
	var config *domain.NotificationChannelConfig
	var err error

	innerErr := sharedpostgres.WithTenantTx(ctx, r.pool, sharedpostgres.TenantID(tenantID.String()), func(ctx context.Context, tx pgx.Tx) error {
		config, err = r.getByTenantFromTx(ctx, tx, tenantID)
		return err
	})

	return config, innerErr
}

func (r *ChannelConfigRepository) getByTenantFromTx(ctx context.Context, tx pgx.Tx, tenantID domain.TenantID) (*domain.NotificationChannelConfig, error) {
	query := `
		SELECT tenant_id, channels, email_config, created_at, updated_at
		FROM notification.notification_channel_configs
		WHERE tenant_id = $1
	`

	var (
		channelStrs     []string
		emailConfigJSON []byte
		createdAt       time.Time
		updatedAt       time.Time
	)

	err := tx.QueryRow(ctx, query, tenantID.String()).Scan(
		&tenantID, &channelStrs, &emailConfigJSON, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get channel config: %w", err)
	}

	// Parse channels
	channels := make([]domain.ChannelType, len(channelStrs))
	for i, chStr := range channelStrs {
		ch, _ := domain.ChannelTypeFromString(chStr)
		channels[i] = ch
	}

	// Parse email config
	var emailConfig *domain.EmailConfig
	if emailConfigJSON != nil {
		var data map[string]interface{}
		if err := json.Unmarshal(emailConfigJSON, &data); err == nil {
			emailConfig = &domain.EmailConfig{
				SMTPHost:     data["smtp_host"].(string),
				SMTPPort:     int(data["smtp_port"].(float64)),
				SMTPUser:     data["smtp_user"].(string),
				SMTPPassword: data["smtp_password"].(string),
				FromAddress:  data["from_address"].(string),
				FromName:     data["from_name"].(string),
			}
		}
	}

	return domain.RehydrateNotificationChannelConfig(
		tenantID,
		channels,
		emailConfig,
		createdAt,
		updatedAt,
	), nil
}

func (r *ChannelConfigRepository) Update(ctx context.Context, config *domain.NotificationChannelConfig) error {
	return sharedpostgres.WithTenantTx(ctx, r.pool, sharedpostgres.TenantID(config.TenantID().String()), func(ctx context.Context, tx pgx.Tx) error {
		channels := make([]string, len(config.EnabledChannels()))
		for i, ch := range config.EnabledChannels() {
			channels[i] = ch.String()
		}

		var emailConfigJSON []byte
		if config.GetEmailConfig() != nil {
			emailCfg := config.GetEmailConfig()
			data := map[string]interface{}{
				"smtp_host":     emailCfg.SMTPHost,
				"smtp_port":     emailCfg.SMTPPort,
				"smtp_user":     emailCfg.SMTPUser,
				"smtp_password": emailCfg.SMTPPassword,
				"from_address":  emailCfg.FromAddress,
				"from_name":     emailCfg.FromName,
			}
			emailConfigJSON, _ = json.Marshal(data)
		}

		query := `
			UPDATE notification.notification_channel_configs
			SET channels = $1, email_config = $2, updated_at = $3
			WHERE tenant_id = $4
		`

		_, err := tx.Exec(ctx, query,
			channels,
			emailConfigJSON,
			time.Now().UTC(),
			config.TenantID().String(),
		)

		return fmt.Errorf("update channel config: %w", err)
	})
}
