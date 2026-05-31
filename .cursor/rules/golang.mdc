---
description: Go microservice rules for HRIS-Stery.
globs: ["services/**/*.go", "services/**/*.mod"]
---

# Go Rules

## Project Layout (per service)

```
cmd/server/main.go                  ← entrypoint: wire deps, start server
internal/
  domain/                           ← aggregates, value objects, domain events (pure Go, no imports)
  application/
    commands/                       ← write side use cases
    queries/                        ← read side use cases
  infrastructure/
    postgres/                       ← sqlc queries + repo implementations
    redis/                          ← token store, cache
  interfaces/
    grpc/                           ← gRPC server handlers
    http/                           ← grpc-gateway (auto-generated, do not edit)
migrations/                         ← golang-migrate files
```

## Mandatory Patterns

- sqlc + pgx/v5 for all DB access. No raw SQL strings in application or domain layers.
- Uber zap for structured logging. Every log must include tenant_id and request_id fields.
- OpenTelemetry spans on every gRPC handler and repository method.
- `context.Context` as first argument to all I/O functions.
- Constructor injection for all dependencies. No package-level vars for DB/NATS/Redis.
- TenantID value object — never compare raw strings for tenant identity.
- Return gRPC status errors: `status.Errorf(codes.NotFound, "...")`.
- Wrap errors: `fmt.Errorf("create employee: %w", err)`.
- Use `services/_shared/postgres/rls.go` for setting RLS session on connections.

## Forbidden

- `interface{}` or `any` in domain types
- Business logic in `interfaces/grpc/` handlers
- Direct DB calls from `interfaces/` layer
- Package-level `var db *sql.DB`
- Missing `down` migration for any `up` migration
- NATS consumer without idempotency table
