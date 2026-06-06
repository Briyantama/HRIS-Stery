package nats

import (
	"context"
	"encoding/json"

	sharednats "github.com/hris-stery/hris-stery/services/_shared/nats"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/application/consumers"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type AuthConsumer struct {
	js          nats.JetStreamContext
	pool        *pgxpool.Pool
	appConsumer *consumers.AuthEventConsumer
	logger      *zap.Logger
}

func NewAuthConsumer(
	js nats.JetStreamContext,
	pool *pgxpool.Pool,
	appConsumer *consumers.AuthEventConsumer,
	logger *zap.Logger,
) *AuthConsumer {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &AuthConsumer{
		js:          js,
		pool:        pool,
		appConsumer: appConsumer,
		logger:      logger,
	}
}

func (c *AuthConsumer) Subscribe(ctx context.Context) error {
	_, err := c.js.Subscribe("hris.identity.user.>", func(msg *nats.Msg) {
		c.handleMessage(context.Background(), msg)
	})
	return err
}

func (c *AuthConsumer) handleMessage(ctx context.Context, msg *nats.Msg) {
	var envelope leaveEventEnvelope
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		c.logger.Error("unmarshaling auth event", zap.Error(err))
		sharednats.AckMessage(c.logger, msg)
		return
	}

	if c.isProcessed(ctx, envelope.EventID) {
		sharednats.AckMessage(c.logger, msg)
		return
	}

	var err error
	switch envelope.EventType {
	case "hris.identity.user.registered":
		var payload struct {
			UserID    string `json:"user_id"`
			Email     string `json:"email"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		if jsonErr := json.Unmarshal(envelope.Payload, &payload); jsonErr != nil {
			err = jsonErr
		} else {
			event := &domain.UserRegisteredEvent{
				EventID:    envelope.EventID,
				TenantID:   domain.MustNewTenantID(envelope.TenantID),
				ActorID:    envelope.ActorID,
				OccurredAt: envelope.OccurredAt,
				UserID:     payload.UserID,
				Email:      payload.Email,
				FirstName:  payload.FirstName,
				LastName:   payload.LastName,
			}
			err = c.appConsumer.ConsumeUserRegistered(ctx, event)
		}
	}

	if err != nil {
		c.logger.Error("processing auth event", zap.Error(err), zap.String("event_id", envelope.EventID))
		sharednats.NakMessage(c.logger, msg)
		return
	}

	c.markProcessed(ctx, envelope.EventID)
	sharednats.AckMessage(c.logger, msg)
}

func (c *AuthConsumer) isProcessed(ctx context.Context, eventID string) bool {
	query := "SELECT 1 FROM notification.processed_events WHERE event_id = $1"
	var exists int
	err := c.pool.QueryRow(ctx, query, eventID).Scan(&exists)
	return err == nil
}

func (c *AuthConsumer) markProcessed(ctx context.Context, eventID string) {
	query := `
		INSERT INTO notification.processed_events (event_id, processed_at)
		VALUES ($1, NOW())
		ON CONFLICT DO NOTHING
	`
	_, _ = c.pool.Exec(ctx, query, eventID)
}
