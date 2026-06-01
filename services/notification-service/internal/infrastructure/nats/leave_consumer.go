package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hris-stery/hris-stery/services/notification-service/internal/application/consumers"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

type LeaveConsumer struct {
	js          nats.JetStreamContext
	pool        *pgxpool.Pool
	appConsumer *consumers.LeaveEventConsumer
}

func NewLeaveConsumer(
	js nats.JetStreamContext,
	pool *pgxpool.Pool,
	appConsumer *consumers.LeaveEventConsumer,
) *LeaveConsumer {
	return &LeaveConsumer{
		js:          js,
		pool:        pool,
		appConsumer: appConsumer,
	}
}

type leaveEventEnvelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	SchemaVersion int             `json:"schema_version"`
	TenantID      string          `json:"tenant_id"`
	ActorID       string          `json:"actor_id"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Payload       json.RawMessage `json:"payload"`
}

func (c *LeaveConsumer) Subscribe(ctx context.Context) error {
	_, err := c.js.Subscribe("hris.operations.leave.>", func(msg *nats.Msg) {
		c.handleMessage(context.Background(), msg)
	})
	return err
}

func (c *LeaveConsumer) handleMessage(ctx context.Context, msg *nats.Msg) {
	var envelope leaveEventEnvelope
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		fmt.Printf("error unmarshaling leave event: %v\n", err)
		return
	}

	// Check idempotency
	if c.isProcessed(ctx, envelope.EventID) {
		msg.Ack()
		return
	}

	// Route to appropriate handler based on event type
	var err error
	switch envelope.EventType {
	case "hris.operations.leave.requested":
		var payload struct {
			EmployeeID string `json:"employee_id"`
			LeaveType  string `json:"leave_type"`
			StartDate  string `json:"start_date"`
			EndDate    string `json:"end_date"`
			ApproverId string `json:"approver_id"`
		}
		if err := json.Unmarshal(envelope.Payload, &payload); err == nil {
			event := &domain.LeaveRequestedEvent{
				EventID:    envelope.EventID,
				TenantID:   domain.MustNewTenantID(envelope.TenantID),
				ActorID:    envelope.ActorID,
				OccurredAt: envelope.OccurredAt,
				EmployeeID: payload.EmployeeID,
				LeaveType:  payload.LeaveType,
				StartDate:  payload.StartDate,
				EndDate:    payload.EndDate,
				ApproverId: payload.ApproverId,
			}
			err = c.appConsumer.ConsumeLeaveRequested(ctx, event)
		}

	case "hris.operations.leave.approved":
		var payload struct {
			EmployeeID string `json:"employee_id"`
			LeaveType  string `json:"leave_type"`
			StartDate  string `json:"start_date"`
			EndDate    string `json:"end_date"`
			ApproverId string `json:"approver_id"`
		}
		if err := json.Unmarshal(envelope.Payload, &payload); err == nil {
			event := &domain.LeaveApprovedEvent{
				EventID:    envelope.EventID,
				TenantID:   domain.MustNewTenantID(envelope.TenantID),
				ActorID:    envelope.ActorID,
				OccurredAt: envelope.OccurredAt,
				EmployeeID: payload.EmployeeID,
				LeaveType:  payload.LeaveType,
				StartDate:  payload.StartDate,
				EndDate:    payload.EndDate,
				ApproverId: payload.ApproverId,
			}
			err = c.appConsumer.ConsumeLeaveApproved(ctx, event)
		}

	case "hris.operations.leave.rejected":
		var payload struct {
			EmployeeID string `json:"employee_id"`
			LeaveType  string `json:"leave_type"`
			StartDate  string `json:"start_date"`
			EndDate    string `json:"end_date"`
			Reason     string `json:"reason"`
		}
		if err := json.Unmarshal(envelope.Payload, &payload); err == nil {
			event := &domain.LeaveRejectedEvent{
				EventID:    envelope.EventID,
				TenantID:   domain.MustNewTenantID(envelope.TenantID),
				ActorID:    envelope.ActorID,
				OccurredAt: envelope.OccurredAt,
				EmployeeID: payload.EmployeeID,
				LeaveType:  payload.LeaveType,
				StartDate:  payload.StartDate,
				EndDate:    payload.EndDate,
				Reason:     payload.Reason,
			}
			err = c.appConsumer.ConsumeLeaveRejected(ctx, event)
		}
	}

	if err != nil {
		fmt.Printf("error processing leave event %s: %v\n", envelope.EventID, err)
		// Don't ack - let NATS retry
		return
	}

	// Mark as processed
	c.markProcessed(ctx, envelope.EventID)
	msg.Ack()
}

func (c *LeaveConsumer) isProcessed(ctx context.Context, eventID string) bool {
	query := "SELECT 1 FROM notification.processed_events WHERE event_id = $1"
	var exists int
	err := c.pool.QueryRow(ctx, query, eventID).Scan(&exists)
	return err == nil
}

func (c *LeaveConsumer) markProcessed(ctx context.Context, eventID string) {
	query := `
		INSERT INTO notification.processed_events (event_id, processed_at)
		VALUES ($1, NOW())
		ON CONFLICT DO NOTHING
	`
	_, _ = c.pool.Exec(ctx, query, eventID)
}
