# Code Review Decision Tree

Pick the row that matches the change, run the skills in order, then ask the reviewer.

## New or changed endpoint (`apps/api/services/<name>/http.go`)

1. `adr-review`: does ADR-0005 or another ADR cover the change?
2. `architecture-review`: handler stays thin; tenant and actor ids come from `authn.FromContext`.
3. Implement in order: domain, application command or query, infrastructure, `http.go`, route in `Register`.
4. Add a row to the service's `http_test.go` table for each new status code.
5. Reviewer: `go-backend-reviewer`.

## Schema change

1. `migration-review`: next free number in the service range, Up and Down sections, RLS on tenant tables.
2. Run `go run ./cmd/migrate up`, `down`, `up` against a throwaway database.
3. Reviewer: `migration-reviewer`.

## Event publisher or consumer

1. `event-review`: full envelope, UUID v7 id, publish after commit.
2. Consumer: `processed_events` table, `ON CONFLICT DO NOTHING`, handler error leaves the entry pending.
3. Register the handler in `cmd/api` under the existing consumer group name.
4. Reviewer: `event-stream-reviewer`.

## Repository or query change

1. `go-style-review`: tenant transaction helper on every tenant-scoped call, `RowsAffected() == 0` means not found.
2. DB test with `testdb.NewMigratedURL(t)`.
3. Reviewer: `go-backend-reviewer`.

## Frontend change (`apps/web`)

Follow `rules/svelte.md`. No reviewer agent; run `npm run check` and `npm run test:unit`.

## Phase 2 request (payroll, tax rules, recruitment, performance)

Stop. ADR-0004 blocks it. A route, if one is added, returns 501.
