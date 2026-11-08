package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Briyantama/stery-kit/go/events"

	"github.com/Briyantama/HRIS-Stery/services/auth/application"
	"github.com/Briyantama/HRIS-Stery/services/auth/domain"
)

// streamPublisher implements application.EventPublisher on top of the shared event stream.
type streamPublisher struct {
	pub *events.Publisher
}

// NewEventPublisher wraps the shared stream publisher. The envelope fields and event type strings
// are the ones the NATS publisher used (ADR-0003, now ADR-0006).
func NewEventPublisher(pub *events.Publisher) application.EventPublisher {
	return &streamPublisher{pub: pub}
}

// PublishSync appends the event to the stream.
func (p *streamPublisher) PublishSync(ctx context.Context, event domain.DomainEvent) error {
	payload, err := json.Marshal(eventPayload(event))
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	tenantID, actorID := eventContext(event)
	if err := p.pub.Publish(ctx, events.Envelope{
		EventID:       event.EventID(),
		EventType:     event.EventType(),
		SchemaVersion: 1,
		TenantID:      tenantID,
		ActorID:       actorID,
		OccurredAt:    event.OccurredAt().UTC(),
		Payload:       payload,
	}); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}

func eventPayload(event domain.DomainEvent) map[string]any {
	switch e := event.(type) {
	case domain.UserRegisteredEvent:
		return map[string]any{"user_id": e.UserID.String(), "email": e.Email, "full_name": e.FullName}
	case domain.UserLoggedInEvent:
		return map[string]any{"user_id": e.UserID.String(), "email": e.Email, "ip_addr": e.IPAddr}
	case domain.UserLoginFailedEvent:
		return map[string]any{"email": e.Email, "tenant_slug": e.TenantSlug, "reason": e.Reason, "ip_addr": e.IPAddr}
	case domain.PasswordChangedEvent:
		return map[string]any{"user_id": e.UserID.String()}
	case domain.TenantCreatedEvent:
		return map[string]any{"tenant_id": e.TenantID.String(), "tenant_slug": e.TenantSlug, "company_name": e.CompanyName}
	default:
		return map[string]any{}
	}
}

// eventContext returns the envelope tenant and actor ids.
func eventContext(event domain.DomainEvent) (tenantID, actorID string) {
	switch e := event.(type) {
	case domain.UserRegisteredEvent:
		return e.TenantID.String(), "system"
	case domain.UserLoggedInEvent:
		return e.TenantID.String(), e.UserID.String()
	case domain.UserLoginFailedEvent:
		return "unknown", "system" // tenant not yet known on failed login
	case domain.PasswordChangedEvent:
		return e.TenantID.String(), e.UserID.String()
	case domain.TenantCreatedEvent:
		return e.TenantID.String(), "system"
	default:
		return "unknown", "unknown"
	}
}
