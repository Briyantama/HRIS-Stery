---
description: Go API rules for HRIS-Stery (one module in apps/api).
globs: ["apps/api/**/*.go", "apps/api/go.mod"]
---

# Go Rules

## Layout (one module, `github.com/Briyantama/HRIS-Stery`, dir `apps/api`)

```text
cmd/api/main.go            run(ctx, getenv): config, pgxpool, Dragonfly client, services, one http.Server
cmd/api/routes.go          newMux(...) and the route table
cmd/migrate/main.go        goose up|down, reads DATABASE_URL
internal/config            Load(getenv) (Config, error); secrets redacted in String()
internal/db/migrations     goose files NNNN_name.sql (-- +goose Up / -- +goose Down)
internal/testdb            NewMigratedURL(t) for repository tests
services/<name>/
  domain/                  aggregates, value objects, domain events (no infrastructure imports)
  application/commands     write side
  application/queries      read side
  infrastructure/          pgx/sqlc repositories, Redis stores, event adapters
  deps.go, http.go         Deps, Build(Deps), HTTPHandler.Register(*http.ServeMux)
```

Shared infrastructure (`httpx`, `health`, `objstore`, `migrate`, `testdb`, `pgtenant`, `events`, `authn`) comes from `github.com/Briyantama/stery-kit/go`, replaced locally by `../../../stery-kit/go`. Do not copy those packages into this repo.

## Mandatory patterns

- `pgx/v5` + `pgxpool` and sqlc for data access. No SQL strings in domain or application code.
- Tenant-scoped repositories run every call inside the kit's tenant transaction helper (`set_config('app.tenant_id', ..., true)`). `auth.permissions` and `auth.tenants` skip it.
- sqlc `:execresult` calls check `RowsAffected() == 0` and return the domain not-found error.
- Handlers answer only through `httpx.WriteSuccess` / `httpx.WriteError` (envelope `{success, data, error}`) and decode bodies with `httpx.DecodeJSON`.
- Tenant and actor ids come from `authn.FromContext(ctx)`, never from the request body.
- Logging is `log/slog`: `slog.InfoContext(ctx, "msg", "tenant_id", tid)`. Never log tokens, passwords or PII.
- `context.Context` is the first argument of every I/O function. Wrap errors with `%w`.
- Constructor injection only. No `init()`, no package-level DB, Redis or publisher variables.
- `TenantID` value object for tenant identity in domain code.
- Tests live in external `_test` packages, table-driven or Arrange/Act/Assert. DB tests use `testdb.NewMigratedURL(t)` and skip when `TEST_DATABASE_URL` is unset.

## Forbidden

- `interface{}` or `any` in domain types
- Business logic in `http.go`
- `err.Error()` from the database or infrastructure in a response; use a fixed sentence, `internal error` for 500
- A migration without a `-- +goose Down` section
- An event consumer without a processed-events idempotency table
- zap, OpenTelemetry, Prometheus, gRPC, NATS, golang-migrate: superseded 2026-11-08 by ADR-0005 and ADR-0006, do not reintroduce
