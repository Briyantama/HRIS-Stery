# HRIS-Stery — Claude Code Rules

This file governs all Claude Code behavior in this repository.
Read every section before generating any file. These rules are non-negotiable.

---

## 0. Pre-Code Checklist

Before writing any code, Claude must:

1. Read the relevant ADR in `docs/adr/` for the area being changed.
2. Identify which bounded context and service owns the change.
3. Confirm the proto contract exists in `proto/hris/{domain}/v1/` before writing handlers.
4. Confirm no business logic will land in Laravel controllers or gRPC handlers directly.
5. Confirm migrations have both `up` and `down` files.

If any of these cannot be confirmed, stop and ask.

---

## 1. Architecture Rules (Non-Negotiable)

### 1.1 Layer Boundaries

```
Frontend (SvelteKit)
  → REST/JSON → Laravel Gateway
    → HTTP/JSON (grpc-gateway) → Go Services (gRPC)
      ↔ gRPC (service-to-service)
      → NATS JetStream (async events)
```

- **Never** call another service's database schema directly. All cross-service data goes through gRPC or event projections.
- **Never** put business logic in Laravel controllers, gRPC handlers, or SvelteKit `+page.svelte` files.
- **Never** use `interface{}` or `any` in Go domain types. Only allowed in infrastructure adapters.
- **Never** use raw `string` for `tenant_id` comparisons in domain code. Use the `TenantID` value object.

### 1.2 Service Ownership

Each service owns exactly one PostgreSQL schema. The mapping is:

| Service | Schema |
|---|---|
| auth-service | `auth` |
| employee-service | `employee` |
| attendance-service | `attendance` |
| leave-service | `leave` |
| notification-service | `notification` |
| audit-service | `audit` |
| document-service | `document` |
| ai-service | no schema (stateless) |

### 1.3 Phase Boundaries

**Phase 1 (MVP) services only:**
`auth`, `employee`, `attendance`, `leave`, `notification`, `audit`, `document`, `ai`

**Phase 2 — Do not implement, do not scaffold business logic:**
`payroll`, `tax-rules`, `recruitment`, `performance`

Phase 2 routes must return `501 Not Implemented`. No payroll domain logic exists in MVP code.

---

## 2. Go Rules

- Go version: **1.24+** with modules. Run `go mod tidy` after any dependency change.
- Project layout per service:
  ```
  cmd/server/main.go          ← entrypoint only, no logic
  internal/domain/            ← aggregates, value objects, domain events
  internal/application/
    commands/                 ← write side (CQRS)
    queries/                  ← read side (CQRS)
  internal/infrastructure/
    postgres/                 ← sqlc-generated queries + repo implementations
    redis/                    ← cache + token store
  internal/interfaces/
    grpc/                     ← gRPC server handlers (call application layer only)
    http/                     ← grpc-gateway HTTP handlers (auto-generated)
  migrations/                 ← golang-migrate files (always reversible)
  ```
- Use **sqlc + pgx/v5** for all database access. No raw `database/sql` strings in domain code.
- Use **Uber zap** for structured logging. Every log entry must include `tenant_id` and `request_id`.
- Use **OpenTelemetry Go SDK** for tracing. Every gRPC handler and repository method must be instrumented.
- Return gRPC status errors: `status.Errorf(codes.NotFound, "employee %s not found", id)`.
- Use `context.Context` as the first argument to every function that touches I/O.
- Inject all dependencies via constructors. No `init()` functions, no package-level vars for DB/NATS/Redis.
- Every NATS consumer must have an idempotency table. Document the idempotency design in the PR.
- Wrap errors with context: `fmt.Errorf("creating employee: %w", err)`.
- Write unit tests for all use case (application layer) logic. Mock at the repository interface boundary.

### Go Forbidden Patterns

