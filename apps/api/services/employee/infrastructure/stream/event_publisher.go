package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Briyantama/HRIS-Stery/services/employee/domain"
	"github.com/Briyantama/stery-kit/go/events"
)

// EventPublisher implements application.EventPublisher on the HRIS event stream. The event type
// (hris.workforce.employee.created|terminated) is the stream entry type.
type EventPublisher struct {
	pub *events.Publisher
}

// NewEventPublisher creates an EventPublisher on top of the shared stream publisher.
func NewEventPublisher(pub *events.Publisher) *EventPublisher {
	return &EventPublisher{pub: pub}
}

// PublishSync wraps the event in the standard envelope and appends it to the stream.
func (p *EventPublisher) PublishSync(ctx context.Context, event domain.DomainEvent) error {
	tenantID, actorID, payload, err := describe(event)
	if err != nil {
		return err
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}
	return p.pub.Publish(ctx, events.Envelope{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		SchemaVersion: 1,
		TenantID:      tenantID,
		ActorID:       actorID,
		OccurredAt:    event.OccurredAt().UTC(),
		Payload:       data,
	})
}

// describe extracts the tenant, actor and payload fields of a known employee event.
func describe(event domain.DomainEvent) (tenantID, actorID string, payload map[string]any, err error) {
	switch e := event.(type) {
	case *domain.EmployeeCreatedEvent:
		return e.TenantID.String(), e.ActorID, map[string]any{
			"employee_id":   e.EmployeeID.String(),
			"email":         e.Email,
			"full_name":     e.FullName,
			"department_id": e.DepartmentID.String(),
			"position_id":   e.PositionID.String(),
		}, nil
	case *domain.EmployeeTerminatedEvent:
		return e.TenantID.String(), e.ActorID, map[string]any{
			"employee_id":      e.EmployeeID.String(),
			"email":            e.Email,
			"full_name":        e.FullName,
			"termination_date": e.TerminationDate.Format(time.RFC3339),
		}, nil
	default:
		return "", "", nil, fmt.Errorf("unsupported event type %T", event)
	}
}
