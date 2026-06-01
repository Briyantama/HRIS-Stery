package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hris-stery/hris-stery/services/employee-service/internal/domain"
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
	case *domain.EmployeeCreatedEvent:
		return map[string]interface{}{
			"employee_id":   e.EmployeeID.String(),
			"email":         e.Email,
			"full_name":     e.FullName,
			"department_id": e.DepartmentID.String(),
			"position_id":   e.PositionID.String(),
		}
	case *domain.EmployeeTerminatedEvent:
		return map[string]interface{}{
			"employee_id":      e.EmployeeID.String(),
			"email":            e.Email,
			"full_name":        e.FullName,
			"termination_date": e.TerminationDate.Format(time.RFC3339),
		}
	case *domain.EmployeeUpdatedEvent:
		return map[string]interface{}{
			"employee_id": e.EmployeeID.String(),
			"email":       e.Email,
			"full_name":   e.FullName,
			"status":      string(e.Status),
		}
	default:
		return map[string]interface{}{}
	}
}

// extractContextFromEvent pulls tenant and actor IDs from the event.
func (p *NATSPublisher) extractContextFromEvent(event domain.DomainEvent) (string, string) {
	switch e := event.(type) {
	case *domain.EmployeeCreatedEvent:
		return e.TenantID.String(), e.ActorID
	case *domain.EmployeeTerminatedEvent:
		return e.TenantID.String(), e.ActorID
	case *domain.EmployeeUpdatedEvent:
		return e.TenantID.String(), e.ActorID
	default:
		return "unknown", "unknown"
	}
}

// eventTypeToSubject converts an event type to a NATS subject.
// Format: hris.{domain}.{entity}.{verb} per ADR-0003.
func eventTypeToSubject(eventType string) string {
	// Event types are already in the correct format (e.g., "hris.workforce.employee.created")
	return eventType
}
