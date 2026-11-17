package attendance

import (
	"fmt"

	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/Briyantama/stery-kit/go/events"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Briyantama/HRIS-Stery/services/attendance/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/attendance/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/attendance/infrastructure/postgres"
	"github.com/Briyantama/HRIS-Stery/services/attendance/infrastructure/stream"
)

// Deps is everything the attendance service needs from the process. It publishes events and
// consumes none, so there is no ConsumerGroup.
type Deps struct {
	Pool     *pgxpool.Pool
	Events   *events.Publisher
	Verifier authn.Verifier
}

// Build wires the repository, command and query handlers and the HTTP handler.
func Build(d Deps) (*HTTPHandler, error) {
	if d.Pool == nil || d.Events == nil || d.Verifier == nil {
		return nil, fmt.Errorf("attendance: Pool, Events and Verifier are required")
	}
	repo := postgres.NewAttendanceRepository(d.Pool)
	pub := stream.NewEventPublisher(d.Events)
	return NewHTTPHandler(
		commands.NewCheckInHandler(repo, pub),
		commands.NewCheckOutHandler(repo, pub),
		commands.NewOverrideAttendanceHandler(repo, pub),
		queries.NewListAttendanceHandler(repo),
		queries.NewGetDailySummaryHandler(repo),
		d.Verifier,
	), nil
}
