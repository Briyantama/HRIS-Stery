package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hris-stery/hris-stery/services/notification-service/internal/application/consumers"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

type AuthConsumer struct {
	js          nats.JetStreamContext
	pool        *pgxpool.Pool
	appConsumer *consumers.AuthEventConsumer
}

func NewAuthConsumer(
	js nats.JetStreamContext,
	pool *pgxpool.Pool,
	appConsumer *consumers.AuthEventConsumer,
) *AuthConsumer {
	return &AuthConsumer{
		js:          js,
		pool:        pool,
		appConsumer: appConsumer,
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
		fmt.Printf("error unmarshaling auth event: %v\n", err)
		return
	}

	// Check idempotency
	if c.isProcessed(ctx, envelope.EventID) {
		msg.Ack()
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
		if err := json.Unmarshal(envelope.Payload, &payload); err == nil {
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
		fmt.Printf("error processing auth event %s: %v\n", envelope.EventID, err)
		return
	}

	c.markProcessed(ctx, envelope.EventID)
	msg.Ack()
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
