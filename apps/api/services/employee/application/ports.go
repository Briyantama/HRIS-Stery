package application

import (
	"context"

	"github.com/Briyantama/HRIS-Stery/services/employee/domain"
)

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	PublishSync(ctx context.Context, event domain.DomainEvent) error
}
