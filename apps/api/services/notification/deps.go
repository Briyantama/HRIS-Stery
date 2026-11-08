package notification

import (
	"errors"

	"github.com/Briyantama/HRIS-Stery/services/notification/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/notification/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/notification/infrastructure/postgres"
	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Deps is everything the notification service needs from the process.
type Deps struct {
	Pool     *pgxpool.Pool
	Verifier authn.Verifier
}

// Build wires repositories and handlers into the HTTP handler.
func Build(d Deps) (*HTTPHandler, error) {
	if d.Pool == nil {
		return nil, errors.New("notification: Deps.Pool is required")
	}
	if d.Verifier == nil {
		return nil, errors.New("notification: Deps.Verifier is required")
	}
	notifications := postgres.NewNotificationRepository(d.Pool)
	configs := postgres.NewChannelConfigRepository(d.Pool)
	return NewHTTPHandler(
		commands.NewSendNotificationHandler(notifications),
		commands.NewMarkNotificationReadHandler(notifications),
		commands.NewUpdateChannelConfigHandler(configs),
		queries.NewListNotificationsHandler(notifications),
		queries.NewGetDeliveryStatusHandler(notifications),
		d.Verifier,
	), nil
}
