package audit

import (
	"errors"

	"github.com/Briyantama/stery-kit/go/authn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Briyantama/HRIS-Stery/services/audit/application/commands"
	"github.com/Briyantama/HRIS-Stery/services/audit/application/queries"
	"github.com/Briyantama/HRIS-Stery/services/audit/infrastructure/postgres"
)

// Deps is everything the audit service needs from the wiring code.
type Deps struct {
	Pool     *pgxpool.Pool
	Verifier authn.Verifier
}

func newRepo(pool *pgxpool.Pool) *postgres.AuditRepository { return postgres.NewAuditRepository(pool) }

// Build wires the repository, command/query handlers and the HTTP handler.
func Build(d Deps) (*HTTPHandler, error) {
	if d.Pool == nil || d.Verifier == nil {
		return nil, errors.New("audit: Pool and Verifier are required")
	}
	repo := newRepo(d.Pool)
	return NewHTTPHandler(
		commands.NewRecordAuditHandler(repo),
		queries.NewQueryAuditTrailHandler(repo),
		queries.NewGetAuditEntryHandler(repo),
		d.Verifier,
	), nil
}
