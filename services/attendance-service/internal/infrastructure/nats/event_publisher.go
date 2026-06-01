package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hris-stery/hris-stery/services/attendance-service/internal/domain"
	"github.com/nats-io/nats.go/jetstream"
)

// EventPublisher publishes domain events to NATS JetStream.
type EventPublisher struct {
	js jetstream.JetStream
}

// NewEventPublisher creates a new NATS event publisher.
func NewEventPublisher(js jetstream.JetStream) *EventPublisher {
	return &EventPublisher{
		js: js,
	}
}

// EventEnvelope is the standard envelope for all HRIS events.
type EventEnvelope struct {
	EventID       string      `json:"event_id"`
	EventType     string      `json:"event_type"`
	SchemaVersion int         `json:"schema_version"`
	TenantID      string      `json:"tenant_id"`
	ActorID       string      `json:"actor_id"`
	OccurredAt    string      `json:"occurred_at"`
	Payload       interface{} `json:"payload"`
}

// Publish publishes a domain event with proper envelope structure.
func (p *EventPublisher) Publish(ctx context.Context, event domain.DomainEvent, tenantID domain.TenantID, actorID string) error {
	envelope := EventEnvelope{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		SchemaVersion: 1,
		TenantID:      tenantID.String(),
		ActorID:       actorID,
		OccurredAt:    event.OccurredAt().Format(time.RFC3339),
		Payload:       event,
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshaling event: %w", err)
	}

	subject := event.EventType() // Subject is the event type (e.g., hris.attendance.record.checked_in)
	_, err = p.js.Publish(ctx, subject, payload)
	if err != nil {
		return fmt.Errorf("publishing event to NATS: %w", err)
	}

	return nil
}

// EventConsumer handles consumption of events from NATS JetStream with idempotency.
// Implementation deferred: will be implemented when building leave-service consumers.
type EventConsumer struct {
	js jetstream.JetStream
}

// NewEventConsumer creates a new event consumer.
func NewEventConsumer(js jetstream.JetStream) *EventConsumer {
	return &EventConsumer{
		js: js,
	}
}