```go
// FORBIDDEN — business logic in handler
func (h *Handler) CreateEmployee(ctx context.Context, req *pb.CreateEmployeeRequest) (*pb.EmployeeResponse, error) {
    if req.Email == "" { /* validation here */ }  // ← use application layer
    db.Exec("INSERT INTO ...")                     // ← never raw SQL in handlers
}

// FORBIDDEN — raw tenant_id string
tenantID := ctx.Value("tenant_id").(string)
db.Query("SELECT * FROM employees WHERE tenant_id = $1", tenantID) // ← use TenantID value object + RLS

// FORBIDDEN — global state
var db *sql.DB  // package-level
```

---

## 3. Laravel Rules

Laravel is **API Gateway and Auth layer only**. It is not a domain service.

- Laravel version: **11.x**
- Auth: **Laravel Sanctum** in API token mode. Tokens are short-lived JWTs signed with RSA. See ADR-0005.
- RBAC: **Spatie Permission**. Roles: `hr_admin`, `manager`, `employee`. Define in seeder.
- Controller maximum: **20 lines** of logic. Anything longer belongs in a Service class.
- Form Requests: every POST/PUT route must use a Form Request class for validation.
- Service classes in `app/Services/` call Go services via HTTP/JSON (grpc-gateway). No gRPC-PHP extension.
- Use **Guzzle** (built into Laravel) to call grpc-gateway endpoints. Wrap in typed Service classes.
- Error handling: catch all exceptions in `app/Exceptions/Handler.php`. Return JSON with `code`, `message`. Never expose stack traces.
- Use **Laravel Horizon** for queue monitoring. Redis driver only.
- Enable **Laravel Pint** (PSR-12). Run `./vendor/bin/pint` before commit.
- Multi-tenancy: `TenantResolver` middleware extracts `tenant_id` from the validated JWT and sets it on the authenticated user model. Every controller can access `auth()->user()->tenant_id`.

### Laravel Forbidden Patterns

```php
// FORBIDDEN — business logic in controller
public function store(Request $request) {
    $salary = $request->base * 1.13; // ← this is domain logic
    DB::table('employees')->insert(...); // ← no raw DB in controllers
}

// FORBIDDEN — calling another service's DB
DB::connection('employee_db')->table('employees')->get(); // ← use gRPC only

// FORBIDDEN — fat controller (> 20 lines of logic)
```

---

## 4. SvelteKit Rules

- SvelteKit version: **2.x** with Svelte **5** (runes mode).
- TypeScript everywhere. No `.js` files in `src/`.
- TanStack Query (`@tanstack/svelte-query`) for all server state. No `fetch` calls in `+page.svelte` scripts directly.
- Zod schemas in `src/lib/schemas/` for all form validation and API response types.
- shadcn-svelte for UI components. Do not write raw HTML form inputs without a component wrapper.
- Auth guard in `hooks.server.ts`. Every protected route returns a redirect if no valid session.
- Use SvelteKit route groups: `(auth)` for public, `(app)` for protected.
- Stores (`$lib/stores/`) for UI-only state (sidebar open/closed, theme). Never store server data in Svelte stores — use TanStack Query cache.
- OpenTelemetry in `instrumentation.server.js` for all server-side traces.
- E2E tests with Playwright in `tests/e2e/`. Unit tests with Vitest in `tests/unit/`.

### SvelteKit Forbidden Patterns

```svelte
<!-- FORBIDDEN — fetch in page script -->
<script lang="ts">
  const res = await fetch('/api/employees'); // ← use createQuery
  const data = await res.json();
</script>

<!-- FORBIDDEN — business logic in page -->
<script lang="ts">
  const tax = salary * 0.05; // ← belongs in a utility or comes from API
</script>
```

---

## 5. Proto / Buf Rules

