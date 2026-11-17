package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Briyantama/HRIS-Stery/services/notification/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/notification/application/consumers"
	"github.com/Briyantama/HRIS-Stery/services/notification/domain"
	"github.com/Briyantama/HRIS-Stery/services/notification/infrastructure/postgres"
	"github.com/Briyantama/stery-kit/go/events"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConsumerGroup is the stream consumer group of this service.
const ConsumerGroup = "notification-service"

// eventPayload is the union of the payload fields the handled events carry.
type eventPayload struct {
	UserID     string `json:"user_id"`
	EmployeeID string `json:"employee_id"`
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	LeaveType  string `json:"leave_type"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	ApproverID string `json:"approver_id"`
	Reason     string `json:"reason"`
}

// Consumers returns the stream handlers keyed by event type (the old auth, employee and leave consumers).
func Consumers(d Deps) map[string]events.Handler {
	configs := postgres.NewChannelConfigRepository(d.Pool)
	send := commands.NewSendNotificationHandler(postgres.NewNotificationRepository(d.Pool))
	auth := consumers.NewAuthEventConsumer(send, configs)
	employee := consumers.NewEmployeeEventConsumer(send, configs)
	leave := consumers.NewLeaveEventConsumer(send, configs)
	idem := func(fn func(context.Context, events.Envelope, eventPayload, domain.TenantID) error) events.Handler {
		return idempotent(d.Pool, fn)
	}

	return map[string]events.Handler{
		"hris.identity.user.registered": idem(func(ctx context.Context, e events.Envelope, p eventPayload, t domain.TenantID) error {
			return auth.ConsumeUserRegistered(ctx, &domain.UserRegisteredEvent{
				EventID: e.EventID, TenantID: t, ActorID: e.ActorID, OccurredAt: e.OccurredAt,
				UserID: p.UserID, Email: p.Email, FirstName: p.FirstName, LastName: p.LastName,
			})
		}),
		"hris.workforce.employee.created": idem(func(ctx context.Context, e events.Envelope, p eventPayload, t domain.TenantID) error {
			return employee.ConsumeEmployeeCreated(ctx, &domain.EmployeeCreatedEvent{
				EventID: e.EventID, TenantID: t, ActorID: e.ActorID, OccurredAt: e.OccurredAt,
				EmployeeID: p.EmployeeID, Email: p.Email, FirstName: p.FirstName, LastName: p.LastName,
			})
		}),
		"hris.workforce.employee.terminated": idem(func(ctx context.Context, e events.Envelope, p eventPayload, t domain.TenantID) error {
			return employee.ConsumeEmployeeTerminated(ctx, &domain.EmployeeTerminatedEvent{
				EventID: e.EventID, TenantID: t, ActorID: e.ActorID, OccurredAt: e.OccurredAt,
				EmployeeID: p.EmployeeID, Email: p.Email, FirstName: p.FirstName, LastName: p.LastName,
			})
		}),
		"hris.operations.leave.requested": idem(func(ctx context.Context, e events.Envelope, p eventPayload, t domain.TenantID) error {
			return leave.ConsumeLeaveRequested(ctx, &domain.LeaveRequestedEvent{
				EventID: e.EventID, TenantID: t, ActorID: e.ActorID, OccurredAt: e.OccurredAt,
				EmployeeID: p.EmployeeID, LeaveType: p.LeaveType, StartDate: p.StartDate, EndDate: p.EndDate, ApproverId: p.ApproverID,
			})
		}),
		"hris.operations.leave.approved": idem(func(ctx context.Context, e events.Envelope, p eventPayload, t domain.TenantID) error {
			return leave.ConsumeLeaveApproved(ctx, &domain.LeaveApprovedEvent{
				EventID: e.EventID, TenantID: t, ActorID: e.ActorID, OccurredAt: e.OccurredAt,
				EmployeeID: p.EmployeeID, LeaveType: p.LeaveType, StartDate: p.StartDate, EndDate: p.EndDate, ApproverId: p.ApproverID,
			})
		}),
		"hris.operations.leave.rejected": idem(func(ctx context.Context, e events.Envelope, p eventPayload, t domain.TenantID) error {
			return leave.ConsumeLeaveRejected(ctx, &domain.LeaveRejectedEvent{
				EventID: e.EventID, TenantID: t, ActorID: e.ActorID, OccurredAt: e.OccurredAt,
				EmployeeID: p.EmployeeID, LeaveType: p.LeaveType, StartDate: p.StartDate, EndDate: p.EndDate, Reason: p.Reason,
			})
		}),
	}
}

// idempotent skips events already recorded in notification.processed_events, decodes the payload,
// runs fn and records the event. Any returned error leaves the stream entry pending. An event whose
// tenant id is not a UUID can never succeed, so it is logged and acknowledged.
func idempotent(pool *pgxpool.Pool, fn func(context.Context, events.Envelope, eventPayload, domain.TenantID) error) events.Handler {
	return func(ctx context.Context, env events.Envelope) error {
		var done bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM notification.processed_events WHERE event_id = $1)`, env.EventID).Scan(&done); err != nil {
			return fmt.Errorf("check processed_events: %w", err)
		}
		if done {
			return nil
		}
		tenant, err := domain.NewTenantID(env.TenantID)
		if err != nil {
			slog.ErrorContext(ctx, "dropping event with invalid tenant id", "event_id", env.EventID, "event_type", env.EventType)
			return nil
		}
		var p eventPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return fmt.Errorf("decode %s payload: %w", env.EventType, err)
		}
		if err := fn(ctx, env, p, tenant); err != nil {
			return fmt.Errorf("handle %s: %w", env.EventType, err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO notification.processed_events (event_id, processed_at) VALUES ($1, NOW()) ON CONFLICT DO NOTHING`, env.EventID); err != nil {
			return fmt.Errorf("insert processed_events: %w", err)
		}
		return nil
	}
}
