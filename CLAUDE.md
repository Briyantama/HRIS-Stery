# CLAUDE.md

Guidance for Claude Code in this repository.

## Project

Multi-tenant HRIS SaaS for the Indonesian market. `docs/FSD_*` is the technical reference; architecture decisions are in `docs/adr/` (0002 RLS, 0004 Phase 2 deferral, 0005 single-module Go API, 0006 Dragonfly streams; 0001 and 0003 are superseded). No architectural change without an ADR.

Phase 1 services: auth, employee, attendance, leave, notification, audit, document. Phase 2 (payroll, tax-rules, recruitment, performance) stays unimplemented (ADR-0004); routes, if added, return 501.

## Layout

```
apps/api/                     one Go module github.com/Briyantama/HRIS-Stery (Go 1.25.1)
  cmd/api/                    HTTP server: main.go (run(ctx, getenv)), routes.go (newMux)
  cmd/migrate/                goose up|down, reads DATABASE_URL
  internal/config/            Load(getenv) with fail-fast validation
  internal/db/migrations/     goose SQL files NNNN_name.sql (0001 bootstrap, auth 0011-0013, employee 0021-0022,
                              attendance 0031-0032, leave 0041-0042, notification 0051, audit 0061, document 0071)
  internal/testdb/            template-clone test databases (wraps stery-kit/testdb)
  services/<name>/            auth employee attendance leave notification audit document
    domain/ application/{commands,queries}/ infrastructure/   business logic, ports, adapters
    deps.go                   Deps + Build(d) (+ ConsumerGroup/Consumers for event consumers)
    http.go                   HTTPHandler.Register(mux): routes "VERB /api/v1/..."
apps/web/                     SvelteKit 2 + Svelte 5, Tailwind 4, TanStack Query, Zod (adapter-auto)
deploy/docker-compose.yml     Dragonfly + RustFS (Postgres runs outside compose)
scripts/check.sh              quality gate, stages go and web
```

Shared Go and web code lives in the sibling repo `../stery-kit` (D5, ADR-0005): Go module `github.com/Briyantama/stery-kit/go` (httpx, health, objstore, migrate, testdb, pgtenant, events, authn, ratelimit) and npm package `@stery/kit` (`stery-kit/web`). `apps/api/go.mod` points at it with `replace ... => ../../../stery-kit/go`. Build and test need that checkout; change kit code in the kit repo, not here.

## Commands

```bash
make up        # compose up (Dragonfly, RustFS), migrate up, start the API in the background (.run/api.pid)
make down      # stop the API and compose
make status    # API pid + curl /healthz
make build     # go build -o .run/api ./cmd/api
make check     # bash scripts/check.sh (needs TEST_DATABASE_URL)
```

In `apps/api`:

```bash
go run ./cmd/migrate up          # or down
go run ./cmd/api                 # :8000
go test ./services/leave/...     # one package
gofmt -l . && go vet ./...
```

Frontend (`apps/web`): `npm install`, `npm run dev`, `npm run check`, `npm run test:unit`. Node >= 20.19.

The full `scripts/check.sh` runs in GitHub CI. Locally run targeted Go tests, `npm run check` and Vitest. `go test -race` needs cgo and runs in CI only. Port 8080 on this machine belongs to an unrelated Apache; the API listens on 8000.

## Conventions

- Handlers respond only through `httpx.WriteSuccess` / `httpx.WriteError`; envelope `{success,data,error}`. Error text is a short fixed sentence; unexpected errors return 500 `internal error`, never `err.Error()`.
- Tenant, actor and employee-self ids come from `authn.FromContext`, never from the request body.
- Every route is public on purpose (`POST /api/v1/auth/login|refresh|register`, `GET /healthz`) or wrapped by `authn.Authenticate`; roles (`hr_admin`, `manager`, `employee`) via `authn.WithRole`.
- Status codes: 200 read/update/action, 201 create with `Location`, 400 bad JSON, 422 validation, 409 conflict, 404 missing or cross-tenant, 401, 403.
- JSON fields are `snake_case`. Paginated lists use `?page=&per_page=` (default 20, max 100) and return `{items,page,per_page,total}`.
- Logging: `log/slog` with `InfoContext`/`ErrorContext`; add `tenant_id` where known.
- Repositories use pgx/v5 + pgxpool (sqlc output where it exists). `:execresult` calls check `RowsAffected()==0` for not-found.
- Errors wrapped with `%w`. Tests in external `_test` packages, table-driven or Arrange/Act/Assert; DB tests use `testdb.NewMigratedURL(t)` and skip when `TEST_DATABASE_URL` is unset.
- Business rules live in `domain/` and `application/`; `infrastructure/` and `http.go` hold none. No `any`/`interface{}` in domain types; use the `TenantID` value object, never raw string compares.
- Migrations: one goose file per change with `-- +goose Up` and `-- +goose Down`, numbered inside the module's range. Never edit an applied migration.
- Comments only for a non-obvious why. Secrets go in `.env`, never in tracked files.

## Invariants

- Postgres RLS on every `tenant_id` table. App role `hris_app` has no BYPASSRLS; `hris_admin` does. Every table: `id UUID PK default gen_random_uuid()`, `tenant_id UUID NOT NULL`, `created_at`/`updated_at TIMESTAMPTZ`.
- Tenant-scoped repositories wrap every call in `pgtenant.WithTenantTx` (`set_config('app.tenant_id', ..., true)` inside the transaction). Non-tenant tables (`auth.permissions`, `auth.tenants`) skip it.
- `audit` entries are insert-only (no UPDATE, DELETE, TRUNCATE).
- Events go to the Dragonfly stream `hris:events` (ADR-0006). Envelope fields: `event_id` (UUIDv7), `event_type`, `tenant_id`, `actor_id`, `occurred_at`, `schema_version`, `payload`. Type naming `hris.{domain}.{entity}.{verb}`.
- Publish after the DB commit, never inside the transaction.
- Every consumer writes a processed-events row with `INSERT ... ON CONFLICT DO NOTHING`. A failed mark is logged and the entry stays un-acked (not swallowed). Consumers use the run context, never `context.Background()`.
- `GetPermissions` stays one query (no N+1).
- JWT RS256 with claims `sub`, `tid`, `roles[]`, `exp`, `iat`, verified locally with the public key; no round trip to auth.
- Never log tokens, passwords or PII.
