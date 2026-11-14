// Package stream publishes attendance domain events to the HRIS event stream.
package stream

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Briyantama/stery-kit/go/events"

	"github.com/Briyantama/HRIS-Stery/services/attendance/domain"
)

// Sender is the part of *events.Publisher this adapter needs.
type Sender interface {
	Publish(ctx context.Context, env events.Envelope) error
}

// EventPublisher implements domain.EventPublisher on top of the stream publisher.
type EventPublisher struct {
	sender Sender
}

// NewEventPublisher creates an event publisher; pass the *events.Publisher of the HRIS stream.
func NewEventPublisher(sender Sender) *EventPublisher {
	return &EventPublisher{sender: sender}
}

// Publish wraps the domain event in the standard envelope. The event type string is the stream
// "type" (formerly the NATS subject), e.g. hris.attendance.record.checked_in.
func (p *EventPublisher) Publish(ctx context.Context, event domain.DomainEvent, tenantID domain.TenantID, actorID string) error {
	payload, err := json.Marshal(payloadOf(event))
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}
	return p.sender.Publish(ctx, events.Envelope{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		SchemaVersion: 1,
		TenantID:      tenantID.String(),
		ActorID:       actorID,
		OccurredAt:    event.OccurredAt(),
		Payload:       payload,
	})
}

// payloadOf exposes the event's fields; the event structs keep them unexported.
func payloadOf(event domain.DomainEvent) map[string]any {
	switch e := event.(type) {
	case *domain.AttendanceCheckedInEvent:
		lat, lon := e.Location()
		return map[string]any{
			"attendance_id": e.AttendanceID(), "employee_id": e.EmployeeID(),
			"check_in_at": e.CheckInAt(), "latitude": lat, "longitude": lon,
		}
	case *domain.AttendanceCheckedOutEvent:
		lat, lon := e.Location()
		return map[string]any{
			"attendance_id": e.AttendanceID(), "employee_id": e.EmployeeID(),
			"check_out_at": e.CheckOutAt(), "latitude": lat, "longitude": lon,
			"work_duration_mins": e.WorkDurationMins(),
		}
	case *domain.AttendanceOverriddenEvent:
		return map[string]any{
			"attendance_id": e.AttendanceID(), "employee_id": e.EmployeeID(), "status": e.Status(),
			"check_in_at": e.CheckInAt(), "check_out_at": e.CheckOutAt(), "reason": e.Reason(),
		}
	default:
		return map[string]any{}
	}
}
