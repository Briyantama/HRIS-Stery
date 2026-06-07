package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sharednats "github.com/hris-stery/hris-stery/services/_shared/nats"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// NotificationConsumer consumes hris.notification.* events and records audit entries
type NotificationConsumer struct {
	js      nats.JetStreamContext
	pool    *pgxpool.Pool
	handler commands.RecordAuditHandler
	logger  *zap.Logger
}

// NewNotificationConsumer creates a new notification event consumer
func NewNotificationConsumer(
	js nats.JetStreamContext,
	pool *pgxpool.Pool,
	handler commands.RecordAuditHandler,
	logger *zap.Logger,
) *NotificationConsumer {
	return &NotificationConsumer{
		js:      js,
		pool:    pool,
		handler: handler,
		logger:  logger,
	}
}

// Subscribe subscribes to notification events
// Propagates shutdown context so graceful shutdown can drain handlers
func (c *NotificationConsumer) Subscribe(ctx context.Context) error {
	_, err := c.js.Subscribe("hris.notification.>", func(msg *nats.Msg) {
		msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		c.handleMessage(msgCtx, msg)
	},
		nats.Durable("audit-notification-consumer"),
		nats.MaxAckPending(1000),
		nats.AckWait(30*time.Second))
	return err
}

func (c *NotificationConsumer) handleMessage(ctx context.Context, msg *nats.Msg) {
	var envelope EventEnvelope
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		c.logger.Error("failed to unmarshal notification event", zap.Error(err), zap.String("event_id", envelope.EventID))
		sharednats.NakMessage(c.logger, msg)
		return
	}

	// Check idempotency
	if c.isProcessed(ctx, envelope.EventID) {
		sharednats.AckMessage(c.logger, msg)
		return
	}

	var action domain.AuditAction
	var resourceType domain.ResourceType
	var description string

	// Map event type to audit action and resource type
	switch envelope.EventType {
	case "hris.notification.sent":
		action = domain.AuditAction("NOTIFICATION_SENT")
		resourceType = domain.ResourceNotification
		description = "Notification sent"

	case "hris.notification.failed":
		action = domain.AuditAction("NOTIFICATION_FAILED")
		resourceType = domain.ResourceNotification
		description = "Notification failed"

	case "hris.notification.read":
		action = domain.AuditAction("NOTIFICATION_READ")
		resourceType = domain.ResourceNotification
		description = "Notification read"

	default:
		// Unknown event type, skip
		c.logger.Warn("unknown notification event type", zap.String("event_type", envelope.EventType))
		sharednats.AckMessage(c.logger, msg)
		return
	}

	// Record audit entry
	cmd := commands.RecordAuditCommand{
		TenantID:     envelope.TenantID,
		ActorID:      envelope.ActorID,
		Action:       action,
		ResourceType: resourceType,
		Description:  description,
		Success:      true,
		Changes:      make(map[string]string),
	}

	result, err := c.handler.Handle(ctx, cmd)
	if err != nil {
		c.logger.Error("failed to record audit entry from notification event",
			zap.Error(err),
			zap.String("event_id", envelope.EventID),
			zap.String("event_type", envelope.EventType),
		)
		sharednats.NakMessage(c.logger, msg)
		return
	}

	// Mark as processed - must succeed before ACK
	if err := c.markProcessed(ctx, envelope.EventID, msg.Subject); err != nil {
		c.logger.Error("failed to mark event as processed, NAKing for retry",
			zap.String("event_id", envelope.EventID),
			zap.Error(err))
		sharednats.NakMessage(c.logger, msg)
		return
	}

	c.logger.Debug("recorded audit entry from notification event",
		zap.String("event_id", envelope.EventID),
		zap.String("audit_entry_id", result.EntryID),
	)
	sharednats.AckMessage(c.logger, msg)
}

func (c *NotificationConsumer) isProcessed(ctx context.Context, eventID string) bool {
	query := "SELECT 1 FROM audit.processed_events WHERE event_id = $1"
	var exists int
	err := c.pool.QueryRow(ctx, query, eventID).Scan(&exists)
	return err == nil
}

func (c *NotificationConsumer) markProcessed(ctx context.Context, eventID string, subject string) error {
	query := `
		INSERT INTO audit.processed_events (event_id, subject)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	if _, err := c.pool.Exec(ctx, query, eventID, subject); err != nil {
		return fmt.Errorf("insert processed_events: %w", err)
	}
	return nil
}
