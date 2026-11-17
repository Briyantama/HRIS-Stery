# 00 - Intent

## Statement

Move HRIS-Stery onto the SELA stack: one Go module (`github.com/Briyantama/HRIS-Stery`) with one API binary and one migrate binary, former services as packages (auth, employee, attendance, leave, notification, audit, document), `net/http` ServeMux, `log/slog`, goose migrations, Dragonfly for cache, rate limits and events, the `{success,data,error}` envelope, Postgres 18, one `scripts/check.sh` and SELA-style CI.

## Why now

The owner runs three products from one workspace. SELA's stack is the one already proven end to end on this machine. Keeping a second and third stack (Laravel 11 gateway (Sanctum) calling seven Go gRPC services through grpc-gateway, NATS JetStream, golang-migrate per service, zap, OpenTelemetry, Prometheus, MinIO, PgBouncer) costs setup time, review rules and CI jobs for no product gain.

## Expected outcome

- `apps/api` builds one binary that serves every route and runs the event consumers in-process.
- Migrations live in `apps/api/internal/db/migrations` and pass up, down, up.
- Domain and application code moves with its unit tests; behavior listed in `01-context.md` is unchanged.
- Frontend toolchain versions match SELA (`apps/web`); UI libraries stay.

## Non-goals

- Merging HRIS-Stery and POS-Stery.
- New features, redesigns, i18n.
- Replacing pgx/sqlc repositories with `database/sql`.

## Success metrics

| Metric | Target | Source |
| --- | --- | --- |
| `go build ./... && go vet ./... && go test ./...` in `apps/api` | green, DB tests not skipped | Verification 1 |
| `cmd/migrate up`, `down`, `up` | exit 0 three times | Verification 2 |
| Event publish, ack, redeliver test | pass | Verification 3 |
| Smoke against the running API | login, list, event in stream `hris:events` | Verification 4/5 |
| `npm run check && npm run build` per frontend | exit 0 | Verification 6 |
