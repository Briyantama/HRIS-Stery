package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sharednats "github.com/hris-stery/hris-stery/services/_shared/nats"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/application/consumers"
	"github.com/hris-stery/hris-stery/services/notification-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type LeaveConsumer struct {
	js          nats.JetStreamContext
	pool        *pgxpool.Pool
	appConsumer *consumers.LeaveEventConsumer
	logger      *zap.Logger
}

func NewLeaveConsumer(
	js nats.JetStreamContext,
	pool *pgxpool.Pool,
	appConsumer *consumers.LeaveEventConsumer,
	logger *zap.Logger,
) *LeaveConsumer {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &LeaveConsumer{
		js:          js,
		pool:        pool,
		appConsumer: appConsumer,
		logger:      logger,
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
		msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		c.handleMessage(msgCtx, msg)
	},
		nats.Durable("notification-leave-consumer"),
		nats.MaxAckPending(1000),
		nats.AckWait(30*time.Second))
	return err
}

func (c *LeaveConsumer) handleMessage(ctx context.Context, msg *nats.Msg) {
	var envelope leaveEventEnvelope
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		c.logger.Error("unmarshaling leave event", zap.Error(err))
		sharednats.AckMessage(c.logger, msg)
		return
	}

	if c.isProcessed(ctx, envelope.EventID) {
		sharednats.AckMessage(c.logger, msg)
		return
	}

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
		if jsonErr := json.Unmarshal(envelope.Payload, &payload); jsonErr != nil {
			err = jsonErr
		} else {
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
		if jsonErr := json.Unmarshal(envelope.Payload, &payload); jsonErr != nil {
			err = jsonErr
		} else {
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
		if jsonErr := json.Unmarshal(envelope.Payload, &payload); jsonErr != nil {
			err = jsonErr
		} else {
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
		c.logger.Error("processing leave event", zap.Error(err), zap.String("event_id", envelope.EventID))
		sharednats.NakMessage(c.logger, msg)
		return
	}

	// Mark as processed - must succeed before ACK
	if err := c.markProcessed(ctx, envelope.EventID); err != nil {
		c.logger.Error("failed to mark event as processed, NAKing for retry",
			zap.String("event_id", envelope.EventID),
			zap.Error(err))
		sharednats.NakMessage(c.logger, msg)
		return
	}

	sharednats.AckMessage(c.logger, msg)
}

func (c *LeaveConsumer) isProcessed(ctx context.Context, eventID string) bool {
	query := "SELECT 1 FROM notification.processed_events WHERE event_id = $1"
	var exists int
	err := c.pool.QueryRow(ctx, query, eventID).Scan(&exists)
	return err == nil
}

func (c *LeaveConsumer) markProcessed(ctx context.Context, eventID string) error {
	query := `
		INSERT INTO notification.processed_events (event_id, processed_at)
		VALUES ($1, NOW())
		ON CONFLICT DO NOTHING
	`
	if _, err := c.pool.Exec(ctx, query, eventID); err != nil {
		return fmt.Errorf("insert processed_events: %w", err)
	}
	return nil
}
