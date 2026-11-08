package audit_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Briyantama/stery-kit/go/events"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Briyantama/HRIS-Stery/internal/testdb"
	"github.com/Briyantama/HRIS-Stery/services/audit"
)

func TestConsumers_RecordOnceAndRetryOnFailure(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.NewMigratedURL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	handlers := audit.Consumers(audit.Deps{Pool: pool})
	const typ = "hris.workforce.employee.created"
	h, ok := handlers[typ]
	if !ok {
		t.Fatalf("no handler for %s", typ)
	}
	env := events.Envelope{
		EventID: uuid.NewString(), EventType: typ, SchemaVersion: 1,
		TenantID: tenant, ActorID: "u-1", OccurredAt: time.Now(), Payload: json.RawMessage(`{}`),
	}

	// Delivered twice: one audit entry (idempotent).
	for range 2 {
		if err := h(ctx, env); err != nil {
			t.Fatalf("handler: %v", err)
		}
	}
	var n int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit.entries WHERE tenant_id = $1 AND action = 'EMPLOYEE_CREATED'`, tenant).Scan(&n); err != nil || n != 1 {
		t.Fatalf("entries = %d, err = %v; want 1", n, err)
	}

	// Missing actor: recorded as "system" so the event is not retried forever.
	env.EventID, env.ActorID = uuid.NewString(), ""
	if err := h(ctx, env); err != nil {
		t.Fatalf("event without actor_id: %v", err)
	}
	var actor string
	if err := pool.QueryRow(ctx, `SELECT actor_id FROM audit.entries WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT 1`, tenant).Scan(&actor); err != nil || actor != "system" {
		t.Fatalf("actor = %q, err = %v; want system", actor, err)
	}
}
