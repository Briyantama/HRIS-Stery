package audit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Briyantama/stery-kit/go/events"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Briyantama/HRIS-Stery/services/audit/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/audit/domain"
)

// ConsumerGroup is the Dragonfly stream consumer group of this service.
const ConsumerGroup = "audit-service"

type auditMapping struct {
	action       domain.AuditAction
	resourceType domain.ResourceType
	description  string
}

// mappings holds exactly the event types the four old consumers (identity, workforce, operations,
// notification) handled. Some never match a publisher; see the port report:
// hris.identity.user.deactivated, hris.operations.attendance.marked|corrected, hris.notification.*.
var mappings = map[string]auditMapping{
	"hris.identity.user.registered":        {domain.ActionLogin, domain.ResourceUser, "User registered"},
	"hris.identity.user.deactivated":       {domain.ActionLogout, domain.ResourceUser, "User deactivated"},
	"hris.workforce.employee.created":      {domain.ActionEmployeeCreate, domain.ResourceEmployee, "Employee created"},
	"hris.workforce.employee.terminated":   {"EMPLOYEE_TERMINATED", domain.ResourceEmployee, "Employee terminated"},
	"hris.workforce.employee.updated":      {"EMPLOYEE_UPDATED", domain.ResourceEmployee, "Employee updated"},
	"hris.operations.attendance.marked":    {"ATTENDANCE_MARKED", "ATTENDANCE", "Attendance marked"},
	"hris.operations.attendance.corrected": {"ATTENDANCE_CORRECTED", "ATTENDANCE", "Attendance corrected"},
	"hris.operations.leave.requested":      {domain.ActionLeaveRequested, domain.ResourceLeave, "Leave request created"},
	"hris.operations.leave.approved":       {domain.ActionLeaveApproved, domain.ResourceLeave, "Leave request approved"},
	"hris.operations.leave.rejected":       {"LEAVE_REJECTED", domain.ResourceLeave, "Leave request rejected"},
	"hris.operations.leave.cancelled":      {"LEAVE_CANCELLED", domain.ResourceLeave, "Leave request cancelled"},
	"hris.notification.sent":               {"NOTIFICATION_SENT", domain.ResourceNotification, "Notification sent"},
	"hris.notification.failed":             {"NOTIFICATION_FAILED", domain.ResourceNotification, "Notification failed"},
	"hris.notification.read":               {"NOTIFICATION_READ", domain.ResourceNotification, "Notification read"},
}

// Consumers returns one handler per handled event type. A handler returns an error (entry stays
// pending) when recording or the idempotency mark fails.
func Consumers(d Deps) map[string]events.Handler {
	rec := commands.NewRecordAuditHandler(newRepo(d.Pool))
	out := make(map[string]events.Handler, len(mappings))
	for typ, m := range mappings {
		out[typ] = func(ctx context.Context, env events.Envelope) error {
			return handleEvent(ctx, d.Pool, rec, m, env)
		}
	}
	return out
}

func handleEvent(ctx context.Context, pool *pgxpool.Pool, rec commands.RecordAuditHandler, m auditMapping, env events.Envelope) error {
	var one int
	err := pool.QueryRow(ctx, `SELECT 1 FROM audit.processed_events WHERE event_id = $1`, env.EventID).Scan(&one)
	switch {
	case err == nil:
		return nil // already processed
	case !errors.Is(err, pgx.ErrNoRows):
		return fmt.Errorf("check processed_events: %w", err)
	}

	actor := env.ActorID
	if actor == "" {
		actor = "system" // events without an actor must not poison the stream
	}
	res, err := rec.Handle(ctx, commands.RecordAuditCommand{
		TenantID:     env.TenantID,
		ActorID:      actor,
		Action:       m.action,
		ResourceType: m.resourceType,
		Description:  m.description,
		Success:      true,
		Changes:      map[string]string{},
	})
	if err != nil {
		return fmt.Errorf("record audit entry for %s: %w", env.EventType, err)
	}

	if _, err := pool.Exec(ctx,
		`INSERT INTO audit.processed_events (event_id, subject) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		env.EventID, env.EventType); err != nil {
		return fmt.Errorf("mark event processed: %w", err)
	}
	slog.DebugContext(ctx, "recorded audit entry from event",
		"event_id", env.EventID, "event_type", env.EventType, "tenant_id", env.TenantID, "audit_entry_id", res.EntryID)
	return nil
}
