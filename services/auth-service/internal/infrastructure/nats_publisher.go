package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hris-stery/hris-stery/services/auth-service/internal/domain"
	"github.com/nats-io/nats.go"
)

// NATSPublisher implements application.EventPublisher using NATS JetStream.
type NATSPublisher struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

// NewNATSPublisher creates a new NATS event publisher.
func NewNATSPublisher(conn *nats.Conn) (*NATSPublisher, error) {
	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("get jetstream context: %w", err)
	}
	return &NATSPublisher{conn: conn, js: js}, nil
}

// EventEnvelope wraps domain events in a standard envelope per ADR-0003.
type EventEnvelope struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	SchemaVersion int                    `json:"schema_version"`
	TenantID      string                 `json:"tenant_id"`
	ActorID       string                 `json:"actor_id"`
	OccurredAt    string                 `json:"occurred_at"`
	Payload       map[string]interface{} `json:"payload"`
}

// PublishAsync publishes an event without waiting for confirmation.
func (p *NATSPublisher) PublishAsync(ctx context.Context, event domain.DomainEvent) error {
	envelope := p.buildEnvelope(event)
	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	subject := eventTypeToSubject(envelope.EventType)
	_, err = p.js.PublishAsync(subject, data)
	if err != nil {
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}

// PublishSync publishes an event and waits for confirmation.
func (p *NATSPublisher) PublishSync(ctx context.Context, event domain.DomainEvent) error {
	envelope := p.buildEnvelope(event)
	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	subject := eventTypeToSubject(envelope.EventType)
	_, err = p.js.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}

// buildEnvelope creates an event envelope from a domain event.
func (p *NATSPublisher) buildEnvelope(event domain.DomainEvent) *EventEnvelope {
	payload := p.eventToPayload(event)
	tenantID, actorID := p.extractContextFromEvent(event)

	return &EventEnvelope{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		SchemaVersion: 1,
		TenantID:      tenantID,
		ActorID:       actorID,
		OccurredAt:    event.OccurredAt().UTC().Format(time.RFC3339),
		Payload:       payload,
	}
}

// eventToPayload extracts payload fields from the event.
func (p *NATSPublisher) eventToPayload(event domain.DomainEvent) map[string]interface{} {
	switch e := event.(type) {
	case domain.UserRegisteredEvent:
		return map[string]interface{}{
			"user_id":   e.UserID.String(),
			"email":     e.Email,
			"full_name": e.FullName,
		}
	case domain.UserLoggedInEvent:
		return map[string]interface{}{
			"user_id": e.UserID.String(),
			"email":   e.Email,
			"ip_addr": e.IPAddr,
		}
	case domain.UserLoginFailedEvent:
		return map[string]interface{}{
			"email":       e.Email,
			"tenant_slug": e.TenantSlug,
			"reason":      e.Reason,
			"ip_addr":     e.IPAddr,
		}
	case domain.PasswordChangedEvent:
		return map[string]interface{}{
			"user_id": e.UserID.String(),
		}
	case domain.TenantCreatedEvent:
		return map[string]interface{}{
			"tenant_id":    e.TenantID.String(),
			"tenant_slug":  e.TenantSlug,
			"company_name": e.CompanyName,
		}
	default:
		return map[string]interface{}{}
	}
}

// extractContextFromEvent pulls tenant and actor IDs from the event.
func (p *NATSPublisher) extractContextFromEvent(event domain.DomainEvent) (string, string) {
	switch e := event.(type) {
	case domain.UserRegisteredEvent:
		return e.TenantID.String(), "system"
	case domain.UserLoggedInEvent:
		return e.TenantID.String(), e.UserID.String()
	case domain.UserLoginFailedEvent:
		return "unknown", "system" // Tenant not yet known on failed login
	case domain.PasswordChangedEvent:
		return e.TenantID.String(), e.UserID.String()
	case domain.TenantCreatedEvent:
		return e.TenantID.String(), "system"
	default:
		return "unknown", "unknown"
	}
}

// eventTypeToSubject converts an event type to a NATS subject.
// Format: hris.{domain}.{entity}.{verb} per ADR-0003.
func eventTypeToSubject(eventType string) string {
	// Event types are already in the correct format (e.g., "hris.identity.user.registered")
	return eventType
}
