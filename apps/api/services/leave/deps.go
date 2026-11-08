package leave

import (
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/leave/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/leave/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/leave/infrastructure/cache"
	"github.com/Briyantama/HRIS-Stery/services/leave/infrastructure/consumers"
	"github.com/Briyantama/HRIS-Stery/services/leave/infrastructure/postgres"
	"github.com/Briyantama/HRIS-Stery/services/leave/infrastructure/stream"
	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/Briyantama/stery-kit/go/events"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConsumerGroup is the event stream consumer group of this service.
const ConsumerGroup = "leave-service"

// Deps is everything the leave service needs from the process.
type Deps struct {
	Pool     *pgxpool.Pool
	Events   *events.Publisher
	Verifier authn.Verifier
}

// Build wires repositories, command and query handlers and the HTTP handler.
func Build(d Deps) (*HTTPHandler, error) {
	if d.Pool == nil || d.Events == nil || d.Verifier == nil {
		return nil, fmt.Errorf("leave: Pool, Events and Verifier are required")
	}
	requests := postgres.NewLeaveRequestRepository(d.Pool)
	balances := postgres.NewLeaveBalanceRepository(d.Pool)
	types := cache.NewLeaveTypeCacheRepository(postgres.NewLeaveTypeRepository(d.Pool))
	pub := stream.NewEventPublisher(d.Events)

	return NewHTTPHandler(
		commands.NewCreateLeaveRequestHandler(requests, balances, pub, postgres.NewEmployeeValidator(d.Pool)),
		commands.NewApproveLeaveRequestHandler(requests, balances, pub),
		commands.NewRejectLeaveRequestHandler(requests, balances, pub),
		commands.NewCancelLeaveRequestHandler(requests, balances, pub),
		queries.NewGetLeaveRequestHandler(requests),
		queries.NewListLeaveRequestsHandler(requests),
		queries.NewGetLeaveBalanceHandler(balances),
		queries.NewListLeaveTypesHandler(types),
		d.Verifier,
	), nil
}

// Consumers returns the stream handlers of this service, keyed by event type.
func Consumers(d Deps) map[string]events.Handler {
	balances := postgres.NewLeaveBalanceRepository(d.Pool)
	types := cache.NewLeaveTypeCacheRepository(postgres.NewLeaveTypeRepository(d.Pool))
	return map[string]events.Handler{
		consumers.EmployeeCreatedType: consumers.NewEmployeeCreated(d.Pool, balances, types).Handle,
	}
}
