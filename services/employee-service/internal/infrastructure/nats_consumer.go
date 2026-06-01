package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/application/commands"
	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// UserRegisteredEventPayload is the auth-service event payload.
type UserRegisteredEventPayload struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

// EventEnvelopeReceived wraps the incoming NATS event.
type EventEnvelopeReceived struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	SchemaVersion int                    `json:"schema_version"`
	TenantID      string                 `json:"tenant_id"`
	ActorID       string                 `json:"actor_id"`
	OccurredAt    string                 `json:"occurred_at"`
	Payload       map[string]interface{} `json:"payload"`
}

// UserRegisteredConsumer subscribes to hris.identity.user.registered events.
type UserRegisteredConsumer struct {
	pool                  *pgxpool.Pool
	createEmployeeHandler *commands.CreateEmployeeHandler
	logger                *zap.Logger
	mu                    sync.Mutex
	sub                   *nats.Subscription
	ctx                   context.Context
	cancel                context.CancelFunc
}

// NewUserRegisteredConsumer creates a new consumer for user registration events.
func NewUserRegisteredConsumer(
	pool *pgxpool.Pool,
	createEmployeeHandler *commands.CreateEmployeeHandler,
	logger *zap.Logger,
) *UserRegisteredConsumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &UserRegisteredConsumer{
		pool:                  pool,
		createEmployeeHandler: createEmployeeHandler,
		logger:                logger,
		ctx:                   ctx,
		cancel:                cancel,
	}
}

// Subscribe starts listening for user registration events.
func (c *UserRegisteredConsumer) Subscribe(conn *nats.Conn) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	js, err := conn.JetStream()
	if err != nil {
		return fmt.Errorf("get jetstream context: %w", err)
	}

	sub, err := js.Subscribe("hris.identity.user.registered", c.handleMessage)
	if err != nil {
		return fmt.Errorf("subscribe to events: %w", err)
	}

	c.sub = sub
	c.logger.Info("subscribed to hris.identity.user.registered")
	return nil
}

// handleMessage processes a single event message.
func (c *UserRegisteredConsumer) handleMessage(msg *nats.Msg) {
	var envelope EventEnvelopeReceived
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		c.logger.Error("unmarshal event", zap.Error(err))
		_ = msg.Ack()
		return
	}

	// Check idempotency
	if c.isProcessed(c.ctx, envelope.EventID) {
		c.logger.Info("event already processed", zap.String("event_id", envelope.EventID))
		_ = msg.Ack()
		return
	}

	// Create employee shell record
	if err := c.handleUserRegistered(envelope); err != nil {
		c.logger.Error("handle user registered", zap.Error(err), zap.String("event_id", envelope.EventID))
		// Still ack to prevent reprocessing; log for audit
		_ = msg.Ack()
		return
	}

	// Mark as processed
	if err := c.markProcessed(c.ctx, envelope.EventID, envelope.EventType); err != nil {
		c.logger.Error("mark event processed", zap.Error(err), zap.String("event_id", envelope.EventID))
		return
	}

	_ = msg.Ack()
}

// handleUserRegistered creates an employee shell record for a newly registered user.
func (c *UserRegisteredConsumer) handleUserRegistered(envelope EventEnvelopeReceived) error {
	// Extract payload
	payload := envelope.Payload
	tenantID, _ := payload["tenant_id"].(string)
	email, _ := payload["email"].(string)
	fullName, _ := payload["full_name"].(string)

	if tenantID == "" || email == "" || fullName == "" {
		return fmt.Errorf("missing required event fields")
	}

	// Validate tenant ID
	_, err := domain.NewTenantID(tenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant_id: %w", err)
	}

	// Get default department and position from tenant (use system defaults for shell record)
	// For MVP, create employee with empty department/position refs
	// These will be filled in by HR during onboarding flow
	defaultDeptID := domain.GenerateDepartmentID()
	defaultPosID := domain.GeneratePositionID()

	// Create employee shell record using the handler
	cmd := commands.CreateEmployeeCommand{
		TenantID:     tenantID,
		Email:        email,
		FullName:     fullName,
		Phone:        "",
		DepartmentID: defaultDeptID.String(),
		PositionID:   defaultPosID.String(),
		ManagerID:    "",
		ContractType: string(domain.ContractPermanent),
		ActorID:      "system",
	}

	_, err = c.createEmployeeHandler.Handle(c.ctx, cmd)
	if err != nil {
		// Check if employee already exists (idempotent)
		if err.Error() == "email already exists for this tenant" {
			return nil
		}
		return fmt.Errorf("create employee shell: %w", err)
	}

	c.logger.Info("created employee shell", zap.String("email", email), zap.String("tenant_id", tenantID))
	return nil
}

// isProcessed checks if an event has been processed before.
func (c *UserRegisteredConsumer) isProcessed(ctx context.Context, eventID string) bool {
	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		c.logger.Error("acquire connection", zap.Error(err))
		return false
	}
	defer conn.Release()

	query := `SELECT 1 FROM employee.processed_events WHERE event_id = $1`
	var exists int
	err = conn.QueryRow(ctx, query, eventID).Scan(&exists)
	if err != nil {
		// If not found, that's fine (not processed yet)
		return false
	}
	return true
}

// markProcessed records that an event has been processed.
func (c *UserRegisteredConsumer) markProcessed(ctx context.Context, eventID, eventType string) error {
	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	query := `INSERT INTO employee.processed_events (event_id, event_type) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err = conn.Exec(ctx, query, eventID, eventType)
	if err != nil {
		return fmt.Errorf("insert processed event: %w", err)
	}
	return nil
}

// Close stops the consumer.
func (c *UserRegisteredConsumer) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sub != nil {
		_ = c.sub.Unsubscribe()
	}
	c.cancel()
}
