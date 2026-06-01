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

type EmployeeConsumer struct {
	js          nats.JetStreamContext
	pool        *pgxpool.Pool
	appConsumer *consumers.EmployeeEventConsumer
}

func NewEmployeeConsumer(
	js nats.JetStreamContext,
	pool *pgxpool.Pool,
	appConsumer *consumers.EmployeeEventConsumer,
) *EmployeeConsumer {
	return &EmployeeConsumer{
		js:          js,
		pool:        pool,
		appConsumer: appConsumer,
	}
}

func (c *EmployeeConsumer) Subscribe(ctx context.Context) error {
	_, err := c.js.Subscribe("hris.workforce.employee.>", func(msg *nats.Msg) {
		c.handleMessage(context.Background(), msg)
	})
	return err
}

func (c *EmployeeConsumer) handleMessage(ctx context.Context, msg *nats.Msg) {
	var envelope leaveEventEnvelope
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		fmt.Printf("error unmarshaling employee event: %v\n", err)
		return
	}

	// Check idempotency
	if c.isProcessed(ctx, envelope.EventID) {
		msg.Ack()
		return
	}

	var err error
	switch envelope.EventType {
	case "hris.workforce.employee.created":
		var payload struct {
			EmployeeID string `json:"employee_id"`
			Email      string `json:"email"`
			FirstName  string `json:"first_name"`
			LastName   string `json:"last_name"`
		}
		if err := json.Unmarshal(envelope.Payload, &payload); err == nil {
			event := &domain.EmployeeCreatedEvent{
				EventID:    envelope.EventID,
				TenantID:   domain.MustNewTenantID(envelope.TenantID),
				ActorID:    envelope.ActorID,
				OccurredAt: envelope.OccurredAt,
				EmployeeID: payload.EmployeeID,
				Email:      payload.Email,
				FirstName:  payload.FirstName,
				LastName:   payload.LastName,
			}
			err = c.appConsumer.ConsumeEmployeeCreated(ctx, event)
		}

	case "hris.workforce.employee.terminated":
		var payload struct {
			EmployeeID string `json:"employee_id"`
			Email      string `json:"email"`
			FirstName  string `json:"first_name"`
			LastName   string `json:"last_name"`
		}
		if err := json.Unmarshal(envelope.Payload, &payload); err == nil {
			event := &domain.EmployeeTerminatedEvent{
				EventID:    envelope.EventID,
				TenantID:   domain.MustNewTenantID(envelope.TenantID),
				ActorID:    envelope.ActorID,
				OccurredAt: envelope.OccurredAt,
				EmployeeID: payload.EmployeeID,
				Email:      payload.Email,
				FirstName:  payload.FirstName,
				LastName:   payload.LastName,
			}
			err = c.appConsumer.ConsumeEmployeeTerminated(ctx, event)
		}
	}

	if err != nil {
		fmt.Printf("error processing employee event %s: %v\n", envelope.EventID, err)
		return
	}

	c.markProcessed(ctx, envelope.EventID)
	msg.Ack()
}

func (c *EmployeeConsumer) isProcessed(ctx context.Context, eventID string) bool {
	query := "SELECT 1 FROM notification.processed_events WHERE event_id = $1"
	var exists int
	err := c.pool.QueryRow(ctx, query, eventID).Scan(&exists)
	return err == nil
}

func (c *EmployeeConsumer) markProcessed(ctx context.Context, eventID string) {
	query := `
		INSERT INTO notification.processed_events (event_id, processed_at)
		VALUES ($1, NOW())
		ON CONFLICT DO NOTHING
	`
	_, _ = c.pool.Exec(ctx, query, eventID)
}