- All protos under `proto/hris/{domain}/v1/`. Breaking changes require `v2/`.
- Run `buf lint` before committing any `.proto` file. Zero warnings allowed.
- Run `buf breaking --against .git#branch=main` on every PR touching `proto/`.
- Use `google.protobuf.Timestamp` for all time fields (never `string` dates in protos).
- Use `google.protobuf.FieldMask` for partial update RPCs.
- Every RPC must have an HTTP annotation for grpc-gateway (see `proto/hris/auth/v1/auth.proto` as reference).
- Message naming: `{Verb}{Entity}Request` / `{Verb}{Entity}Response`. No ambiguous names.

---

## 6. Event / NATS Rules

- All events must conform to the envelope schema: `event_id` (UUID v7), `event_type`, `tenant_id`, `actor_id`, `occurred_at`, `schema_version`, `payload`.
- Subject taxonomy: `hris.{domain}.{entity}.{verb}` (all lowercase, dot-separated).
- Every consumer service must maintain an `idempotency_processed_events` table. Log `event_id` before processing. Use `INSERT ... ON CONFLICT DO NOTHING` to skip duplicates.
- Never publish events from within a database transaction. Publish after the transaction commits (transactional outbox pattern if needed).
- JetStream stream name: `HRIS_EVENTS`. Consumer names: `{service-name}-{subject-slug}`.

---

## 7. Database / Migration Rules

- Every `up` migration must have a corresponding `down` migration. CI will run `up → down → up` to verify.
- Use `golang-migrate` in all Go services. Migration files: `{NNN}_{description}.up.sql` / `{NNN}_{description}.down.sql`.
- PostgreSQL RLS is mandatory on every table that contains `tenant_id`. See ADR-0002 and `services/auth-service/migrations/` for reference policy patterns.
- Every table must have: `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`, `tenant_id UUID NOT NULL`, `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`, `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`.
- No `TRUNCATE` or `DELETE` in `audit.audit_events` — the table has an INSERT-ONLY RLS policy.

---

## 8. Security Rules

- JWT payload must include: `sub` (user ID), `tid` (tenant ID), `roles` (array), `exp`, `iat`.
- Go services validate JWT signature using the public key from `AUTH_PUBLIC_KEY` env var. No round-trip to auth-service per request.
- Refresh tokens stored in Redis with TTL. Revocation via Redis key deletion.
- Never log JWT tokens, passwords, or PII. Mask fields in structured logs.
- Rate limiting at gateway level (Laravel): per-tenant, per-endpoint. Redis-backed counters.
- NATS subjects secured with ACLs. Each service has a dedicated NATS user with minimal permissions.

---

## 9. Observability Rules

- Every gRPC handler must create a child span: `tracer.Start(ctx, "EmployeeService.CreateEmployee")`.
- Every repository method must create a child span with DB statement as attribute (sanitized — no values).
- Metrics to expose per service: `rpc_requests_total`, `rpc_duration_seconds`, `db_query_duration_seconds`, `nats_messages_published_total`, `nats_messages_consumed_total`.
- All services export to OTLP collector at `OTEL_EXPORTER_OTLP_ENDPOINT` env var.

---

## 10. Comment Policy

Write no comments except when the **why** is non-obvious:
- A hidden constraint or regulatory requirement
- A workaround for a specific external system bug
- An invariant that would surprise a future reader

Do not write comments that explain **what** the code does. Well-named identifiers do that.
Do not reference the current task, issue number, or PR in code comments.

---

## 11. What Claude Must Never Do

- Generate payroll calculation logic in Phase 1. Return `501` and stop.
- Generate recruitment or performance review business logic in Phase 1.
- Write a Laravel controller longer than 20 lines of logic.
- Add `WHERE tenant_id = ?` without also checking that RLS is active on the table.
- Create a NATS consumer without an idempotency table.
- Write a migration without a `down` file.
- Use `interface{}` or `any` in a domain type.
- Call an external service (OpenAI, etc.) without a per-tenant rate limit check.
- Push secrets, API keys, or connection strings into any tracked file.
- Implement a feature not in the Phase 1 service list without explicit instruction.
