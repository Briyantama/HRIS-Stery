package nats

import (
	"context"
	"encoding/json"

	sharednats "github.com/hris-stery/hris-stery/services/_shared/nats"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/audit-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// OperationsConsumer consumes hris.operations.* events and records audit entries
type OperationsConsumer struct {
	js      nats.JetStreamContext
	pool    *pgxpool.Pool
	handler commands.RecordAuditHandler
	logger  *zap.Logger
}

// NewOperationsConsumer creates a new operations event consumer
func NewOperationsConsumer(
	js nats.JetStreamContext,
	pool *pgxpool.Pool,
	handler commands.RecordAuditHandler,
	logger *zap.Logger,
) *OperationsConsumer {
	return &OperationsConsumer{
		js:      js,
		pool:    pool,
		handler: handler,
		logger:  logger,
	}
}

// Subscribe subscribes to operations events
func (c *OperationsConsumer) Subscribe(ctx context.Context) error {
	_, err := c.js.Subscribe("hris.operations.>", func(msg *nats.Msg) {
		c.handleMessage(context.Background(), msg)
	}, nats.Durable("audit-operations-consumer"))
	return err
}

func (c *OperationsConsumer) handleMessage(ctx context.Context, msg *nats.Msg) {
	var envelope EventEnvelope
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		c.logger.Error("failed to unmarshal operations event", zap.Error(err), zap.String("event_id", envelope.EventID))
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
	case "hris.operations.attendance.marked":
		action = domain.AuditAction("ATTENDANCE_MARKED")
		resourceType = domain.ResourceType("ATTENDANCE")
		description = "Attendance marked"

	case "hris.operations.attendance.corrected":
		action = domain.AuditAction("ATTENDANCE_CORRECTED")
		resourceType = domain.ResourceType("ATTENDANCE")
		description = "Attendance corrected"

	case "hris.operations.leave.approval_changed":
		action = domain.AuditAction("LEAVE_APPROVAL_CHANGED")
		resourceType = domain.ResourceLeave
		description = "Leave approval changed"

	default:
		// Unknown event type, skip
		c.logger.Warn("unknown operations event type", zap.String("event_type", envelope.EventType))
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
		c.logger.Error("failed to record audit entry from operations event",
			zap.Error(err),
			zap.String("event_id", envelope.EventID),
			zap.String("event_type", envelope.EventType),
		)
		sharednats.NakMessage(c.logger, msg)
		return
	}

	// Mark as processed
	c.markProcessed(ctx, envelope.EventID, msg.Subject)
	c.logger.Debug("recorded audit entry from operations event",
		zap.String("event_id", envelope.EventID),
		zap.String("audit_entry_id", result.EntryID),
	)
	sharednats.AckMessage(c.logger, msg)
}

func (c *OperationsConsumer) isProcessed(ctx context.Context, eventID string) bool {
	query := "SELECT 1 FROM audit.processed_events WHERE event_id = $1"
	var exists int
	err := c.pool.QueryRow(ctx, query, eventID).Scan(&exists)
	return err == nil
}

func (c *OperationsConsumer) markProcessed(ctx context.Context, eventID string, subject string) {
	query := `
		INSERT INTO audit.processed_events (event_id, subject)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, _ = c.pool.Exec(ctx, query, eventID, subject)
}
