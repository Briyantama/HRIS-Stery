package stream

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/leave/domain"
	"github.com/Briyantama/stery-kit/go/events"
)

// EventPublisher publishes domain events to the HRIS event stream. The event type string
// (for example hris.operations.leave.requested) is the stream entry type.
type EventPublisher struct {
	pub *events.Publisher
}

// NewEventPublisher creates an EventPublisher on top of the shared stream publisher.
func NewEventPublisher(pub *events.Publisher) *EventPublisher {
	return &EventPublisher{pub: pub}
}

// Publish wraps the event in the standard HRIS envelope and appends it to the stream.
func (p *EventPublisher) Publish(ctx context.Context, event domain.DomainEvent, tenantID domain.TenantID, actorID string) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshaling event: %w", err)
	}
	return p.pub.Publish(ctx, events.Envelope{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		SchemaVersion: 1,
		TenantID:      tenantID.String(),
		ActorID:       actorID,
		OccurredAt:    event.OccurredAt(),
		Payload:       payload,
	})
}
