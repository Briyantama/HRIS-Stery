package employee

import (
	"fmt"

	"github.com/Briyantama/HRIS-Stery/services/employee/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/employee/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/employee/infrastructure/cache"
	"github.com/Briyantama/HRIS-Stery/services/employee/infrastructure/consumers"
	"github.com/Briyantama/HRIS-Stery/services/employee/infrastructure/postgres"
	"github.com/Briyantama/HRIS-Stery/services/employee/infrastructure/stream"
	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/Briyantama/stery-kit/go/events"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConsumerGroup is the event stream consumer group of this service.
const ConsumerGroup = "employee-service"

// Deps is everything the employee service needs from the process.
type Deps struct {
	Pool     *pgxpool.Pool
	Events   *events.Publisher
	Verifier authn.Verifier
}

type repos struct {
	employees   *postgres.EmployeeRepository
	departments *cache.DepartmentCacheRepository
	positions   *cache.PositionCacheRepository
}

func newRepos(pool *pgxpool.Pool) repos {
	return repos{
		employees:   postgres.NewEmployeeRepository(pool),
		departments: cache.NewDepartmentCacheRepository(postgres.NewDepartmentRepository(pool)),
		positions:   cache.NewPositionCacheRepository(postgres.NewPositionRepository(pool)),
	}
}

// Build wires repositories, command and query handlers and the HTTP handler.
func Build(d Deps) (*HTTPHandler, error) {
	if d.Pool == nil || d.Events == nil || d.Verifier == nil {
		return nil, fmt.Errorf("employee: Pool, Events and Verifier are required")
	}
	r := newRepos(d.Pool)
	pub := stream.NewEventPublisher(d.Events)
	return NewHTTPHandler(
		commands.NewCreateEmployeeHandler(r.employees, r.departments, r.positions, pub),
		commands.NewTerminateEmployeeHandler(r.employees, pub),
		commands.NewCreateDepartmentHandler(r.departments),
		commands.NewCreatePositionHandler(r.positions),
		queries.NewGetEmployeeHandler(r.employees),
		queries.NewListEmployeesHandler(r.employees),
		queries.NewListDepartmentsHandler(r.departments),
		queries.NewListPositionsHandler(r.positions),
		d.Verifier,
	), nil
}

// Consumers returns the stream handlers of this service, keyed by event type.
func Consumers(d Deps) map[string]events.Handler {
	r := newRepos(d.Pool)
	create := commands.NewCreateEmployeeHandler(r.employees, r.departments, r.positions, stream.NewEventPublisher(d.Events))
	return map[string]events.Handler{
		consumers.UserRegisteredType: consumers.NewUserRegistered(d.Pool, r.departments, r.positions, create).Handle,
	}
}
