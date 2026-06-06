package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sharednats "github.com/hris-stery/hris-stery/services/_shared/nats"
	"github.com/hris-stery/hris-stery/services/leave-service/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

// EmployeeCreatedConsumer processes employee creation events and initializes leave balances.
type EmployeeCreatedConsumer struct {
	js                   jetstream.JetStream
	pool                 *pgxpool.Pool
	leaveBalanceRepo     domain.LeaveBalanceRepository
	leaveTypeRepo        domain.LeaveTypeRepository
	employeeCreatedEvent chan struct{}
	logger               *zap.Logger
}

// EmployeeCreatedPayload represents the payload of hris.workforce.employee.created event.
type EmployeeCreatedPayload struct {
	EmployeeID string `json:"employee_id"`
	TenantID   string `json:"tenant_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
}

// NewEmployeeCreatedConsumer creates a new employee creation event consumer.
func NewEmployeeCreatedConsumer(
	js jetstream.JetStream,
	pool *pgxpool.Pool,
	leaveBalanceRepo domain.LeaveBalanceRepository,
	leaveTypeRepo domain.LeaveTypeRepository,
	logger *zap.Logger,
) *EmployeeCreatedConsumer {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &EmployeeCreatedConsumer{
		js:                   js,
		pool:                 pool,
		leaveBalanceRepo:     leaveBalanceRepo,
		leaveTypeRepo:        leaveTypeRepo,
		employeeCreatedEvent: make(chan struct{}, 100),
		logger:               logger,
	}
}

// Subscribe subscribes to the employee creation event stream.
func (c *EmployeeCreatedConsumer) Subscribe(ctx context.Context) error {
	subject := "hris.workforce.employee.created"
	consumerName := "leave-service-employee-created"

	// Create or retrieve the consumer
	_, err := c.js.CreateOrUpdateConsumer(ctx, "HRIS_EVENTS", jetstream.ConsumerConfig{
		Name:              consumerName,
		Durable:           consumerName,
		AckPolicy:         jetstream.AckExplicitPolicy,
		FilterSubject:     subject,
		MaxAckPending:     1000,
		InactiveThreshold: 5 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("creating consumer: %w", err)
	}

	// Start consuming messages
	go func() {
		for {
			consumer, err := c.js.Consumer(ctx, "HRIS_EVENTS", consumerName)
			if err != nil {
				c.logger.Error("getting jetstream consumer", zap.Error(err), zap.String("consumer", consumerName))
				time.Sleep(5 * time.Second)
				continue
			}

			msg, err := consumer.Next()
			if err != nil {
				c.logger.Warn("waiting for next jetstream message", zap.Error(err), zap.String("consumer", consumerName))
				time.Sleep(1 * time.Second)
				continue
			}

			if err := c.processMessage(ctx, msg.Data()); err != nil {
				c.logger.Error("processing employee created event", zap.Error(err))
				sharednats.NakJetStream(c.logger, msg)
				continue
			}
			sharednats.AckJetStream(c.logger, msg)
		}
	}()

	return nil
}

// processMessage processes a single employee creation event.
func (c *EmployeeCreatedConsumer) processMessage(ctx context.Context, data []byte) error {
	var envelope EventEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("unmarshaling envelope: %w", err)
	}

	// Check idempotency
	eventID := envelope.EventID
	alreadyProcessed, err := c.isEventProcessed(ctx, eventID)
	if err != nil {
		return fmt.Errorf("checking event idempotency: %w", err)
	}
	if alreadyProcessed {
		return nil
	}

	// Unmarshal payload
	var payload EmployeeCreatedPayload
	payloadBytes, _ := json.Marshal(envelope.Payload)
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("unmarshaling payload: %w", err)
	}

	// Initialize leave balances for all leave types
	tenantID := domain.MustNewTenantID(payload.TenantID)
	employeeID := domain.MustNewEmployeeID(payload.EmployeeID)

	leaveTypes, err := c.leaveTypeRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("listing leave types: %w", err)
	}

	currentYear := time.Now().Year()
	for _, lt := range leaveTypes {
		balance, err := domain.NewLeaveBalance(tenantID, employeeID, lt.ID(), currentYear, lt.MaxDaysPerYear())
		if err != nil {
			return fmt.Errorf("creating leave balance: %w", err)
		}

		if err := c.leaveBalanceRepo.Create(ctx, balance); err != nil {
			return fmt.Errorf("persisting leave balance: %w", err)
		}
	}

	// Mark event as processed
	if err := c.markEventProcessed(ctx, eventID); err != nil {
		return fmt.Errorf("marking event as processed: %w", err)
	}

	return nil
}

// isEventProcessed checks if an event has already been processed.
func (c *EmployeeCreatedConsumer) isEventProcessed(ctx context.Context, eventID string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM leave.processed_events WHERE event_id = $1)
	`
	var exists bool
	err := c.pool.QueryRow(ctx, query, eventID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("querying processed_events: %w", err)
	}
	return exists, nil
}

// markEventProcessed marks an event as processed.
func (c *EmployeeCreatedConsumer) markEventProcessed(ctx context.Context, eventID string) error {
	query := `
		INSERT INTO leave.processed_events (event_id, processed_at)
		VALUES ($1, NOW())
		ON CONFLICT DO NOTHING
	`
	_, err := c.pool.Exec(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("inserting processed event: %w", err)
	}
	return nil
}

// Close stops the consumer gracefully.
func (c *EmployeeCreatedConsumer) Close() error {
	close(c.employeeCreatedEvent)
	return nil
}
