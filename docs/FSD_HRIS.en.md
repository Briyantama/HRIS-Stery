# FSD HRIS-Stery

*Functional Specification Document, multi-tenant Human Resource Management System (HRIS). Describes the behaviour that already exists in the code (status + evidence) and the target architecture aligned to the standard Stery Suite stack (SELA).*

| | |
|---|---|
| **Document Name** | FSD HRIS-Stery |
| **Version** | 1.0 (Draft) |
| **Date** | 6 November 2026 |
| **Prepared by** | Product & Engineering Team |
| **Status** | Draft for review |
| **Related Documents** | `HRIS-Stery/docs/BRD_HRIS.id.md`, `HRIS-Stery/docs/PRD_HRIS.id.md`, `stery-suite/docs/00_Analisis_Dampak_Risiko.id.md`, `stery-suite/docs/BRD_Stery_Suite.id.md`, `stery-suite/docs/PRD_Stery_Suite.id.md`, `stery-suite/docs/FSD_Stery_Suite.id.md`, `SELA/CLAUDE.md`, `SELA/docs/FSD_Sela.md`, `HRIS-Stery/docs/adr/0001..0004` |
| **Changes from previous** | Changes from previous: First issue |

---

## 1. Introduction

### 1.1 Purpose of the document

This document is the functional and technical specification of HRIS-Stery (product code `HRS`). It does two things that are deliberately kept apart:

1. **Current state (as-is):** the functional behaviour that actually exists in the code, each item given a status and a file-path evidence. Truth = code.
2. **Target state (to-be):** the HRIS architecture after alignment to the standard Stery Suite stack sourced from SELA (STACK-SPEC in 2.1) and decisions D-HRS-01..12 (BRD 9.3). The as-is to to-be gap is recorded in the delta table 2.7.

This document changes no code. Stack alignment is a workstream scheduled on the roadmap (BRD 12, PRD 11), not part of writing this document.

### 1.2 Scope and related documents

| Document | Role | Reference from this FSD |
|---|---|---|
| `BRD_HRIS.id.md` | Business requirements BR-HRS-01..22, decisions D-HRS, open questions Q-HRS, risks RISK-HRS | Matrix 11.2 maps every BR to FR |
| `PRD_HRIS.id.md` | User stories US-HRS, personas, NFR-HRS, roadmap R0-R5 | Every FR in section 4 derives from the PRD 4 MoSCoW features |
| `00_Analisis_Dampak_Risiko.id.md` | Cross-product impact and risk analysis | Sections 2.7 and 11.3 |
| `FSD_Stery_Suite.id.md` | Platform contracts (ports, SessionService, EntitlementService, IntegrationService) | Sections 2.3, 2.5, 7.2 |
| `SELA/CLAUDE.md`, `SELA/docs/FSD_Sela.md` | Source of STACK-SPEC | Section 2.1 |
| `docs/adr/0001..0004` | Existing HRIS architecture decisions | Sections 2.7 and 4.09-4.12 |

In scope: modules 01-08 (AUTH, EMP, ATT, LEV, NTF, AUD, DOC, AI) and cross-module security requirements (`FR-HRS-SEC.*`). Modules 09-12 (PAY, TAX, REC, PRF) are **Deferred** (ADR-0004) and only have a deferral rationale section, with no FR.

### 1.3 Conventions

- **Status** (fixed vocabulary): `Selesai` (behaviour verified correct in code), `Sebagian` (exists but defective or incomplete, defect described), `Belum` (not in code), `Pending Verification` (failure inferred from static reading, not yet run), `Ditunda` (officially deferred, ADR-0004).
- **Evidence:** file paths relative to the `HRIS-Stery` repo (without the `HRIS-Stery/` prefix), in a column titled `Evidence`. Line numbers, when needed, are written outside the backticks. Target paths (not yet existing) appear only in the delta table 2.7 and in text, never in the `Evidence` column.
- **Priority:** `M` (Must), `S` (Should), `C` (Could).
- **IDs:** `FR-HRS-<MM>.<n>` (MM = catalogue module number), `FR-HRS-SEC.<n>` for cross-module security. BR, D, Q, RISK are defined in the BRD; US in the PRD; NFR-HRS is referenced in section 9. SELA is cited by its own IDs (e.g. `SELA FR-05.1`).
- **Business values not found in the repo** (prices, SLAs, quotas, thresholds) are written `tbd` with a Q-HRS reference. This document contains no `Rp` figure.
- **Estimates** (effort, readiness) are marked `[INFERENCE]`.

### 1.4 Origin of evidence

All "current" claims come from static reading of the code and configuration files of the `HRIS-Stery` repo (branch `development`, last commit 10 June 2026). **No build, test, linter or migration was run** to prepare this document; therefore a claim that a step "fails" is marked `Pending Verification` unless the failure is visible directly from the text (e.g. invalid SQL syntax). Some initial scan results from the assistant turned out to be wrong after re-reading; the corrections are recorded in 11.3.

---

## 2. Architecture and Stack

### 2.1 Stack (STACK-SPEC, source SELA)

The target stack of every Stery Suite product is the SELA stack (owner decision; D-HRS-11). The table below is copied from STACK-SPEC. The `Sumber (SELA)` column points to the SELA file that is the reference.

| # | Layer | Standard | Source (SELA) |
|---|---|---|---|
| 1 | Repo layout | per product: `apps/api`, `apps/web`, `deploy/`, `scripts/`, `docs/`, `Makefile`, `CLAUDE.md`, `AGENTS.md` | `SELA/CLAUDE.md` ("Layout" section); `SELA/docs/FSD_Sela.md` 2.6 |
| 2 | Language | Go >= 1.25.1, one module per product `github.com/<owner>/<product>` (owner handle `Briyantama` as in SELA, unverified for new repos) | `SELA/CLAUDE.md`; `SELA/apps/api/go.mod` |
| 3 | Decomposition | module = Go package `services/<name>` in one process; core `Service` + Postgres repository + HTTP adapter + optional gRPC adapter; boundaries via Go interfaces and proto contracts so they can be split later | `SELA/CLAUDE.md` ("Layout", "Conventions"); `SELA/docs/FSD_Sela.md` 2.2 |
| 4 | Transport | REST `/api/v1` on stdlib `net/http` ServeMux, envelope `{success,data,error}` via `kit/httpx`; gRPC (`api/proto/v1`, buf, committed `gen/go`, `scripts/proto-gen.sh --check`) for machine/internal clients with a deny-by-default interceptor; no gateway tier; Caddy only as reverse proxy | `SELA/docs/FSD_Sela.md` 2.3, 6 |
| 5 | Database | PostgreSQL 18 (`uuidv7()`), pgx/v5 raw SQL (no sqlc, no ORM), goose v3.26.0 embedded migrations + `cmd/migrate up\|down`, `testdb.NewMigrated` (template cloning) | `SELA/CLAUDE.md` ("Conventions"); `SELA/docs/FSD_Sela.md` 5 |
| 6 | Tenant isolation | per D8 (HRS, POS, platform): RLS ENABLE+FORCE, `tenant_isolation` with `USING` and `WITH CHECK`, `current_setting('app.tenant_id', true)::uuid`, set per transaction by `kit/tenantdb.WithTenant`; app role NOBYPASSRLS, separate migration role | Suite addition (SELA has no tenant concept); pattern from `SELA/CLAUDE.md` ("Invariants") |
| 7 | State store | Dragonfly via go-redis/v9: sessions (hash only), rate limits, quotas, queues, Lua scripts | `SELA/docs/FSD_Sela.md` 2.4 |
| 8 | Object storage | RustFS (S3) via aws-sdk-go-v2 behind `objstore.Store`; presigned URLs only; ClamAV `clamd` scan fail-closed on upload paths | `SELA/docs/FSD_Sela.md` 2.1, 8 |
| 9 | Authentication | opaque session in HttpOnly Secure cookie, hash in Dragonfly; RBAC tables (roles, permissions, actions, role_actions) + `RequirePermission(code)`; no JWT, no tokens in browser storage | `SELA/docs/FSD_Sela.md` 3, 8 |
| 10 | Async | in-process ticker workers + Postgres outbox + idempotent consumers; no NATS/Kafka | `SELA/docs/FSD_Sela.md` 2.5 |
| 11 | Observability | `log/slog`, `/healthz`; see D11 (OpenTelemetry/Prometheus is an owner decision `tbd`, Q-HRS-08) | `SELA/CLAUDE.md` |
| 12 | Configuration | environment variables, fail-fast validation, secrets never logged, `.env.example` | `SELA/CLAUDE.md` ("Conventions") |
| 13 | PII and audit | `kit/piicrypt` (AES-256-GCM, per-purpose AAD, blind index, key rotation), `kit/audit` append-only `audit_logs` with DB triggers | `SELA/docs/FSD_Sela.md` 8 |
| 14 | Payments | `payment.Provider`, Midtrans QRIS (Xendit has been removed), signed webhook. **Not used directly by HRIS:** suite billing lives in the platform (D5); HRIS only reads entitlements | `SELA/docs/FSD_Sela.md` 2.1 |
| 15 | Web | SvelteKit 2 + Svelte 5 + Vite 8 + TypeScript 6, `ssr=false`, `adapter-static`, npm, Node >= 20.19; `apiFetch` + envelope mirror; i18n `{id,en}` with parity test; Vitest 4 + Testing Library; Playwright (Chrome channel) | `SELA/CLAUDE.md`; `SELA/docs/FSD_Sela.md` 2.1, 10 |
| 16 | Deploy | Caddy; static web; `deploy/docker-compose.yml` for Dragonfly, RustFS, ClamAV profile; default standalone ports API 8080, gRPC 9090, web 5173 | `SELA/docs/FSD_Sela.md` 2.1 |
| 17 | Quality gate | `scripts/check.sh` (proto drift, gofmt, vet, golangci-lint, `go test -cover`, svelte-check, vitest coverage, real-stack e2e; `CHECK_STAGES`), GitHub Actions jobs go/web/e2e, 80% coverage floor, `-race` in CI | `SELA/CLAUDE.md` |
| 18 | Process | TDD RED/GREEN, `docs/tracking/tracker.json` + `scripts/build_tracker.py`, FSD updated together with code, `docs/aidlc/` artefacts | `SELA/CLAUDE.md` |
| 19 | Locale | Indonesian default, English supported, IDR, WIB | `SELA/CLAUDE.md` |

Note: in `standalone` mode (default) HRIS uses the stack above **with** its own users, sessions and tenants; `suite` mode only swaps the identity, entitlement, notification, audit and event adapters (2.3).

Implementation status 2026-11-08 (ADR-0005, ADR-0006). Rows 1-3, 5, 11, 12 and 17 are in place: one module `apps/api` with `cmd/api`, `cmd/migrate` and `services/<name>`, goose migrations in `apps/api/internal/db/migrations`, slog, Postgres 18, `scripts/check.sh` and CI. Shared code comes from the sibling repo `stery-kit` (`httpx`, `health`, `objstore`, `migrate`, `testdb`, `pgtenant`, `events`, `authn`, `ratelimit`; npm `@stery/kit`), not from a `kit/` folder in a suite repo. Where the implementation differs from this table:

| Row | As built | Why |
|---|---|---|
| 4 | REST only; no gRPC adapter, no `api/proto`, no `gen/go` | no machine client exists (ADR-0005, option C rejected) |
| 7, 10 | Dragonfly stream `hris:events` with four consumer groups and `processed_events` idempotency; no Postgres outbox; events published after commit | ADR-0006 |
| 9 | RS256 JWT verified in the API (`authn.Authenticate`, roles per route), refresh tokens and revocation list in Dragonfly; the web keeps the access token in an HttpOnly cookie (`hris_session`) behind a SvelteKit `/api` proxy, never in browser storage | ADR-0005; opaque sessions stay the target (FR-HRS-SEC.4) |
| 15 | SvelteKit 2.63 + Svelte 5 + Vite 8 + TypeScript 6 with server routes (login action, `/api` proxy, `adapter-auto`), not `ssr=false` + `adapter-static` | the HttpOnly-cookie BFF needs a server |
| 16 | `deploy/docker-compose.yml` (Dragonfly, RustFS, optional `api` profile) and `docker-compose.production.yml`; API port 8000 | ports kept from the previous gateway |

### 2.2 Module decomposition

Current state: seven standalone Go services, each with its own process, Postgres schema and gRPC port, plus an empty ai-service. Target: a single `apps/api` process with `services/<name>` modules.

| Current service | Schema | Default gRPC port | Target module (without repo prefix) | Status | Evidence |
|---|---|---|---|---|---|
| auth-service | `auth` | 50051 | `services/auth` (module 01) | Partial | `services/auth-service/cmd/server/main.go` |
| employee-service | `employee` | 50052 | `services/employee` (module 02) | Partial | `services/employee-service/cmd/server/main.go` |
| attendance-service | `attendance` | 50053 | `services/attendance` (module 03) | Partial | `services/attendance-service/cmd/server/main.go` |
| leave-service | `leave` | 50054 | `services/leave` (module 04) | Partial | `services/leave-service/cmd/server/main.go` |
| notification-service | `notification` | 50055 | `services/notification` (module 05) | Partial | `services/notification-service/cmd/server/main.go` |
| audit-service | `audit` | 50054 (collides with leave-service) | `services/audit` (module 06) | Partial | `services/audit-service/cmd/server/main.go` |
| document-service | `document` | 50056 | `services/document` (module 07) | Partial | `services/document-service/cmd/server/main.go` |
| ai-service | (none) | (none) | `services/ai` (module 08) | Not started | `go.work` (line 13 commented out); `proto/hris/ai/v1/ai.proto` |

Observations:
- `leave-service` and `audit-service` both bind the default gRPC `50054`; running both on one host without overriding the configuration will collide (`Pending Verification`, depends on env).
- `ai-service` only has a proto contract (3 RPCs); the `cmd/server/` directory is empty and the service is commented out of `go.work`.
- There is a shared package `services/_shared` (config, database, errors, middleware, nats, observability, postgres, server, telemetry) that is the raw material for the target `kit/*` (Evidence: `services/_shared/postgres/rls.go`, `services/_shared/server/grpc_interceptors.go`).
- The internal structure of each service already follows domain/application/infrastructure/interfaces (light DDD). These domain and application parts are what is kept during alignment (2.7).

### 2.3 Interface components

**Current (as-is):**

| Component | Behaviour | Status | Evidence |
|---|---|---|---|
| SvelteKit web (SSR) | Four pages; login calls `http://localhost:3000/api/auth/login` (hard-coded value, not a real gateway); the API client does not send an Authorization header | Partial | `apps/web/src/routes/(auth)/login/+page.server.ts`, `apps/web/src/lib/api/client.ts` |
| Laravel 11 gateway | Routes `auth/login`, `auth/refresh`, `auth/logout`, `health`, `/v1/leaves*`, `leave-balance/{employeeId}`, `leave-types`; employee/attendance/document/audit routes are only comments (lines 100-108); the `throttle` middleware is never applied; Sanctum and spatie/permission are installed but unused | Partial | `apps/api-gateway/routes/api.php`, `apps/api-gateway/composer.json` |
| Gateway client to services | `BaseGrpcGatewayClient` calls the grpc-gateway HTTP path; `AuthServiceClient` calls the real gRPC path `/hris.auth.v1.AuthService/Login`, not `/v1/auth/login` | Partial | `apps/api-gateway/app/Services/BaseGrpcGatewayClient.php`, `apps/api-gateway/app/Services/AuthServiceClient.php` |
| grpc-gateway HTTP path | There is no `runtime.NewServeMux` or `ListenAndServe` in `services/**` outside `*.pb.gw.go`; the HTTP path the gateway calls is never served (11.3) | Not started | `services/auth-service/cmd/server/main.go` |
| Inter-service gRPC | Each service's gRPC server with a shared interceptor; no authentication/authorisation in the interceptor; no mTLS | Partial | `services/_shared/server/grpc_interceptors.go` |
| Metrics HTTP | `StartMetricsServer` binds `:8081` for all services | Partial | `services/_shared/observability/metrics.go` |

**Target (to-be), per D-HRS-12 and STACK-SPEC 4, 9:**

| Target component | Behaviour |
|---|---|
| `apps/web` (SPA) | SvelteKit `ssr=false`, `adapter-static`; calls REST `/api/v1` with a session cookie; no tokens in browser storage |
| REST `/api/v1` | HTTP adapter in each module on stdlib ServeMux; envelope `{success,data,error}`; explicit permission per route (`RequirePermission`); tenant and actor from the session |
| gRPC `api/proto/v1` | Machine/internal contracts and platform client; deny-by-default interceptor; HMAC service token |
| Suite ports (`suite` mode) | `Identity`, `Tenant`, `Entitlement`, `Notification`, `Audit`, `Events` with two adapters (standalone, suite); Go interface details in `FSD_Stery_Suite.id.md` 2.4 |

The `standalone` adapter uses HRIS's own user, session, role and audit tables. The `suite` adapter calls the platform's `SessionService/Introspect` (30 s local cache), `EntitlementService/Check` (+ signed snapshot, tolerance `tbd`, Q-HRS-11), and `IntegrationService/Publish` (D-HRS-03..05).

### 2.4 Dragonfly usage

**Current:** Redis 7 is used only by auth-service: keys `refresh_token:<sha256>` and `revoked_token:<hash>`; there is no per-family refresh token reuse detection. The department/employee caches in employee-service and leave-service are in-process maps with a 24-hour TTL, not Redis.

| Usage | Status | Evidence |
|---|---|---|
| Refresh tokens and revocation list in Redis | Partial | `services/auth-service/internal/infrastructure/token_service.go`, `services/auth-service/internal/application/commands/refresh_token.go` |
| In-process department cache (24-hour TTL) | Partial | `services/employee-service/internal/infrastructure/cache/department_cache.go` |

**Target:** one Dragonfly instance per product in the suite profile (port 6381 for HRIS, 6379 in the standalone profile). Keys and their purposes:

| Key (target) | Content | TTL | Module |
|---|---|---|---|
| `sess:<sha256(session_id)>` | session metadata (user_id, tenant_id, role, expires_at); hash only, no raw session value | `tbd` (Q-HRS-11) | 01 AUTH |
| `rl:login:<ip>` and `rl:login:<acct>` | login rate-limit counters | `tbd` | 01 AUTH, SEC |
| `lock:<user_id>` | account lockout after repeated failures | `tbd` | 01 AUTH |
| `rl:api:<tenant>:<route-class>` | rate limit of sensitive endpoints | `tbd` | SEC |
| `ai:quota:<tenant>:<yyyymm>` | AI usage quota per tenant | `tbd` (Q-HRS-09) | 08 AI |
| `ent:snap:<org>` | signed entitlement snapshot (suite mode only) | 24 hours default D4, `tbd` | 02 EMP |

### 2.5 Events and outbox

**Current (NATS JetStream, ADR-0003):** stream `HRIS_EVENTS` on subject `hris.>` (created in `services/audit-service/cmd/server/main.go`). Base envelope: `event_id`, `event_type`, `schema_version` (=1), `tenant_id`, `actor_id`, `occurred_at`, `payload`.

| Subject / event type (as-is) | Producer | Existing consumers | Status | Evidence |
|---|---|---|---|---|
| `hris.identity.user.registered` | auth | employee (creates employee shell), audit, notification | Partial: the payload does not carry `tenant_id`, so the employee consumer cannot create the record; the error is acked then lost | `services/auth-service/internal/infrastructure/nats_publisher.go`, `services/employee-service/internal/infrastructure/nats_consumer.go` |
| `hris.identity.user.login_succeeded`, `login_failed` | auth | audit (unknown type: warning then ack) | Partial | `services/audit-service/internal/infrastructure/nats/identity_consumer.go` |
| `hris.identity.user.password_changed` | (declared, never emitted) | - | Not started | `services/auth-service/internal/domain/user.go` |
| `hris.identity.tenant.created` | auth | - | Partial | `services/auth-service/internal/application/commands/register_tenant.go` |
| `hris.workforce.employee.created` | employee | leave (creates balances), notification, audit | Partial | `services/leave-service/internal/infrastructure/nats/employee_created_consumer.go` |
| `hris.workforce.employee.terminated` | employee | notification, audit | Partial | `services/employee-service/internal/infrastructure/nats_publisher.go` |
| `hris.workforce.employee.updated` | (declared, never emitted because `UpdateEmployee` does not exist yet) | - | Not started | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| `hris.attendance.record.checked_in\|checked_out\|overridden` | attendance | none: audit only subscribes to `hris.operations.>` and maps `hris.operations.attendance.marked\|corrected` | Partial: subject does not match the consumer | `services/attendance-service/internal/domain/events.go`, `services/audit-service/internal/infrastructure/nats/operations_consumer.go` |
| `hris.operations.leave.requested\|approved\|rejected\|cancelled` | leave | notification (`leave_event_consumer`), audit | Partial: payload `{}` because private fields have no `MarshalJSON` | `services/leave-service/internal/infrastructure/nats/event_publisher.go`, `services/notification-service/internal/application/consumers/leave_event_consumer.go` |
| `hris.operations.leave.balance_updated` | (declared, never published) | - | Not started | `services/leave-service/internal/domain/events.go` |
| `hris.notification.*` | (no producer) | audit (consumer exists) | Not started | `services/audit-service/internal/infrastructure/nats/notification_consumer.go` |

Cross-event defects readable from the code: (a) attendance and leave payloads are empty so consumers receive no business data; (b) the attendance subject does not match what the audit consumer expects; (c) `CLAUDE.md` section 6 requires `event_id` as UUID v7 while the id generator is in each publisher and its type has not been verified (`Pending Verification`); there is no outbox; (d) consumers record idempotency in a `processed_events` table per schema (a pattern that is kept).

**Target (D-HRS-05):** no NATS. Per product: in-process domain events + Postgres transactional outbox (`FOR UPDATE SKIP LOCKED`) + idempotent consumers (`processed_events`, present in all as-is schemas). Cross-product: outbox -> platform `IntegrationService/Publish` -> platform delivers to the target product's `Deliver` endpoint; opt-in per org via entitlement feature `integration.<a>_<b>`; at-least-once delivery.

Target envelope: `{event_id UUIDv7, event_type, org_id, store_id?, actor_id, occurred_at, schema_version, payload}`. For HRIS `store_id` is not used; `org_id` = HRIS `tenant_id` in standalone mode.

Target event catalogue (final names follow a single `hrs.<agregat>.<kejadian>` scheme; naming is `tbd` until an owner decision):

| Target event | Producer | In-product consumers | Cross-product consumers (S5 hypothesis) |
|---|---|---|---|
| `hrs.employee.created` | 02 EMP | 04 LEV (balances), 05 NTF, 06 AUD | INT-1 (POS) |
| `hrs.employee.updated` | 02 EMP | 06 AUD | - |
| `hrs.employee.terminated` | 02 EMP | 04 LEV, 05 NTF, 06 AUD | INT-1: deactivate POS user |
| `hrs.attendance.checked_in` / `checked_out` / `overridden` | 03 ATT | 06 AUD | - |
| `hrs.leave.requested` / `approved` / `rejected` / `cancelled` | 04 LEV | 03 ATT (ON_LEAVE status), 05 NTF, 06 AUD | - |
| `hrs.identity.tenant_registered` / `user_registered` | 01 AUTH | 02 EMP (employee shell), 05 NTF, 06 AUD | - |
| `hrs.gathering.scheduled` (hypothesis) | 02 EMP | - | INT-2 (SELA) |

INT-1 and INT-2 are hypotheses (BR-HRS-21); validate market demand first.

### 2.6 Code structure and contracts

Target `HRIS-Stery` repo structure (target paths, not yet existing; rows of this table are not `Evidence`):

| Target path | Content |
|---|---|
| `apps/api/cmd/api` | server entry point (REST + gRPC + worker) |
| `apps/api/cmd/migrate` | goose `up\|down` |
| `apps/api/services/<name>` | modules `auth`, `employee`, `attendance`, `leave`, `notification`, `audit`, `document`, `ai` (domain, application, pgx repository, HTTP adapter, gRPC adapter) |
| `apps/api/internal/db/migrations` | embedded goose migrations (replace `services/*/migrations`) |
| `api/proto/v1`, `gen/go` | committed proto contracts; `scripts/proto-gen.sh --check` |
| `apps/web` | SvelteKit SPA |
| `kit/` (repo `stery-suite`) | `httpx`, `config`, `db`, `testdb`, `audit`, `piicrypt`, `objstore`, `health`, `tenantdb`, `session`, `outbox` (D-HRS-01) |
| `deploy/`, `scripts/`, `docs/` | `docker-compose.yml`, `scripts/check.sh`, documents |

Contracts maintained across alignment: (1) tenant-scoped table and column names remain with `tenant_id` so data can be migrated; (2) the `hris.*.v1` proto contracts are kept as the basis of `api/proto/v1`, with fixes to the problematic RPCs (section 6); (3) the RLS invariant (D8) is never relaxed during migration.

### 2.7 As-is -> to-be alignment delta

Reuse column: domain and application are kept, infrastructure and interfaces are rewritten. Effort estimate S (days), M (weeks), L (several weeks); all `[INFERENCE]`.

| as-is -> to-be | modules affected | reuse (domain/application kept, infrastructure/interfaces rewritten) | effort S/M/L [INFERENCE] |
|---|---|---|---|
| Laravel 11 gateway removed -> Go HTTP adapter `/api/v1` per module (D-HRS-12) | all modules; existing leave and auth routes in `apps/api-gateway` | Proto contracts and domain validation kept; PHP controllers and PHP gRPC client dropped; route map from proto annotations (section 6) | L |
| SvelteKit SSR (adapter-auto, cookie set on the server) -> SPA `ssr=false` + `adapter-static`, `apiFetch`, i18n `{id,en}` (D-HRS-07) | 4 web pages | UI components `src/lib/components/ui/*`, leave zod schemas kept; `+page.server.ts` and `hooks.server.ts` rewritten | M |
| JWT RS256 (15 minutes/7 days, hardcoded) -> opaque HttpOnly session, hash in Dragonfly (D3) | 01 AUTH, all interceptors, web | `User` entity, password rules, bcrypt, domain RBAC kept; JWT signer and claims dropped | M |
| NATS JetStream -> Postgres outbox + idempotent consumers (D-HRS-05) | 01, 02, 03, 04, 05, 06 | Domain event definitions and `processed_events` kept; `nats_*` rewritten; event payloads completed | M |
| golang-migrate -> embedded goose v3.26.0, `cmd/migrate`, CI up->down->up | all schemas | DDL content kept after fixing (attendance 001 invalid); file format changed | M |
| sqlc (auth-service only) + pgx -> raw pgx SQL | 01 AUTH | Repository interfaces in domain kept; sqlc code dropped | S |
| zap -> `log/slog` JSON | all | Log calls mapped; `services/_shared/observability/logging.go` dropped | S |
| MinIO -> RustFS via `objstore.Store` (presign-only, ClamAV fail-closed) | 07 DOC | Document business requirements kept; `storage/minio_client.go`, `presigner.go` rewritten | S |
| Redis 7 -> Dragonfly (go-redis/v9) | 01 AUTH, SEC | Token store interface kept | S |
| Microservices (7 processes, 7 gRPC ports) -> modules `services/<name>` in a single `apps/api` process | all | All domain/application; per-service `main.go` replaced by a single `cmd/api`; inter-module gRPC replaced by Go interface calls | L |
| Per-service RLS via `postgres/rls.go` -> RLS ENABLE+FORCE+`WITH CHECK` on all tenant tables via `kit/tenantdb.WithTenant`; NOBYPASSRLS role (D8) | all schemas | Existing RLS policies fixed (FORCE, `WITH CHECK`); `rls.go` replaced | M |
| OTel/Prometheus (`_shared/observability`, never initialised) -> `slog` + `/healthz` baseline; OTel optional (D11, Q-HRS-08) | all | Nothing kept; OTel code is not active today | S |
| Dev credentials written in tracked files -> `.env.example` + fail-fast validation | deploy | `deploy/docker/*` rewritten without secrets | S |
| Plain `TEXT` NIK/NPWP/birth_date columns -> `kit/piicrypt` (AES-256-GCM + blind index) | 02 EMP | Column schema kept, blind index column added; old data encrypted during migration | M |
| Leave/attendance without `employee_number`, UPPERCASE vs lowercase enums, missing `description` column -> schema and SQL fixes | 02, 03, 04 | Domain kept; repository SQL rewritten | M |
| Documents: empty presign/rollback/download RPCs -> real implementation + tests | 07 DOC | `File`/`FileVersion` domain kept | M |
| `suite` mode: port adapters `Identity`, `Tenant`, `Entitlement`, `Notification`, `Audit`, `Events` + product manifest | 01, 05, 06, SEC | Interfaces placed in application; new adapters | M |

---

## 3. Actors and Access Rights Model

### 3.1 Actors

| Actor | Description | As-is status | Evidence |
|---|---|---|---|
| HR Admin (`hr_admin`) | manages master data, views audit, approves leave | Partial: role exists, enforcement does not | `services/auth-service/internal/domain/permission.go` |
| Manager (`manager`) | views subordinate employees and attendance, approves leave | Partial: without an enforced supervisor-subordinate hierarchy | `services/auth-service/internal/domain/permission.go` |
| Employee (`employee`) | applies for leave, views own data | Partial | `services/auth-service/internal/domain/permission.go` |
| Auditor | reads the audit trail | Not started: no built-in role; only `audit:read` owned by `hr_admin` | `services/auth-service/internal/domain/permission.go` |
| Platform admin (suite) | manages orgs/tenants, entitlements | Not started (lives in the `stery-suite` platform) | `proto/hris/auth/v1/auth.proto` |
| System/worker | event consumers, scheduled jobs | Partial | `services/leave-service/internal/infrastructure/nats/employee_created_consumer.go` |

### 3.2 Built-in permissions and role matrix (as-is)

Eleven built-in permissions are created by `RegisterTenant`: `employee:read|write|delete`, `attendance:read|write`, `leave:read|write|approve`, `notification:read`, `audit:read`, `settings:write`. Three system roles are created per tenant.

| Permission | `hr_admin` | `manager` | `employee` |
|---|---|---|---|
| `employee:read` | yes | yes | yes |
| `employee:write` | yes | - | - |
| `employee:delete` | yes | - | - |
| `attendance:read` | yes | yes | - |
| `attendance:write` | yes | - | - |
| `leave:read` | yes | yes | yes |
| `leave:write` | - | - | yes |
| `leave:approve` | yes | yes | - |
| `notification:read` | yes | yes | yes |
| `audit:read` | yes | - | - |
| `settings:write` | yes | - | - |

Findings (Evidence `services/auth-service/internal/domain/permission.go`, lines 33-58):
- `hr_admin` does not have `leave:write` (cannot apply for leave on its own behalf if that is its only role).
- `employee` does not have `attendance:*`: with the built-in RBAC an employee cannot check in. This is a design defect that must be corrected (FR-HRS-03.1, FR-HRS-01.6).
- There are no permissions for documents and AI (`document:*`, `ai:*`); their addition is in FR-HRS-07.* and FR-HRS-08.*.
- Leave approval in leave-service reads the `x-user-roles` header sent by the caller (`services/leave-service/internal/interfaces/grpc/context_guards.go`, lines 92-140); that header can be forged because there is no authentication interceptor.

### 3.3 Target access rights model

- Tenant and actor always come from the session (opaque cookie), never from the request body (FR-HRS-SEC.4, SEC.5).
- RBAC tables (roles, permissions, actions, role_actions) + `RequirePermission(code)` (STACK-SPEC 9). Target permissions add `document:read|write|delete`, `ai:use`, `user:manage`.
- Data scope: `employee` only their own row (`employee_id` from the session); `manager` the direct team by `manager_id`; `hr_admin` the whole tenant. Per-module rules are in section 4.
- Deny-by-default: routes/RPCs without a declared permission are rejected (FR-HRS-SEC.5).

---

## 4. Functional Specification per Module

Each module states: a summary of the current state, the FR table (`ID | Requirement | Prioritas | Status | Bukti`), then business rules, state machine (if any), and events produced/consumed. The Requirement column states the **target** behaviour; the Status and Evidence columns state the **current** condition.

### 4.01 AUTH: identity, tenant registration, RBAC

The `auth-service` service has six RPCs and all six have real handlers (Evidence `services/auth-service/internal/interfaces/grpc/auth_service.go`). Schema `auth`: `tenants`, `users`, `roles`, `permissions`, `role_permissions`, `user_roles`, `login_attempts`, `refresh_tokens` (the last two are unused by code), `processed_events`. Passwords are hashed with bcrypt default cost (10).

| ID | Requirement | Priority | Status | Evidence |
|---|---|---|---|---|
| FR-HRS-01.1 | A tenant registers itself (name, unique slug, first admin). In one atomic transaction the tenant, three system roles, built-in permissions, and an admin user with role `hr_admin` are created; events `tenant_registered` and `user_registered` are emitted via the outbox | M | Partial: handler and rules exist, but creation is not atomic and INSERT/UPDATE on `auth.tenants` is only granted to role `hris_admin`, so it fails under the default DSN `hris_app` (`Pending Verification`, not yet run) | `services/auth-service/internal/application/commands/register_tenant.go` |
| FR-HRS-01.2 | Login with email + password; success issues an opaque session (HttpOnly Secure cookie) containing tenant, user, roles; wrong credentials do not distinguish "email does not exist" from "wrong password" | M | Partial: the `Login` handler exists and issues a JWT; the web uses a mock login to `localhost:3000` and `setSession` is never called | `services/auth-service/internal/application/commands/login.go`, `apps/web/src/routes/(auth)/login/+page.server.ts` |
| FR-HRS-01.3 | Password policy: 8-128 characters with upper case, lower case, digit, symbol; bcrypt hash (hash scheme kept, D-HRS-06 and D10); tenant slug 3-50 characters `^[a-z0-9-]+$` | M | Done | `services/auth-service/internal/domain/user.go` |
| FR-HRS-01.4 | Session lifecycle: issue, validate (introspection), extend, revoke/logout; expiry with TTL `tbd` (Q-HRS-11); only the session hash is stored in Dragonfly. As-is: JWT RS256 15 minutes + refresh 7 days (hardcoded), refresh and revocation list in Redis | M | Partial: refresh and revocation exist but without token-family reuse detection; TTL hardcoded in `cmd/server/main.go` line 81 | `services/auth-service/internal/application/commands/refresh_token.go`, `services/auth-service/internal/application/commands/revoke_token.go`, `services/auth-service/internal/infrastructure/jwt_signer.go` |
| FR-HRS-01.5 | Account lockout after repeated failed attempts (threshold and duration `tbd`) and login rate limiting per address and per account | M | Not started: table `login_attempts` exists in the schema but is unused; no lockout code | `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| FR-HRS-01.6 | RBAC: roles, permissions, actions; `RequirePermission(code)` on every route; built-in permissions corrected (employee gains `attendance:write` and their own `attendance:read`; `hr_admin` gains `leave:write`; document and AI permissions added) | M | Partial: 11 permissions and 3 roles are created, but nothing enforces them; matrix error in 3.2 | `services/auth-service/internal/domain/permission.go` |
| FR-HRS-01.7 | User management by admin: create, deactivate, change role, reset access; user linked to employee (`employee_id`) | S | Not started: the auth proto has no user RPCs | `proto/hris/auth/v1/auth.proto` |
| FR-HRS-01.8 | Email verification and forgot/reset password via a short-lived single-use link | S | Not started: no RPC, and no `password_changed` event is emitted | `services/auth-service/internal/domain/user.go` |
| FR-HRS-01.9 | Client IP address and user agent recorded in login/audit events | S | Partial: `extractIPAddr` always returns `0.0.0.0` (lines 196-199) | `services/auth-service/internal/interfaces/grpc/auth_service.go` |
| FR-HRS-01.10 | `suite` mode: identity is delegated to the platform via the `Identity` port (the `suite` adapter calls `SessionService/Introspect`); `standalone` mode is unchanged (D-HRS-02, D-HRS-03) | M | Not started | `proto/hris/auth/v1/auth.proto` |

**Business rules**
- One unique email per tenant (target); tenant slug globally unique.
- A `deactivated` user cannot log in and all their sessions are revoked.
- System roles cannot be deleted; custom roles `tbd` (Q-HRS-06 covers the policy, no decision yet).
- Tenant and actor on operations are never taken from the request body (as-is: taken from the body; see 8.2).

**Session state machine (target)**

| From | Event | To |
|---|---|---|
| (none) | successful login | `active` |
| `active` | extension (sliding, TTL `tbd`) | `active` |
| `active` | logout / admin revocation / user deactivated | `revoked` |
| `active` | TTL elapsed | `expired` |

**Events.** Produced: `identity.tenant_registered`, `identity.user_registered`, `user.login_succeeded`, `user.login_failed` (as-is `hris.identity.*`, 2.5). Consumed: none.

### 4.02 EMP: employees, departments, positions, org chart

The `employee-service` service has 10 RPCs; 8 have handlers, `UpdateEmployee` and `GetOrgChart` are absent from `internal/interfaces/grpc/employee_service.go`. Schema `employee`: `departments`, `positions`, `employees`, `processed_events`; lowercase PG enums (`employment_status`, `contract_type`) while the Go domain writes upper case. Sensitive columns: `national_id` (NIK), `tax_id` (NPWP), `birth_date` (lines 87-91 of the migration) are plain-typed.

| ID | Requirement | Priority | Status | Evidence |
|---|---|---|---|---|
| FR-HRS-02.1 | Create an employee: automatic or manual `employee_number` unique per tenant, personal data, department, position, `manager_id`, contract type, join date; reject beyond `hrs.employees.max` (FR-HRS-02.11) | M | Partial: handler exists, but the INSERT does not fill `employee_number` (NOT NULL), the domain has no `employeeNumber`, `ContractType` is ignored, and the upper-case domain enum is written to a lowercase PG enum (`Pending Verification`, not yet run) | `services/employee-service/internal/application/commands/create_employee.go`, `services/employee-service/internal/infrastructure/postgres/employee_repository.go` |
| FR-HRS-02.2 | Employee list and detail with filters (status, department, name search, number), pagination with a correct total | M | Partial: `ListEmployees` returns Total = page length and offset is always 0; `GetEmployee` exists | `services/employee-service/internal/application/queries/list_employees.go`, `services/employee-service/internal/application/queries/get_employee.go` |
| FR-HRS-02.3 | Update employee data (personal data, department, position, manager, status) with an audit trail and an `employee.updated` event | M | Not started: RPC `UpdateEmployee` (PATCH `/v1/employees/{id}`) has no handler | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| FR-HRS-02.4 | Employee termination (date, reason) closes the status, deactivates the linked user, triggers closing of balances/attendance, and emits `employee.terminated` | M | Partial: termination and event exist; no cascade to user/leave/attendance | `services/employee-service/internal/application/commands/terminate_employee.go` |
| FR-HRS-02.5 | Hierarchical departments (`parent_id`), department head, list | M | Partial: create and list exist; `parent_id` is unused; in-process cache with 24-hour TTL | `services/employee-service/internal/application/commands/create_department.go`, `services/employee-service/internal/infrastructure/cache/department_cache.go` |
| FR-HRS-02.6 | Positions with level (JUNIOR/SENIOR/LEAD/MANAGER, aligned with the domain), list | M | Partial: INSERT/SELECT use a `description` column that does not exist in the schema; the default of column `level` `'staff'` is not in the domain enum (`Pending Verification`) | `services/employee-service/internal/infrastructure/postgres/position_repository.go`, `services/employee-service/internal/domain/position.go` |
| FR-HRS-02.7 | Org chart (department tree and manager-subordinate tree) for the tenant, including department filter | M | Not started: RPC `GetOrgChart` (GET `/v1/employees/org-chart`) has no handler | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| FR-HRS-02.8 | NIK, NPWP, birth date encrypted with `kit/piicrypt` (AES-256-GCM, blind index for NIK/number search); display masked for roles that do not need it; access recorded in audit | M | Not started: plain `TEXT` columns (lines 89-91) | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| FR-HRS-02.9 | Bulk employee import (CSV) with error preview and quota limit | C | Not started | `proto/hris/employee/v1/employee.proto` |
| FR-HRS-02.10 | The `user_registered` event creates an employee shell for the tenant's first admin | S | Partial: consumer exists but cannot work (payload without `tenant_id`, random department/position ids, error acked then lost) | `services/employee-service/internal/infrastructure/nats_consumer.go` |
| FR-HRS-02.11 | Limit `hrs.employees.max` per entitlement (suite mode) or tenant configuration (standalone); value `tbd` (Q-HRS-02); exceeding it only rejects creation of new employees, existing data stays | M | Not started | `services/employee-service/internal/application/commands/create_employee.go` |

**Business rules**
- `employee_number` is unique per tenant (`UNIQUE (tenant_id, employee_number)` already in the schema).
- A terminated employee is not physically deleted; status `terminated` with `termination_date`.
- `manager_id` refers to another employee in the same tenant (self FK, `ON DELETE SET NULL`); manager cycles are rejected (target).
- Deletion of personal data follows the UU PDP retention policy (8.3).

**Employee state machine** (PG enum `employment_status`: `active`, `inactive`, `terminated`, `on_leave`; the Go domain uses upper-case values, `services/employee-service/internal/domain/employee.go`):

| From | Event | To |
|---|---|---|
| (new) | create | `active` |
| `active` | deactivate | `inactive` |
| `inactive` | activate | `active` |
| `active` | long leave `tbd` (policy Q-HRS-06) | `on_leave` |
| `active`, `inactive`, `on_leave` | terminate | `terminated` (final) |

**Events.** Produced: `employee.created`, `employee.terminated`, `employee.updated` (not yet emitted). Consumed: `identity.user_registered` (FR-HRS-02.10).

### 4.03 ATT: attendance

The `attendance-service` service has five RPCs with handlers. Schema `attendance` has one table `attendance_records` (columns: `employee_id`, `date`, `check_in_at`, `check_out_at`, coordinates, `status`, `work_duration_mins`, `notes`, `created_by`, `is_override`). **Migration 001 is invalid in PostgreSQL** (line 25 `CONSTRAINT ... UNIQUE (...) WHERE NOT is_override` is not valid syntax; lines 45-46 `GRANT ... TO app` use role `app`, not `hris_app`); as a result the service cannot run on a fresh DB until the migration is fixed.

| ID | Requirement | Priority | Status | Evidence |
|---|---|---|---|---|
| FR-HRS-03.1 | Daily check-in by the employee: one active record per employee per day; stores server time (UTC) and optional coordinates; rejected if already checked in | M | Partial: command exists; depends on the invalid migration 001; built-in employee RBAC has no `attendance:*` | `services/attendance-service/internal/application/commands/check_in.go`, `services/attendance-service/migrations/001_create_attendance_schema.up.sql` |
| FR-HRS-03.2 | Check-out: requires an active check-in; computes `work_duration_mins`; out time after in time (constraint `attendance_check_in_before_out`) | M | Partial | `services/attendance-service/internal/application/commands/check_out.go` |
| FR-HRS-03.3 | Correction by HR/supervisor with mandatory reason and audit trail (`is_override`, `created_by`); may create a new record if none exists | M | Partial: override fails for new records | `services/attendance-service/internal/application/commands/override_attendance.go` |
| FR-HRS-03.4 | Attendance list with filters by date range, employee, status; pagination | M | Partial: `ListAttendance` ignores the date filter | `services/attendance-service/internal/interfaces/grpc/attendance_service.go` |
| FR-HRS-03.5 | Attendance summary per employee per period (present, late, absent, leave, total hours) and daily summary for HR | M | Partial: `GetAttendanceSummary` ignores `employee_id` and `period`; the daily summary exists | `services/attendance-service/internal/application/queries/get_daily_summary.go` |
| FR-HRS-03.6 | Work rules per tenant: working hours, lateness threshold, working days, optional geofence; `LATE`/`HALF_DAY` status computed automatically; default values `tbd` (Q-HRS-05) | S | Not started: statuses `LATE` and `HALF_DAY` exist in the enum but are never computed | `services/attendance-service/internal/domain/attendance_record.go` |
| FR-HRS-03.7 | Approved leave marks the related days `ON_LEAVE` (consumer of the `leave.approved` event) | S | Not started | `services/attendance-service/internal/domain/attendance_record.go` |
| FR-HRS-03.8 | Dates and day boundaries are computed in the tenant time zone (default WIB); times stored in UTC | M | Not started: column `date` has no zone; tenant zone does not exist | `services/attendance-service/migrations/001_create_attendance_schema.up.sql` |

**Business rules**
- One non-override record per `(tenant_id, employee_id, date)` is unique; in PostgreSQL this is expressed as a partial unique index (`CREATE UNIQUE INDEX ... WHERE NOT is_override`), not a `CONSTRAINT`.
- Check-out must not be before check-in. A correction does not overwrite original data without an audit trail.
- Migration 002 creates index `idx_attendance_employee_date` without `IF NOT EXISTS`, the same name as the index in migration 001 (line 29): migration 002 is expected to fail with a "relation already exists" error even after 001 is fixed (`Pending Verification`, not yet run). Both must be fixed together.

**Daily attendance state machine**

| From | Event | To |
|---|---|---|
| (none) | check-in | `PRESENT` (or `LATE` if past the threshold, target) |
| `PRESENT` | check-out with duration below the full-day limit | `HALF_DAY` (target) |
| (none) | leave approved | `ON_LEAVE` (target) |
| (none) | end of day without check-in | `ABSENT` (day-closing job, target) |

**Events.** Produced: `attendance.checked_in`, `checked_out`, `overridden` (as-is subject `hris.attendance.record.*`, does not match the audit consumer; empty payload). Consumed: `leave.approved` (target, FR-HRS-03.7).

### 4.04 LEV: leave

The `leave-service` service has eight RPCs with handlers. Schema `leave`: `leave_types` (code, name, `max_days_per_year`, `requires_document`, `is_paid`), `leave_requests` (status `PENDING|APPROVED|REJECTED|CANCELLED`, `days_count INTEGER CHECK (> 0)`, `approved_by_id`, `rejection_reason`), `leave_balances` (`entitled_days`, `used_days`, `pending_days`, unique per year), `processed_events`. RLS policies use `current_setting('app.tenant_id')` without `missing_ok`, without `WITH CHECK` and without `FORCE`; the migration grants rights to role `app`.

| ID | Requirement | Priority | Status | Evidence |
|---|---|---|---|---|
| FR-HRS-04.1 | Leave types per tenant can be created/changed/deactivated by HR; built-in seed per policy `tbd` (Q-HRS-06) | M | Partial: `ListLeaveTypes` exists; there is no leave-type creator and no seed, so the list is empty and no balances exist | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| FR-HRS-04.2 | Annual balance per employee per type is created on `employee.created` (idempotent) and can be adjusted by HR | M | Partial: consumer exists but the `EmployeeValidator` in `main.go` is a no-op and without leave types no balance is formed | `services/leave-service/internal/infrastructure/nats/employee_created_consumer.go` |
| FR-HRS-04.3 | Apply for leave: type, date range, reason; the system computes `days_count` of working days; rejects insufficient balance, dates overlapping an active request, and reversed ranges | M | Partial: balance is enforced in application code (without row lock); overlap is not checked; `ApplyLeave` does not fill `DaysCount`, so the INSERT may violate `CHECK (days_count > 0)` (`Pending Verification`) | `services/leave-service/internal/application/commands/create_leave_request.go`, `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| FR-HRS-04.4 | Approve or reject by direct supervisor or HR with a rejection reason; the approval actor is taken from the session; the requester must not approve their own request; decision only from `PENDING` | M | Partial: handler exists; `approver_id` is taken from the request body, no self-approval prevention, roles read from the `x-user-roles` header | `services/leave-service/internal/interfaces/grpc/context_guards.go`, `services/leave-service/internal/domain/leave_request.go` |
| FR-HRS-04.5 | Cancellation by the requester (and HR) restores `pending_days`/`used_days`; the rule on which statuses may be cancelled is set in the domain | M | Partial: `Cancel` allows `PENDING` although `ErrCannotCancelPending` is declared (`leave_request.go` lines 159-172); balance restoration not verified | `services/leave-service/internal/domain/leave_request.go`, `services/leave-service/internal/domain/errors.go` |
| FR-HRS-04.6 | Request list and detail with filters by status/employee/period; an employee sees only their own, a manager their team, HR all | M | Partial: status filter is sent as `LEAVE_STATUS_*` which does not match DB values (`PENDING|...`) | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| FR-HRS-04.7 | View balance per type for the current year | M | Partial: `GetLeaveBalance` exists; requests spanning years are counted entirely in the start year | `services/leave-service/internal/domain/leave_balance.go` |
| FR-HRS-04.8 | National/tenant holiday calendar excluded from the working-day count | C | Not started | `services/leave-service/migrations/001_create_leave_schema.up.sql` |
| FR-HRS-04.9 | Manual balance adjustment by HR (add/subtract) with reason and audit; annual carry-over per policy `tbd` | C | Not started | `services/leave-service/internal/domain/leave_balance.go` |

**Business rules**
- Available balance = `entitled_days - used_days - pending_days` and must not be negative (DB CHECK already exists on the three columns).
- A `PENDING` request adds to `pending_days`; `APPROVED` moves it to `used_days`; `REJECTED`/`CANCELLED` restores it.
- A request ending before it starts is rejected; types with `requires_document` need a document attachment (target, related to 4.07).

**Leave state machine**

| From | Event | To | Balance effect |
|---|---|---|---|
| (new) | apply | `PENDING` | `pending_days += days_count` |
| `PENDING` | approve | `APPROVED` | `pending_days -= n; used_days += n` |
| `PENDING` | reject | `REJECTED` | `pending_days -= n` |
| `PENDING` | cancel | `CANCELLED` | `pending_days -= n` |
| `APPROVED` | cancel (before the start date, target) | `CANCELLED` | `used_days -= n` |

**Events.** Produced: `leave.requested`, `approved`, `rejected`, `cancelled` (as-is payload empty). Consumed: `employee.created` (creates balances), `employee.terminated` (target: closes balances).

### 4.05 NTF: notifications

The `notification-service` service has five RPCs with handlers. Schema `notification`: `notifications` (channel `EMAIL|IN_APP`, `template_key`, `variables` JSONB, status `PENDING|SENT|FAILED|READ`), `notification_channel_configs` (per tenant: `channels TEXT[]`, `email_config` JSONB), `processed_events`. The email and in-app adapters exist in code but are **not wired**: `cmd/server/main.go` lines 112-132 assign `_ = emailAdapter` and `_ = inAppAdapter`; six HTML templates exist in `templates/` but are never loaded. The migration grants rights to role `app`; RLS ENABLE without FORCE, `WITH CHECK` present via a separate policy.

| ID | Requirement | Priority | Status | Evidence |
|---|---|---|---|---|
| FR-HRS-05.1 | Notifications stored per recipient (`recipient_id`) with channel, template, variables, status, and send/read time; created in the same transaction as the triggering event (outbox) | M | Partial: table and repository exist; storage runs via consumers/handlers | `services/notification-service/migrations/001_create_notification_schema.up.sql` |
| FR-HRS-05.2 | Email delivery via SMTP with `{id,en}` localised templates, limited retries, and recording of `SENT`/`FAILED` | M | Partial: the email adapter exists but is not wired; templates are not loaded; `MarkSent`/`MarkFailed` have no callers so status never leaves `PENDING` | `services/notification-service/internal/infrastructure/channels/email_adapter.go`, `services/notification-service/internal/infrastructure/template_engine.go` |
| FR-HRS-05.3 | Event consumption: leave request (to supervisor/HR), leave decision (to requester), new employee, registered user; idempotent consumption | M | Partial: leave (with tests), auth, and employee consumers exist; the leave payload is empty so templates cannot be filled | `services/notification-service/internal/application/consumers/leave_event_consumer.go` |
| FR-HRS-05.4 | User notification list (pagination, unread count) and mark-as-read (one/all) | M | Partial: `MarkRead` only accepts status `SENT`, so it always fails silently; `unread_count` is always 0; `next_page_token` contains the literal `"next"` | `services/notification-service/internal/application/commands/mark_notification_read.go`, `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| FR-HRS-05.5 | Channel configuration per tenant (active channels, SMTP); SMTP secret stored encrypted, not in plain JSONB | S | Partial: `UpdateChannelConfig` overwrites the channel list; `smtp_password` is stored plain in JSONB | `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| FR-HRS-05.6 | Additional channels (WhatsApp, push) | C | Not started (provider `tbd`, Q-HRS-13) | `proto/hris/notification/v1/notification.proto` |
| FR-HRS-05.7 | Preferences and opt-out per user per notification type; mandatory security notifications cannot be turned off | C | Not started | `proto/hris/notification/v1/notification.proto` |

**Business rules**
- A delivery failure does not roll back the business transaction; the notification is queued and retried up to a limit `tbd`.
- Notification content does not contain class 2 PII (NIK, NPWP, birth date).
- Recipients are determined from employee/supervisor data; in suite mode delivery goes through the `Notification` port (platform hub).

**Notification state machine**

| From | Event | To |
|---|---|---|
| (new) | created | `PENDING` |
| `PENDING` | delivered | `SENT` |
| `PENDING` | permanent failure | `FAILED` |
| `SENT` or (IN_APP) `PENDING` | read | `READ` |

**Events.** Consumed: `identity.user_registered`, `employee.created|terminated`, `leave.requested|approved|rejected|cancelled`. Produced: `notification.*` (no producer yet; audit has a consumer).

### 4.06 AUD: audit

The `audit-service` service has three RPCs with handlers. Schema `audit`: `entries` (partitioned monthly 2023-01 to 2027-12, without a default partition; columns `actor_id`, `action`, `resource_type`, `resource_id UUID`, `changes JSONB`, `success`), `processed_events`. RLS is not enabled on `entries` (migration comment: the audit service reads all tenants); role `hris_app` is only granted INSERT.

| ID | Requirement | Priority | Status | Evidence |
|---|---|---|---|---|
| FR-HRS-06.1 | Record audit entries for sensitive actions from other modules synchronously (port `Audit.Record`) or via events | M | Partial: identity/workforce/operations/notification consumers exist; `resource_id` is of type UUID while consumers send an empty string, so the INSERT is expected to fail (`Pending Verification`) | `services/audit-service/internal/infrastructure/nats/identity_consumer.go`, `services/audit-service/internal/infrastructure/postgres/audit_repository.go` |
| FR-HRS-06.2 | Query entries with filters by actor, action, resource type/ID, time range; pagination; only role `audit:read`; always limited to the caller's tenant | M | Pending Verification: `hris_app` only has INSERT on `entries`, so `Query` is expected to be permission-denied (`[INFERENCE]`) | `services/audit-service/internal/interfaces/grpc/audit_service.go`, `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| FR-HRS-06.3 | Fetch a single entry by ID | S | Partial: `GetEntry` uses a query with `Limit:1` so it does not return the requested entry | `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| FR-HRS-06.4 | Entries cannot be changed or deleted (UPDATE/DELETE/TRUNCATE rejected at DB level, `kit/audit` trigger) | M | Partial: append-only only via rights restriction (`hris_app` INSERT); no rejecting trigger | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| FR-HRS-06.5 | Entry retention per policy (period `tbd`, Q-HRS-04) with automatic partition maintenance | S | Partial: static partitions up to 2027-12; entries outside the range will fail | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| FR-HRS-06.6 | Minimum coverage: login (success/failure), employee changes, attendance corrections, leave decisions, access/download of confidential documents, PII changes, role changes | M | Partial: the attendance subject does not match the consumer; `login_*` types are not mapped; document and PII access are not recorded | `services/audit-service/internal/infrastructure/nats/operations_consumer.go` |
| FR-HRS-06.7 | Tenant isolation on audit entries: RLS ENABLE+FORCE with `tenant_id`; cross-tenant access only for the platform operator role | M | Not started: RLS is disabled on `entries` | `services/audit-service/migrations/001_create_audit_schema.up.sql` |

**Business rules**
- An entry contains who (`actor_id`), what (`action`, `resource_type`, `resource_id`), when (`created_at`), result (`success`, `error_message`), and before/after changes (`changes`) without class 2 PII values (only a "changed" marker).
- Writing audit must not roll back a sensitive business transaction that already succeeded; write failures are recorded and retried (outbox).

**Events.** Consumed: all `identity.*`, `workforce.*` (employee), `operations.*` (leave/attendance), `notification.*`. Produced: none.

### 4.07 DOC: employee documents

The `document-service` service has seven RPCs; four have real handlers (`UploadDocument`, `ListDocuments`, `GetDocumentMetadata`, `DeleteDocument`), three are empty (`DownloadDocument`, `GetPresignedUrl`, `RollbackDocumentVersion`; `document_service.go` lines 215-221 for the last two). There are no test files in this service. Schema `document`: `files` (columns `minio_bucket`, `minio_key`, `status`, `classification` default `'PUBLIC'`, `retention_days`, `current_version`, `total_versions`), `file_versions` (without `tenant_id` and without RLS), `processed_events`. `UploadDocument` always returns `Unauthenticated` because the context key `actor_id` is never filled anywhere (`internal/interfaces/grpc/mapping.go` line 13 reads it).

| ID | Requirement | Priority | Status | Evidence |
|---|---|---|---|---|
| FR-HRS-07.1 | Two-step upload: the client requests a presign URL (size/MIME limits validated), uploads directly to RustFS, then confirms; metadata is created in a transaction together with audit | M | Partial: `UploadDocument` exists but always `Unauthenticated`; `GetPresignedUrl` is empty; the bucket (`<tenantID>-documents`) is never created | `services/document-service/internal/application/commands/upload_document.go`, `services/document-service/internal/interfaces/grpc/document_service.go` |
| FR-HRS-07.2 | Malware scanning (ClamAV `clamd`) fail-closed: a file with status `SCANNING` cannot be downloaded until `CLEAN`; `SCAN_FAILED` is rejected | M | Not started: status enums `SCANNED`/`SCAN_FAILED` exist without a scanner | `services/document-service/internal/domain`, `services/document-service/migrations/001_create_document_schema.up.sql` |
| FR-HRS-07.3 | Download only via short-lived presign URL (TTL `tbd`; as-is 3600 seconds hardcoded) with permission and classification checks and an audit record for confidential documents | M | Not started: `DownloadDocument` is a placeholder | `services/document-service/internal/interfaces/grpc/document_service.go` |
| FR-HRS-07.4 | Document list and metadata per entity (`entity_type`, `entity_id`), not showing deleted documents | M | Partial: list/metadata exist; deleted documents are not filtered | `services/document-service/internal/interfaces/grpc/document_service.go` |
| FR-HRS-07.5 | Delete document (soft-delete, audit) and physical deletion per retention (`retention_days`) | S | Partial: `DeleteDocument` exists; there is no retention worker | `services/document-service/internal/application/commands/delete_document.go` |
| FR-HRS-07.6 | Document versions: upload a new version and restore an old version (`RollbackDocumentVersion`) with tests | S | Not started: `CreateDocumentVersion` is not wired to any RPC; rollback is empty | `services/document-service/internal/application/commands/create_document_version.go` |
| FR-HRS-07.7 | Document classification (`PUBLIC`, `INTERNAL`, `CONFIDENTIAL`, final `tbd`) determines access rights and audit recording; employee documents (contracts, identity) default to `CONFIDENTIAL` | S | Partial: column `classification` exists with default `PUBLIC`; not enforced | `services/document-service/migrations/001_create_document_schema.up.sql` |
| FR-HRS-07.8 | Size limit, allowed MIME types, and storage quota per tenant (values `tbd`) | S | Not started: no size/MIME validation | `services/document-service/internal/application/commands/upload_document.go` |

**Business rules**
- One bucket and one access key for HRIS in RustFS (D-HRS-06); objects are prefixed `tenant_id/` and are never public.
- `CONFIDENTIAL` document actions are always audited (FR-HRS-06.6). Deleting an employee does not automatically delete documents; retention decides.

**Document state machine**

| From | Event | To |
|---|---|---|
| (new) | presign request | `PENDING_UPLOAD` (target) |
| `PENDING_UPLOAD` | upload confirmed | `SCANNING` |
| `SCANNING` | clean | `CLEAN` (downloadable) |
| `SCANNING` | infected/failed | `SCAN_FAILED` (rejected) |
| `CLEAN` | delete | `DELETED` |

As-is status values: `UPLOADED` (column default), `SCANNED`, `SCAN_FAILED`; the target names above are `[INFERENCE]` and adjusted during implementation.

**Events.** Produced: `document.uploaded|deleted` (target). Consumed: `employee.terminated` (target, triggers retention scheduling).

### 4.08 AI: AI assistant

`ai-service` only has the contract `proto/hris/ai/v1/ai.proto` (three RPCs: `AnalyzeResume`, `ForecastAttrition`, `Chat` streaming). The `cmd/server/` directory is empty and the service is commented out in `go.work`. The whole module has status **Not started**. Processing PII with a third-party LLM provider requires a policy first (BR-HRS-11, RISK-HRS-12).

| ID | Requirement | Priority | Status | Evidence |
|---|---|---|---|---|
| FR-HRS-08.1 | Resume analysis (matching candidates to a job description) | C | Not started | `proto/hris/ai/v1/ai.proto` |
| FR-HRS-08.2 | Attrition prediction based on aggregate attendance/leave data (no direct PII) | C | Not started | `proto/hris/ai/v1/ai.proto` |
| FR-HRS-08.3 | HR assistant chat (streaming) with refusal of out-of-scope topics and without sending NIK/NPWP/birth date to the LLM | C | Not started | `proto/hris/ai/v1/ai.proto` |
| FR-HRS-08.4 | Policy and quota: consent withdrawal, cost quota per tenant (`tbd`), usage log without sensitive content; feature only active if entitlement/tenant enables it | C | Not started | `go.work` |

**Business rules.** No class 2 PII is sent to third parties without a processing basis and an agreement (Q-HRS-09, Q-HRS-14). The module is off by default.

**Events.** Produced/consumed: none (not yet defined).

### 4.09 PAY: payroll (Deferred)

**Status: Deferred** (ADR-0004, `docs/adr/0004-defer-payroll-to-phase2.md`). No FR in this version. Rationale: payroll logic involves PPh21, BPJS, and THR which have legal and financial consequences; calculation errors directly affect employees and the company. Deferral to phase 2 is a documented decision (BR-HRS-13, D-HRS-10).

Prerequisites before payroll FRs can be written: (1) legal review of the applicable PPh21, BPJS, THR rules (schedule `tbd`, Q-HRS-10); (2) tax rules ADR (module 10); (3) `PayrollRun` immutability ADR (an issued run does not change); (4) saga/period-closing design; (5) retention and UU PDP review of salary data.

### 4.10 TAX: tax rules (Deferred)

**Status: Deferred** (ADR-0004). No FR. Tax rules (rates, PTKP, TER) must come from legal review and be maintained as effective-dated data, not code constants. Prerequisites: legal review (Q-HRS-10), tax rules ADR, and a regulation update strategy.

### 4.11 REC: recruitment (Deferred)

**Status: Deferred** (ADR-0004, BR-HRS-14). No FR. Prerequisites: candidate domain and application flow design, candidate data retention policy (UU PDP), and a decision related to the AI module (FR-HRS-08.1).

### 4.12 PRF: performance (Deferred)

**Status: Deferred** (ADR-0004, BR-HRS-14). No FR. Prerequisites: appraisal cycle and rating scale design, review of access rights to appraisal data, and a retention policy.

---

### 4.SEC Cross-module security requirements

The following requirements apply across modules 4.01 to 4.08; the Requirement column states the **target** behaviour, Status and Evidence state the **current** condition.

| ID | Requirement | Priority | Status | Evidence |
|---|---|---|---|---|
| FR-HRS-SEC.1 | Every tenant table uses RLS ENABLE+FORCE with `tenant_isolation` (`USING` and `WITH CHECK`) and `current_setting('app.tenant_id', true)::uuid`, set per transaction by `kit/tenantdb.WithTenant`; app role NOBYPASSRLS; separate migration role; `audit.entries` included | M | Partial: correct on auth (users, roles, refresh_tokens), employee, document.files; not yet on attendance, leave, notification, audit | `services/_shared/postgres/rls.go`, `docs/adr/0002-postgres-rls-for-tenant-isolation.md` |
| FR-HRS-SEC.2 | Mandatory cross-tenant tests in every module (tenant A cannot read/write tenant B data) run in CI against a real Postgres with an NOBYPASSRLS app role | M | Partial: integration test files exist for auth and employee; CI job `test-integration` only covers auth, employee, leave and connects as `hris_admin` created as `POSTGRES_USER` (bypasses RLS, `[INFERENCE]`), so RLS policies are untested; attendance, notification, audit, document are not covered; the content of cross-tenant tests has not been examined | `services/auth-service/internal/integration/auth_integration_test.go`, `services/employee-service/internal/integration/employee_integration_test.go`, `.github/workflows/ci-go.yml` |
| FR-HRS-SEC.3 | NIK, NPWP, birth date encrypted with `kit/piicrypt` with a blind index; old data encrypted during migration; plain `TEXT` columns removed | M | Not started | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| FR-HRS-SEC.4 | Opaque session: HttpOnly Secure SameSite cookie, hash in Dragonfly, no JWT and no token in the browser; tenant and actor always from the session | M | Not started: JWT and cookie `auth_user` readable by the client | `services/auth-service/internal/infrastructure/jwt_signer.go`, `apps/web/src/lib/stores/auth.ts` |
| FR-HRS-SEC.5 | Deny-by-default: interceptor/middleware rejects every route/RPC without an explicit permission; internal routes use an HMAC service token | M | Not started | `services/_shared/server/grpc_interceptors.go` |
| FR-HRS-SEC.6 | Rate limiting of login and sensitive endpoints per address and per account with Dragonfly; threshold `tbd` | M | Not started | `apps/api-gateway/routes/api.php` |
| FR-HRS-SEC.7 | No default secrets or credentials in code or tracked files; secrets from env with fail-fast validation; `.env.example`; secrets ever committed are rotated and purged from history | M | Partial: configuration from env exists but dev credentials are written in tracked files and default values are in code | `services/_shared/config/loader.go`, `deploy/docker/docker-compose.yml` |
| FR-HRS-SEC.8 | Data subject rights: HR can export/correct/delete one employee's data on request, recorded in audit; processing basis and consent recorded | S | Not started | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| FR-HRS-SEC.9 | Web security headers (CSP, `X-Content-Type-Options`, HSTS, `Referrer-Policy`), strict CORS, CSRF protection for the session cookie | S | Not started | `apps/web/svelte.config.js` |
| FR-HRS-SEC.10 | `slog` logs contain no class 2 PII, passwords, tokens, or credentials; request IDs correlated | M | Not started: no redaction policy; `zap` is used | `services/_shared/observability/logging.go` |
| FR-HRS-SEC.11 | Retention and deletion per data class run by a scheduled worker with an audit trail; retention period `tbd` (Q-HRS-04) | S | Not started | `services/document-service/migrations/001_create_document_schema.up.sql` |

## 5. Data Model

Current state: one Postgres schema per service in a single database `hris_db` (`deploy/docker/init-postgres.sql`) or per service in `docker-compose.production.yml` (per-service databases are not created by the init script; `Pending Verification`). Target: a single database `hris`, a single app role `hris` NOBYPASSRLS, a separate migration role, no cross-database FK (D-HRS-06).

PII classes: **PII-0** none; **PII-1** direct identity (name, email, phone); **PII-2** regulated/sensitive, must be encrypted with `kit/piicrypt` (NIK, NPWP, birth date); **SEC** technical secrets (password hash, SMTP credentials). Mapping of NIK/NPWP/leave data to the final UU PDP category is a legal decision (Q-HRS-04); sick leave reasons may be health data `[INFERENCE]`.

| Entity | Key columns | Current RLS | PII class | Evidence |
|---|---|---|---|---|
| `auth.tenants` | `id`, `name`, `slug`, status | No RLS; INSERT/UPDATE only `hris_admin` | PII-0 | `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| `auth.users` | `id`, `tenant_id`, `email`, `password_hash`, status | ENABLE+FORCE+WITH CHECK | PII-1, SEC (hash) | `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| `auth.roles`, `auth.permissions`, `auth.role_permissions`, `auth.user_roles` | `tenant_id` (roles), permission code | `role_permissions` and `user_roles` without RLS | PII-0 | `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| `auth.login_attempts`, `auth.refresh_tokens` | login attempts, tokens | `refresh_tokens` ENABLE+FORCE+WITH CHECK; both unused by code | PII-1 (IP) | `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| `employee.departments` | `tenant_id`, `name`, `parent_id` (unused) | ENABLE+FORCE+WITH CHECK | PII-0 | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| `employee.positions` | `tenant_id`, `name`, `level` | ENABLE+FORCE+WITH CHECK | PII-0 | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| `employee.employees` | `tenant_id`, `employee_number` (unique), `manager_id`, `national_id`, `tax_id`, `birth_date`, `phone`, `termination_date` | ENABLE+FORCE+WITH CHECK | PII-1; **PII-2: `national_id`, `tax_id`, `birth_date`** | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| `attendance.attendance_records` | `tenant_id`, `employee_id`, `date`, `check_in_at`, `check_out_at`, coordinates, `status`, `is_override` | ENABLE without FORCE, policy without `WITH CHECK` | PII-1 (location) | `services/attendance-service/migrations/001_create_attendance_schema.up.sql` |
| `leave.leave_types`, `leave_requests`, `leave_balances` | `tenant_id`, `employee_id`, `leave_type_id`, date range, `days_count`, balance | ENABLE without FORCE, without `WITH CHECK` | PII-1; `reason` may be health `[INFERENCE]` | `services/leave-service/migrations/001_create_leave_schema.up.sql` |
| `notification.notifications` | `tenant_id`, `recipient_id`, `channel`, `template_key`, `variables`, status | ENABLE without FORCE; `WITH CHECK` via a separate policy | PII-1 | `services/notification-service/migrations/001_create_notification_schema.up.sql` |
| `notification.notification_channel_configs` | `tenant_id`, `channels`, `email_config` | ENABLE without FORCE; `WITH CHECK` via a separate policy | SEC (plain SMTP credentials) | `services/notification-service/migrations/001_create_notification_schema.up.sql` |
| `audit.entries` (monthly partitions) | `tenant_id`, `actor_id`, `action`, `resource_type`, `resource_id`, `changes`, `created_at` | RLS disabled | PII-1 (actor) | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| `document.files` | `tenant_id`, `entity_type`, `entity_id`, `minio_bucket`, `minio_key`, `classification`, `retention_days`, versions | ENABLE+FORCE+WITH CHECK | PII-2 (document content) | `services/document-service/migrations/001_create_document_schema.up.sql` |
| `document.file_versions` | `file_id`, `version_number`, `minio_key` | no `tenant_id`, no RLS | PII-2 (content) | `services/document-service/migrations/001_create_document_schema.up.sql` |
| `*.processed_events` (auth, employee, leave, notification, audit, document) | `event_id`, `processed_at` | no RLS (internal idempotency table) | PII-0 | `services/leave-service/migrations/001_create_leave_schema.up.sql` |


Additional target tables: `outbox` (per product, columns `event_id` UUIDv7, `event_type`, `payload`, `created_at`, `published_at`), `audit_logs` (`kit/audit`, append-only with trigger), `sessions` metadata (optional; hash in Dragonfly), `entitlement_snapshot` (suite mode), `org_id`/`external_ref` on `tenants` for the link to the platform org. All tenant tables declare `tenant_id` NOT NULL and a `tenant_isolation` policy (D8).

Schema findings that must be resolved at the stabilisation gate: (a) attendance migration 001 has invalid syntax; (b) duplicate indexes without `IF NOT EXISTS`: `idx_attendance_employee_date` (attendance 001 line 29 and 002) and `idx_leave_requests_employee_id`, `idx_leave_requests_status`, `idx_leave_requests_approved_by` (leave 001 and 002), so migration 002 of both services is expected to fail (`Pending Verification`); (c) role `app` in the attendance, leave, notification migrations while the init script only creates `hris_app` and `hris_admin`; (d) RLS without FORCE and without `WITH CHECK` on attendance and leave (notification without FORCE, with `WITH CHECK` via a separate policy); (e) `current_setting('app.tenant_id')` without `missing_ok` on attendance, leave, and notification will raise an error if the variable has not been set.

---

## 6. API Specification

### 6.1 Provisions

- **Target:** REST `/api/v1` on stdlib ServeMux, envelope `{success,data,error}` (`kit/httpx`), session cookie `stery_session` in suite mode (standalone mode cookie name `tbd`), explicit permission per route (`RequirePermission`), tenant and actor from the session. gRPC `api/proto/v1` for machine/platform clients with a deny-by-default interceptor.
- **As-is:** 47 RPCs across eight proto files (`proto/hris/*/v1/*.proto`); `google.api.http` annotations give `/v1/...` paths. Those HTTP paths are **not served by any process** (no `runtime.NewServeMux`/`ListenAndServe` in `services/**` outside `*.pb.gw.go`; 11.3). The Laravel gateway only exposes `auth/*` and leave routes (2.3). Therefore the status column below assesses the **RPC implementation**, while HTTP availability for all routes is `Belum`.
- Target route = proto-annotated route with an `/api` prefix. Static routes take precedence over `{id}` patterns in ServeMux (e.g. `/api/v1/employees/org-chart` before `/api/v1/employees/{id}`).
- Target permissions use `resource:action` codes (3.3). `publik` = no session, with rate limit.

### 6.2 Map of 47 RPCs to target routes

Status follows 1.3: real but defective handler = `Sebagian`; empty or missing handler = `Belum`.

| # | RPC | Target route (`/api/v1`) | Target permission | Status | Evidence |
|---|---|---|---|---|---|
| 1 | `AuthService.Login` | `POST /api/v1/auth/login` | public | Partial | `services/auth-service/internal/interfaces/grpc/auth_service.go` |
| 2 | `AuthService.RefreshToken` | removed in target (sliding opaque session); as-is `POST /v1/auth/refresh` | - | Partial | `services/auth-service/internal/application/commands/refresh_token.go` |
| 3 | `AuthService.ValidateToken` | gRPC `SessionService/Introspect` (internal); no public route | service token | Partial | `services/auth-service/internal/interfaces/grpc/auth_service.go` |
| 4 | `AuthService.RevokeToken` | `POST /api/v1/auth/logout` | valid session | Partial | `services/auth-service/internal/application/commands/revoke_token.go` |
| 5 | `AuthService.GetPermissions` | `GET /api/v1/auth/me` (current user's permissions); `GET /api/v1/users/{id}/permissions` for admin | `user:manage` (admin) | Partial | `services/auth-service/internal/interfaces/grpc/auth_service.go` |
| 6 | `AuthService.RegisterTenant` | `POST /api/v1/auth/register` | public | Partial | `services/auth-service/internal/application/commands/register_tenant.go` |
| 7 | `EmployeeService.CreateEmployee` | `POST /api/v1/employees` | `employee:write` | Partial | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| 8 | `EmployeeService.GetEmployee` | `GET /api/v1/employees/{id}` | `employee:read` | Partial | `services/employee-service/internal/application/queries/get_employee.go` |
| 9 | `EmployeeService.UpdateEmployee` | `PATCH /api/v1/employees/{id}` | `employee:write` | Not started | `proto/hris/employee/v1/employee.proto` |
| 10 | `EmployeeService.TerminateEmployee` | `POST /api/v1/employees/{id}/terminate` | `employee:delete` | Partial | `services/employee-service/internal/application/commands/terminate_employee.go` |
| 11 | `EmployeeService.ListEmployees` | `GET /api/v1/employees` | `employee:read` | Partial | `services/employee-service/internal/application/queries/list_employees.go` |
| 12 | `EmployeeService.GetOrgChart` | `GET /api/v1/employees/org-chart` | `employee:read` | Not started | `proto/hris/employee/v1/employee.proto` |
| 13 | `EmployeeService.CreateDepartment` | `POST /api/v1/departments` | `employee:write` | Partial | `services/employee-service/internal/application/commands/create_department.go` |
| 14 | `EmployeeService.ListDepartments` | `GET /api/v1/departments` | `employee:read` | Partial | `services/employee-service/internal/application/queries/list_departments.go` |
| 15 | `EmployeeService.CreatePosition` | `POST /api/v1/positions` | `employee:write` | Partial | `services/employee-service/internal/application/commands/create_position.go` |
| 16 | `EmployeeService.ListPositions` | `GET /api/v1/positions` | `employee:read` | Partial | `services/employee-service/internal/application/queries/list_positions.go` |
| 17 | `AttendanceService.CheckIn` | `POST /api/v1/attendance/check-in` | `attendance:write` (self) | Partial | `services/attendance-service/internal/application/commands/check_in.go` |
| 18 | `AttendanceService.CheckOut` | `POST /api/v1/attendance/check-out` | `attendance:write` (self) | Partial | `services/attendance-service/internal/application/commands/check_out.go` |
| 19 | `AttendanceService.GetAttendanceSummary` | `GET /api/v1/attendance/summary/{employee_id}` | `attendance:read` | Partial | `services/attendance-service/internal/application/queries/get_daily_summary.go` |
| 20 | `AttendanceService.ListAttendance` | `GET /api/v1/attendance` | `attendance:read` | Partial | `services/attendance-service/internal/interfaces/grpc/attendance_service.go` |
| 21 | `AttendanceService.OverrideAttendance` | `POST /api/v1/attendance/override` | `attendance:write` (HR/supervisor) | Partial | `services/attendance-service/internal/application/commands/override_attendance.go` |
| 22 | `LeaveService.ApplyLeave` | `POST /api/v1/leave-requests` | `leave:write` | Partial | `services/leave-service/internal/application/commands/create_leave_request.go` |
| 23 | `LeaveService.GetLeaveRequest` | `GET /api/v1/leave-requests/{id}` | `leave:read` | Partial | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 24 | `LeaveService.ListLeaveRequests` | `GET /api/v1/leave-requests` | `leave:read` | Partial | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 25 | `LeaveService.ApproveLeave` | `POST /api/v1/leave-requests/{id}/approve` | `leave:approve` | Partial | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 26 | `LeaveService.RejectLeave` | `POST /api/v1/leave-requests/{id}/reject` | `leave:approve` | Partial | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 27 | `LeaveService.CancelLeave` | `POST /api/v1/leave-requests/{id}/cancel` | `leave:write` | Partial | `services/leave-service/internal/domain/leave_request.go` |
| 28 | `LeaveService.GetLeaveBalance` | `GET /api/v1/leave-balance/{employee_id}` | `leave:read` | Partial | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 29 | `LeaveService.ListLeaveTypes` | `GET /api/v1/leave-types` | `leave:read` | Partial | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 30 | `NotificationService.SendNotification` | internal (gRPC/port `Notification`); no public route | service token | Partial | `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| 31 | `NotificationService.GetDeliveryStatus` | `GET /api/v1/notifications/{notification_id}` | `notification:read` | Partial | `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| 32 | `NotificationService.ListNotifications` | `GET /api/v1/notifications` | `notification:read` | Partial | `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| 33 | `NotificationService.MarkRead` | `POST /api/v1/notifications/mark-read` | `notification:read` | Partial | `services/notification-service/internal/application/commands/mark_notification_read.go` |
| 34 | `NotificationService.UpdateChannelConfig` | `PUT /api/v1/notifications/channel-config` | `settings:write` | Partial | `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| 35 | `AuditService.Record` | internal (port `Audit.Record`); no public route | service token | Partial | `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| 36 | `AuditService.Query` | `GET /api/v1/audit/entries` | `audit:read` | Partial | `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| 37 | `AuditService.GetEntry` | `GET /api/v1/audit/entries/{entry_id}` | `audit:read` | Partial | `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| 38 | `DocumentService.UploadDocument` | `POST /api/v1/documents/upload` (target: request upload presign) | `document:write` | Partial | `services/document-service/internal/application/commands/upload_document.go` |
| 39 | `DocumentService.DownloadDocument` | `GET /api/v1/documents/{file_id}/download` | `document:read` | Not started | `services/document-service/internal/interfaces/grpc/document_service.go` |
| 40 | `DocumentService.ListDocuments` | `GET /api/v1/documents` | `document:read` | Partial | `services/document-service/internal/interfaces/grpc/document_service.go` |
| 41 | `DocumentService.GetDocumentMetadata` | `GET /api/v1/documents/{file_id}/metadata` | `document:read` | Partial | `services/document-service/internal/interfaces/grpc/document_service.go` |
| 42 | `DocumentService.DeleteDocument` | `DELETE /api/v1/documents/{file_id}` | `document:delete` | Partial | `services/document-service/internal/application/commands/delete_document.go` |
| 43 | `DocumentService.GetPresignedUrl` | `POST /api/v1/documents/{file_id}/presigned-url` | `document:read` | Not started | `services/document-service/internal/interfaces/grpc/document_service.go` |
| 44 | `DocumentService.RollbackDocumentVersion` | `POST /api/v1/documents/{file_id}/rollback` | `document:write` | Not started | `services/document-service/internal/interfaces/grpc/document_service.go` |
| 45 | `AIService.AnalyzeResume` | `POST /api/v1/ai/analyze-resume` | `ai:use` | Not started | `proto/hris/ai/v1/ai.proto` |
| 46 | `AIService.ForecastAttrition` | `POST /api/v1/ai/forecast-attrition` | `ai:use` | Not started | `proto/hris/ai/v1/ai.proto` |
| 47 | `AIService.Chat` | `POST /api/v1/ai/chat` (SSE stream) | `ai:use` | Not started | `proto/hris/ai/v1/ai.proto` |

Count result: 47 RPCs = auth 6, employee 10, attendance 5, leave 8, notification 5, audit 3, document 7, ai 3. Real handlers: 39; empty or missing: 8 (employee 2, document 3, ai 3). This corrects the initial claim "document 5/7": only 4 of 7 are real (`DownloadDocument` is also a placeholder).

### 6.3 gRPC contract and error mapping

- The proto contracts are the basis of the target `api/proto/v1`; required changes: add user-management RPCs and document/AI permissions (FR-HRS-01.7, FR-HRS-07.*), remove the JWT-token `RefreshToken`/`ValidateToken` and replace them with the platform `SessionService` in suite mode, add `X-Request-ID` as metadata, and the field `org_id`.
- Current gRPC code mapping (usage counts in `internal/interfaces/grpc/*.go`): attendance-service only uses `codes.Internal` (5 times), so domain errors (e.g. already checked in) reach the client as `Internal`; leave-service uses the widest variety (`InvalidArgument`, `NotFound`, `PermissionDenied`, `Unauthenticated`, `FailedPrecondition`, `Internal`). Target: a single mapping table in `kit/httpx` from domain errors to HTTP status and gRPC code: validation -> 400/`InvalidArgument`; not found -> 404/`NotFound`; not permitted -> 403/`PermissionDenied`; no session -> 401/`Unauthenticated`; state conflict -> 409/`FailedPrecondition`; rate limit -> 429/`ResourceExhausted`; entitlement -> 402 or 403 `tbd`.
- REST error envelope: `{"success":false,"error":{"code":"...","message":"..."}}`; `code` is stable, `message` is localised `{id,en}` by the client.

---

## 7. Critical Process Flows

### 7.1 Leave request and approval

Target flow with as-is break points in the right column.

| Step | Target flow | As-is condition | Status | Evidence |
|---|---|---|---|---|
| 1 | The employee opens the leave form; the client loads `GET /leave-types` and `GET /leave-balance/{employee_id}` | the leave type list is empty (no creator/seed); page `/leaves/apply` exists | Partial | `apps/web/src/routes/(app)/leaves/apply/+page.svelte`, `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 2 | `POST /leave-requests` with type, dates, reason; `employee_id` from the session | `employee_id` from the request body; `ApplyLeave` does not fill `days_count` (CHECK `> 0`) | Partial | `services/leave-service/internal/application/commands/create_leave_request.go` |
| 3 | Validation: valid range, working days computed, sufficient balance, no overlap; transaction: insert `PENDING` request, `pending_days += n`, write outbox `leave.requested` | balance checked in the application without row lock; overlap not checked; event payload `{}` | Partial | `services/leave-service/internal/domain/leave_request.go`, `services/leave-service/internal/infrastructure/nats/event_publisher.go` |
| 4 | Consumers: notification to direct supervisor/HR; audit | the notification consumer exists (with tests) but email is not wired and the payload is empty | Partial | `services/notification-service/internal/application/consumers/leave_event_consumer.go` |
| 5 | Supervisor/HR opens the queue (`GET /leave-requests?status=PENDING`) | the status filter uses `LEAVE_STATUS_*` which does not match DB values | Partial | `services/leave-service/internal/interfaces/grpc/leave_service.go`, `apps/web/src/routes/(app)/admin/leaves/+page.svelte` |
| 6 | `POST /leave-requests/{id}/approve|reject`; actor from the session, not the requester themselves, only from `PENDING`; transaction updates the balance and writes outbox | `approver_id` from the body; no self-approval prevention; roles from the `x-user-roles` header | Partial | `services/leave-service/internal/interfaces/grpc/context_guards.go` |
| 7 | Consumers: notification to the requester; attendance for the related days becomes `ON_LEAVE` (FR-HRS-03.7); audit | attendance does not consume leave events | Not started | `services/attendance-service/internal/domain/attendance_record.go` |
| 8 | Cancellation by the requester restores the balance | `Cancel` allows `PENDING`; balance restoration not verified | Partial | `services/leave-service/internal/domain/leave_request.go` |

### 7.2 Onboarding event chain

| Step | Target flow | As-is condition | Status | Evidence |
|---|---|---|---|---|
| 1 | `POST /auth/register` (public, rate-limited): one transaction creates the tenant, three roles, permissions, `hr_admin` admin; outbox `tenant_registered` and `user_registered` (payload carries `tenant_id`) | not atomic; INSERT `auth.tenants` only allowed for `hris_admin` (`Pending Verification`); `user_registered` payload without `tenant_id` | Partial | `services/auth-service/internal/application/commands/register_tenant.go`, `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| 2 | Employee consumes `user_registered`, creates an employee shell for the admin (tenant default department/position), emits `employee.created` | consumer exists but cannot work: `tenant_id` missing, department/position ids randomised, error acked then lost | Partial | `services/employee-service/internal/infrastructure/nats_consumer.go` |
| 3 | Leave consumes `employee.created` and creates balances for each active leave type, idempotent via `processed_events` | consumer exists; `EmployeeValidator` is a no-op; without leave types there are no balances | Partial | `services/leave-service/internal/infrastructure/nats/employee_created_consumer.go` |
| 4 | Notification sends a welcome/activation email and in-app notification | email adapter not wired | Partial | `services/notification-service/cmd/server/main.go` |
| 5 | Audit records `tenant_registered`, `user_registered`, `employee.created` | consumer exists; `resource_id` UUID vs empty string | Pending Verification | `services/audit-service/internal/infrastructure/nats/identity_consumer.go` |

In suite mode, step 1 is replaced: the org and admin are created by the platform; HRIS receives the `Deliver` `org_created` and creates the HRIS tenant linked by `external_ref` (suite FSD 7.5).

### 7.3 Attendance check-in and check-out

| Step | Target flow | As-is condition | Status | Evidence |
|---|---|---|---|---|
| 1 | The employee opens the attendance page (SPA) | there is no attendance page in the web | Not started | `apps/web/src/routes/(app)/dashboard/+page.svelte` |
| 2 | `POST /attendance/check-in` (optional coordinates); `employee_id` and `tenant_id` from the session; rejects if there is already an active record today (tenant zone) | `CheckIn` exists; tenant from the body; migration 001 is invalid so it cannot execute on a fresh DB | Partial | `services/attendance-service/internal/application/commands/check_in.go` |
| 3 | Status `PRESENT` or `LATE` per tenant rules (FR-HRS-03.6) | `LATE` is never computed | Not started | `services/attendance-service/internal/domain/attendance_record.go` |
| 4 | `POST /attendance/check-out`: an active check-in must exist; duration computed; outbox `checked_out` | `CheckOut` exists | Partial | `services/attendance-service/internal/application/commands/check_out.go` |
| 5 | HR/supervisor corrects via `POST /attendance/override` with a reason; new record if none exists | override fails for new records | Partial | `services/attendance-service/internal/application/commands/override_attendance.go` |
| 6 | Event consumed by audit | subject `hris.attendance.record.*` does not match the `hris.operations.>` consumer; payload `{}` | Partial | `services/attendance-service/internal/domain/events.go` |

---

## 8. Security, Privacy and UU PDP

### 8.1 Data classification

| Class | Examples | Target treatment |
|---|---|---|
| PII-2 (regulated/sensitive) | NIK (`national_id`), NPWP (`tax_id`), birth date, employee document content, sick leave reasons `[INFERENCE]` | `kit/piicrypt` encryption (AES-256-GCM, per-purpose AAD, blind index), masked display, audited access, not in logs, not to third parties without a legal basis (FR-HRS-SEC.3) |
| PII-1 (direct identity) | name, email, phone, attendance coordinates | role-based access, not in logs in plain form |
| SEC (technical secrets) | password hash, SMTP credentials, encryption keys | never recorded; SMTP credentials encrypted; keys per environment |
| PII-0 | non-personal organisation data | no special treatment |

### 8.2 Current security posture

| Finding | Status | Evidence |
|---|---|---|
| No authentication/authorisation in any gRPC interceptor; tenant and actor are trusted from the request body or metadata | Not started | `services/_shared/server/grpc_interceptors.go` |
| Client IP address is always `0.0.0.0` | Partial | `services/auth-service/internal/interfaces/grpc/auth_service.go` |
| No account lockout or rate limiting (Laravel `throttle` is never applied; table `login_attempts` is unused) | Not started | `apps/api-gateway/routes/api.php`, `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| JWT RS256 with hardcoded TTL 15 minutes/7 days; no refresh token family reuse detection | Partial | `services/auth-service/cmd/server/main.go`, `services/auth-service/internal/infrastructure/jwt_signer.go` |
| The gateway does not extract the token: `getBearerToken()` returns an empty string; `LeaveController` reads roles from an attribute that is never filled | Not started | `apps/api-gateway/app/Services/BaseGrpcGatewayClient.php`, `apps/api-gateway/app/Http/Middleware/JwtValidation.php` |
| The SvelteKit hook trusts the `auth_user` cookie (role JSON) which is not HttpOnly; the JWT is not verified on the web server | Partial | `apps/web/src/hooks.server.ts`, `apps/web/src/routes/(auth)/login/+page.server.ts` |
| SMTP credentials stored plain in JSONB | Partial | `services/notification-service/migrations/001_create_notification_schema.up.sql` |
| Dev and default credentials written in git-tracked files (values not quoted here) | Partial | `deploy/docker/docker-compose.yml`, `deploy/docker/nats.conf`, `deploy/docker/init-postgres.sql`, `docker-compose.production.yml` |
| RLS ENABLE+FORCE+WITH CHECK only on auth (users, roles, refresh_tokens), employee, and document.files; the rest partial or without RLS | Partial | `services/employee-service/migrations/001_create_employee_schema.up.sql`, `services/document-service/migrations/001_create_document_schema.up.sql` |
| Inter-service gRPC without TLS and without service authentication; UploadDocument cannot succeed because the actor context is not filled | Not started | `services/document-service/internal/interfaces/grpc/mapping.go` |
| The compiled binary `server.exe` is in the working tree of several services but is **not** tracked by git (ignored by `*.exe`); it needs to be cleaned from the working tree | Partial | `.gitignore` |

### 8.3 UU PDP requirements

Law No. 27 of 2022 on Personal Data Protection (UU PDP) applies to the processing of employee data. The repo contains no policy, consent wording, retention, or incident procedure (ADR-0004 line 47 mentions PDPA only once in the payroll context; `docs/adr/0004-defer-payroll-to-phase2.md`). The following requirements are **product needs**; the final legal content (data categories, processing basis, deadlines, notification obligations) requires legal review and is written `tbd` until Q-HRS-04 and Q-HRS-14 are answered.

| Area | Product requirement | Related FR | Legal content |
|---|---|---|---|
| Classification and data map | Inventory of PII columns per class (8.1); NIK and NPWP treated as PII-2 by default | FR-HRS-02.8, FR-HRS-SEC.3 | mapping to UU PDP categories `tbd` |
| Encryption | `kit/piicrypt` for NIK, NPWP, birth date; separate keys per environment; key rotation | FR-HRS-SEC.3 | - |
| Processing basis and consent | Record the processing basis per data class; explicit consent for processing outside the employment contract (e.g. sending data to an LLM, FR-HRS-08.*) | FR-HRS-SEC.8 | legal basis `tbd` |
| Retention and deletion | Retention period per data class (employee, attendance, leave, documents, audit); deletion/anonymisation at the deadline | FR-HRS-SEC.11, FR-HRS-06.5, FR-HRS-07.5 | retention period `tbd` (Q-HRS-04) |
| Data subject rights | Access, correction, deletion, restriction, portability of employee data on request via HR; request log | FR-HRS-SEC.8 | response deadline `tbd` |
| Controller and processor | Tenant = controller, Stery = processor; data processing agreement template | - | template `tbd` (Q-HRS-14) |
| Incidents | Procedure for detecting and notifying data breach incidents; security logs | FR-HRS-SEC.10 | deadline and format `tbd` |
| Transfer to third parties | No PII-2 to LLM/email/SMS providers without a basis and an agreement | FR-HRS-08.3, FR-HRS-08.4 | `tbd` (Q-HRS-09) |
| Tenant isolation | RLS and cross-tenant tests as a technical confidentiality control | FR-HRS-SEC.1, FR-HRS-SEC.2 | - |

### 8.4 Cross-module security requirements

The cross-module security requirements FR-HRS-SEC.1 to FR-HRS-SEC.11 are defined in section 4.SEC.

---

## 9. Technical Non-Functional Requirements

This section translates NFR-HRS (defined in `PRD_HRIS.id.md` section 6) into technical realisation and verification methods. Status and evidence assess the current condition. Numeric target values not found in the repo are written `tbd` (Q-HRS-07, Q-HRS-15).

| NFR | Technical realisation | Verification | Status | Evidence |
|---|---|---|---|---|
| NFR-HRS-01 | RLS ENABLE+FORCE+`WITH CHECK` on all tenant tables via `kit/tenantdb.WithTenant`; NOBYPASSRLS role (FR-HRS-SEC.1) | cross-tenant tests per module on real Postgres (FR-HRS-SEC.2) | Partial | `services/_shared/postgres/rls.go` |
| NFR-HRS-02 | Opaque session, hash in Dragonfly, HttpOnly Secure cookie; TTL `tbd` (FR-HRS-SEC.4) | login/logout/expiry integration tests; cookie checks in e2e | Not started | `services/auth-service/internal/infrastructure/jwt_signer.go` |
| NFR-HRS-03 | `RequirePermission(code)` + deny-by-default gRPC interceptor (FR-HRS-SEC.5) | route-permission table test: every route has a permission (fails if any route lacks one) | Not started | `services/_shared/server/grpc_interceptors.go` |
| NFR-HRS-04 | `kit/piicrypt` AES-256-GCM, per-purpose AAD, blind index, key rotation (FR-HRS-SEC.3) | unit round-trip and rotation tests; old-data migration test | Not started | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| NFR-HRS-05 | Append-only `audit_logs` with a trigger rejecting UPDATE/DELETE/TRUNCATE (FR-HRS-06.4) | an integration test that attempts UPDATE/DELETE must fail | Partial | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| NFR-HRS-06 | Availability target `tbd`; `/healthz`; automatic restart by the process manager | monitoring; no hosting target yet (Q-HRS-12) | Not started | `docker-compose.production.yml` |
| NFR-HRS-07 | p95 latency `tbd`; indexes matching query patterns; pagination mandatory | load test on the real stack (none yet) | Not started | `services/leave-service/migrations/002_add_performance_indexes.up.sql` |
| NFR-HRS-08 | Dragonfly rate limiter (Lua script) per address/account and route class; account lockout (FR-HRS-01.5, FR-HRS-SEC.6) | integration test: N failed attempts yield 429 | Not started | `apps/api-gateway/routes/api.php` |
| NFR-HRS-09 | Reversible goose migrations; CI job `up -> down -> up` for **all** modules on a fresh DB | green CI job | Partial: CI runs `up -> down -> up` only for auth, employee, leave with golang-migrate; attendance, audit, notification, document are not covered, and the attendance migration is invalid | `.github/workflows/ci-go.yml` |
| NFR-HRS-10 | `scripts/check.sh` like SELA; coverage >= 80%; `-race` | green CI; coverage threshold | Partial: Go CI uses `-race`, but CI `GO_VERSION` is 1.24 while `go.work` is 1.25.0; no coverage threshold; document-service has no test files | `.github/workflows/ci-go.yml`, `go.work` |
| NFR-HRS-11 | `slog` JSON without PII, `/healthz`; OTel/Prometheus optional (Q-HRS-08) | test that logs contain no PII; green `/healthz` | Partial: OTel `InitTracer` and `NewMetricsRegistry` are defined in `_shared` but never called by services; `StartMetricsServer` binds `:8081` for all services | `services/_shared/observability/otel.go`, `services/_shared/observability/metrics.go` |
| NFR-HRS-12 | `{id,en}` i18n catalogues with parity test; WIB; IDR when there are monetary values | catalogue parity test in Vitest | Not started: no i18n library | `apps/web/package.json` |
| NFR-HRS-13 | Env + fail-fast validation + `.env.example`; no secrets in tracked files | secret scanner in CI; review | Partial | `services/_shared/config/loader.go`, `deploy/docker/docker-compose.yml` |
| NFR-HRS-14 | Scheduled Postgres backups + restore test; RPO/RTO `tbd` (Q-HRS-07) | scheduled restore drill | Not started | `docker-compose.production.yml` |
| NFR-HRS-15 | Retention worker per data class (FR-HRS-SEC.11) | test that the worker deletes/anonymises expired data | Not started | `services/document-service/migrations/001_create_document_schema.up.sql` |
| NFR-HRS-16 | `STERY_MODE=standalone\|suite` with two adapters per port; standalone regression test | the same test suite passes in both modes for standalone behaviour | Not started | `proto/hris/auth/v1/auth.proto` |
| NFR-HRS-17 | UI components accessible by keyboard/screen reader; WCAG level `tbd` | automated + manual accessibility audit on main flows | Not started | `apps/web/src/routes/(auth)/login/+page.svelte` |
| NFR-HRS-18 | Transactional outbox (`FOR UPDATE SKIP LOCKED`) + idempotent consumers `processed_events` | test: a failure mid-publication does not lose the event | Partial: `processed_events` exists in all schemas; no outbox | `services/leave-service/migrations/001_create_leave_schema.up.sql` |

---

## 10. Screen Structure

### 10.1 Current pages

| Web route | Function | Status | Evidence |
|---|---|---|---|
| `/` | landing page | Partial | `apps/web/src/routes/+page.svelte` |
| `/login` | login form; the server action calls `http://localhost:3000/api/auth/login` (hard-coded value, not a real backend); after success it sets cookies `auth_token` (HttpOnly) and `auth_user` (not HttpOnly) | Partial | `apps/web/src/routes/(auth)/login/+page.server.ts`, `apps/web/src/routes/(auth)/login/+page.svelte` |
| `/dashboard` | user dashboard | Partial | `apps/web/src/routes/(app)/dashboard/+page.svelte` |
| `/leaves/apply` | leave request form | Partial | `apps/web/src/routes/(app)/leaves/apply/+page.svelte` |
| `/admin/leaves` | leave approval queue for managers/HR; role protection in the server hook reads the role from the `auth_user` cookie | Partial | `apps/web/src/routes/(app)/admin/leaves/+page.svelte`, `apps/web/src/hooks.server.ts` |
| `403.svelte` | access denied page; probably not a route because the file name has no `+` prefix (`[INFERENCE]`) | Pending Verification | `apps/web/src/routes/403.svelte` |

Note: the API client (`src/lib/api/client.ts`) does not send an Authorization header; the `auth.ts` store defines `setSession` which nobody calls, so the initial claim "token stored twice in cookie and localStorage" is not proven (only the cookie is filled). The e2e tests in `tests/e2e/` exist but refer to paths that do not exist (`/leaves` without `/apply`, `[INFERENCE]`) (`apps/web/tests/e2e/auth-and-rbac.spec.ts`). There are no pages for employees, attendance, departments, notifications, audit, documents, and settings.

### 10.2 Target screens (SPA `apps/web`)

SPA route groups: `(auth)` public and `(app)` with a client session guard that calls `GET /api/v1/auth/me`. Target routes:

| Target route | Screen | Related FR | Role |
|---|---|---|---|
| `/login`, `/register` | login; tenant registration | FR-HRS-01.1, FR-HRS-01.2 | public |
| `/dashboard` | summary: today's attendance, leave balance, approval queue | FR-HRS-03.5, FR-HRS-04.7 | all |
| `/employees`, `/employees/new`, `/employees/{id}` | list, create, edit, terminate employees | FR-HRS-02.1..02.4 | `hr_admin` (manager: read team) |
| `/org` | org chart | FR-HRS-02.7 | all (read) |
| `/departments`, `/positions` | manage departments and positions | FR-HRS-02.5, FR-HRS-02.6 | `hr_admin` |
| `/attendance` | check-in/out, my history | FR-HRS-03.1, FR-HRS-03.2 | `employee` |
| `/admin/attendance` | list, summary, correction | FR-HRS-03.3..03.5 | `hr_admin`, `manager` |
| `/leaves`, `/leaves/apply` | history, apply, cancel, balance | FR-HRS-04.3, FR-HRS-04.5..04.7 | `employee` |
| `/admin/leaves`, `/admin/leave-types` | approval; leave types | FR-HRS-04.1, FR-HRS-04.4 | `manager`, `hr_admin` |
| `/notifications` | in-app inbox | FR-HRS-05.4 | all |
| `/documents` | list, upload, download employee documents | FR-HRS-07.1, FR-HRS-07.3, FR-HRS-07.4 | per classification |
| `/admin/audit` | audit trail browsing | FR-HRS-06.2 | `hr_admin`/auditor |
| `/admin/settings` | notification channels, work rules, users and roles | FR-HRS-05.5, FR-HRS-03.6, FR-HRS-01.7 | `hr_admin` |

Every screen complies with: `{id,en}` strings (NFR-HRS-12), empty/error/loading states, keyboard accessibility (NFR-HRS-17), and does not display PII-2 except with valid permission (masking by default).

---

## 11. Traceability and Status Matrix

### 11.1 Status per FR

Concise reference; the full definition of each FR is in sections 4 and 8.4.

| FR | Status | Main gap | Evidence |
|---|---|---|---|
| FR-HRS-01.1 | Partial | not atomic; `auth.tenants` rights only `hris_admin` | `services/auth-service/internal/application/commands/register_tenant.go` |
| FR-HRS-01.2 | Partial | web uses a mock login; JWT is not an opaque session | `services/auth-service/internal/application/commands/login.go` |
| FR-HRS-01.3 | Done | - | `services/auth-service/internal/domain/user.go` |
| FR-HRS-01.4 | Partial | no reuse detection; hardcoded TTL | `services/auth-service/internal/application/commands/refresh_token.go` |
| FR-HRS-01.5 | Not started | `login_attempts` unused | `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| FR-HRS-01.6 | Partial | not enforced; wrong matrix | `services/auth-service/internal/domain/permission.go` |
| FR-HRS-01.7 | Not started | no user RPCs | `proto/hris/auth/v1/auth.proto` |
| FR-HRS-01.8 | Not started | no verification/reset | `services/auth-service/internal/domain/user.go` |
| FR-HRS-01.9 | Partial | IP `0.0.0.0` | `services/auth-service/internal/interfaces/grpc/auth_service.go` |
| FR-HRS-01.10 | Not started | no suite adapter | `proto/hris/auth/v1/auth.proto` |
| FR-HRS-02.1 | Partial | `employee_number` not filled; enum case mismatch | `services/employee-service/internal/infrastructure/postgres/employee_repository.go` |
| FR-HRS-02.2 | Partial | Total = page length; offset 0 | `services/employee-service/internal/application/queries/list_employees.go` |
| FR-HRS-02.3 | Not started | handler does not exist | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| FR-HRS-02.4 | Partial | no cascade | `services/employee-service/internal/application/commands/terminate_employee.go` |
| FR-HRS-02.5 | Partial | `parent_id` unused | `services/employee-service/internal/application/commands/create_department.go` |
| FR-HRS-02.6 | Partial | column `description` does not exist | `services/employee-service/internal/infrastructure/postgres/position_repository.go` |
| FR-HRS-02.7 | Not started | handler does not exist | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| FR-HRS-02.8 | Not started | plain PII | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| FR-HRS-02.9 | Not started | - | `proto/hris/employee/v1/employee.proto` |
| FR-HRS-02.10 | Partial | payload without `tenant_id` | `services/employee-service/internal/infrastructure/nats_consumer.go` |
| FR-HRS-02.11 | Not started | no limit | `services/employee-service/internal/application/commands/create_employee.go` |
| FR-HRS-03.1 | Partial | invalid migration; employee RBAC | `services/attendance-service/internal/application/commands/check_in.go` |
| FR-HRS-03.2 | Partial | same | `services/attendance-service/internal/application/commands/check_out.go` |
| FR-HRS-03.3 | Partial | fails for new records | `services/attendance-service/internal/application/commands/override_attendance.go` |
| FR-HRS-03.4 | Partial | date filter ignored | `services/attendance-service/internal/interfaces/grpc/attendance_service.go` |
| FR-HRS-03.5 | Partial | `employee_id`/`period` ignored | `services/attendance-service/internal/application/queries/get_daily_summary.go` |
| FR-HRS-03.6 | Not started | `LATE`/`HALF_DAY` not computed | `services/attendance-service/internal/domain/attendance_record.go` |
| FR-HRS-03.7 | Not started | no leave consumer | `services/attendance-service/internal/domain/attendance_record.go` |
| FR-HRS-03.8 | Not started | no tenant zone | `services/attendance-service/migrations/001_create_attendance_schema.up.sql` |
| FR-HRS-04.1 | Partial | no creator/seed | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| FR-HRS-04.2 | Partial | no-op validator; no leave types | `services/leave-service/internal/infrastructure/nats/employee_created_consumer.go` |
| FR-HRS-04.3 | Partial | `days_count` not filled; no overlap check | `services/leave-service/internal/application/commands/create_leave_request.go` |
| FR-HRS-04.4 | Partial | actor from body; no self-approval prevention | `services/leave-service/internal/interfaces/grpc/context_guards.go` |
| FR-HRS-04.5 | Partial | `Cancel` allows `PENDING` | `services/leave-service/internal/domain/leave_request.go` |
| FR-HRS-04.6 | Partial | status filter mismatch | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| FR-HRS-04.7 | Partial | cross-year counted in start year | `services/leave-service/internal/domain/leave_balance.go` |
| FR-HRS-04.8 | Not started | - | `services/leave-service/migrations/001_create_leave_schema.up.sql` |
| FR-HRS-04.9 | Not started | - | `services/leave-service/internal/domain/leave_balance.go` |
| FR-HRS-05.1 | Partial | - | `services/notification-service/migrations/001_create_notification_schema.up.sql` |
| FR-HRS-05.2 | Partial | adapter not wired; templates not loaded | `services/notification-service/internal/infrastructure/channels/email_adapter.go` |
| FR-HRS-05.3 | Partial | empty payload | `services/notification-service/internal/application/consumers/leave_event_consumer.go` |
| FR-HRS-05.4 | Partial | `MarkRead` fails silently | `services/notification-service/internal/application/commands/mark_notification_read.go` |
| FR-HRS-05.5 | Partial | plain SMTP | `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| FR-HRS-05.6 | Not started | - | `proto/hris/notification/v1/notification.proto` |
| FR-HRS-05.7 | Not started | - | `proto/hris/notification/v1/notification.proto` |
| FR-HRS-06.1 | Partial | `resource_id` UUID vs empty string | `services/audit-service/internal/infrastructure/nats/identity_consumer.go` |
| FR-HRS-06.2 | Pending Verification | `hris_app` rights INSERT only | `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| FR-HRS-06.3 | Partial | `GetEntry` uses `Limit:1` | `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| FR-HRS-06.4 | Partial | rights restriction, no trigger | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| FR-HRS-06.5 | Partial | static partitions | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| FR-HRS-06.6 | Partial | subject mismatch | `services/audit-service/internal/infrastructure/nats/operations_consumer.go` |
| FR-HRS-06.7 | Not started | RLS disabled | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| FR-HRS-07.1 | Partial | always `Unauthenticated` | `services/document-service/internal/application/commands/upload_document.go` |
| FR-HRS-07.2 | Not started | no scanner | `services/document-service/migrations/001_create_document_schema.up.sql` |
| FR-HRS-07.3 | Not started | placeholder | `services/document-service/internal/interfaces/grpc/document_service.go` |
| FR-HRS-07.4 | Partial | deleted not filtered | `services/document-service/internal/interfaces/grpc/document_service.go` |
| FR-HRS-07.5 | Partial | no retention worker | `services/document-service/internal/application/commands/delete_document.go` |
| FR-HRS-07.6 | Not started | not wired | `services/document-service/internal/application/commands/create_document_version.go` |
| FR-HRS-07.7 | Partial | not enforced | `services/document-service/migrations/001_create_document_schema.up.sql` |
| FR-HRS-07.8 | Not started | no validation | `services/document-service/internal/application/commands/upload_document.go` |
| FR-HRS-08.1 | Not started | proto only | `proto/hris/ai/v1/ai.proto` |
| FR-HRS-08.2 | Not started | proto only | `proto/hris/ai/v1/ai.proto` |
| FR-HRS-08.3 | Not started | proto only | `proto/hris/ai/v1/ai.proto` |
| FR-HRS-08.4 | Not started | `ai-service` commented out | `go.work` |
| FR-HRS-SEC.1 | Partial | RLS uneven | `services/_shared/postgres/rls.go` |
| FR-HRS-SEC.2 | Partial | CI 3 services, bypass RLS | `.github/workflows/ci-go.yml` |
| FR-HRS-SEC.3 | Not started | plain PII | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| FR-HRS-SEC.4 | Not started | JWT | `services/auth-service/internal/infrastructure/jwt_signer.go` |
| FR-HRS-SEC.5 | Not started | no authn/authz | `services/_shared/server/grpc_interceptors.go` |
| FR-HRS-SEC.6 | Not started | `throttle` not applied | `apps/api-gateway/routes/api.php` |
| FR-HRS-SEC.7 | Partial | dev credentials tracked | `deploy/docker/docker-compose.yml` |
| FR-HRS-SEC.8 | Not started | - | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| FR-HRS-SEC.9 | Not started | - | `apps/web/svelte.config.js` |
| FR-HRS-SEC.10 | Not started | `zap`; no redaction | `services/_shared/observability/logging.go` |
| FR-HRS-SEC.11 | Not started | - | `services/document-service/migrations/001_create_document_schema.up.sql` |

Count summary (counted from the table above, 75 FRs): `Selesai` 1; `Sebagian` 40; `Belum` 33; `Pending Verification` 1; `Ditunda` 0 (modules 09-12 have no FR). This count is not a readiness percentage.

### 11.2 BR -> FR matrix

Every BR-HRS-01..22 (defined in `BRD_HRIS.id.md` section 6) is mapped to the FRs that fulfil it. The two Deferred BRs have no FR (modules 09-12 only contain the deferral rationale, 4.09-4.12).

| BR | Fulfilling FR | Notes |
|---|---|---|
| BR-HRS-01 | FR-HRS-01.1, FR-HRS-01.6, FR-HRS-01.9, FR-HRS-02.10 | Tenant self-registration and first admin |
| BR-HRS-02 | FR-HRS-SEC.1, FR-HRS-SEC.2, FR-HRS-06.7 | Tenant isolation and cross-tenant tests |
| BR-HRS-03 | FR-HRS-01.6, FR-HRS-01.7, FR-HRS-SEC.4, FR-HRS-SEC.5 | Authorisation in the backend |
| BR-HRS-04 | FR-HRS-02.1, FR-HRS-02.2, FR-HRS-02.3, FR-HRS-02.4, FR-HRS-02.5, FR-HRS-02.6, FR-HRS-02.7, FR-HRS-02.9 | Master data |
| BR-HRS-05 | FR-HRS-03.1, FR-HRS-03.2, FR-HRS-03.3, FR-HRS-03.4, FR-HRS-03.5, FR-HRS-03.7 | Attendance and correction |
| BR-HRS-06 | FR-HRS-03.6, FR-HRS-03.8 | Work rules and time zone per tenant |
| BR-HRS-07 | FR-HRS-04.1, FR-HRS-04.2, FR-HRS-04.3, FR-HRS-04.4, FR-HRS-04.5, FR-HRS-04.6, FR-HRS-04.7, FR-HRS-04.8, FR-HRS-04.9 | End-to-end leave |
| BR-HRS-08 | FR-HRS-05.1, FR-HRS-05.2, FR-HRS-05.3, FR-HRS-05.4, FR-HRS-05.5, FR-HRS-05.6, FR-HRS-05.7 | Notifications |
| BR-HRS-09 | FR-HRS-06.1, FR-HRS-06.2, FR-HRS-06.3, FR-HRS-06.4, FR-HRS-06.5, FR-HRS-06.6 | Audit trail |
| BR-HRS-10 | FR-HRS-07.1, FR-HRS-07.2, FR-HRS-07.3, FR-HRS-07.4, FR-HRS-07.5, FR-HRS-07.6, FR-HRS-07.7, FR-HRS-07.8 | Documents |
| BR-HRS-11 | FR-HRS-08.1, FR-HRS-08.2, FR-HRS-08.3, FR-HRS-08.4 | AI (policy prerequisite) |
| BR-HRS-12 | FR-HRS-02.8, FR-HRS-SEC.3, FR-HRS-SEC.8, FR-HRS-SEC.10, FR-HRS-SEC.11 | UU PDP (8.3) |
| BR-HRS-13 | (Deferred, no FR; see 4.09 and 4.10) | ADR-0004 |
| BR-HRS-14 | (Deferred, no FR; see 4.11 and 4.12) | ADR-0004 |
| BR-HRS-15 | FR-HRS-01.10, FR-HRS-02.11 | Two modes (standalone/suite) |
| BR-HRS-16 | FR-HRS-02.11, FR-HRS-01.10 | Dimension `hrs.employees.max` and behaviour when exceeding the limit |
| BR-HRS-17 | FR-HRS-01.2, FR-HRS-02.3, FR-HRS-02.7, FR-HRS-07.1, FR-HRS-07.6, FR-HRS-SEC.3, FR-HRS-SEC.6, FR-HRS-SEC.7 | HRIS stabilisation gate (BRD 12); up->down->up migration via NFR-HRS-09 |
| BR-HRS-18 | FR-HRS-SEC.4, FR-HRS-SEC.5, FR-HRS-SEC.9, FR-HRS-05.3, FR-HRS-06.1 | Stack alignment via 2.1 and 2.7; technical requirements in section 9 |
| BR-HRS-19 | FR-HRS-03.8, FR-HRS-05.2 | Two languages, WIB; interface i18n via NFR-HRS-12 |
| BR-HRS-20 | FR-HRS-01.5, FR-HRS-SEC.6, FR-HRS-SEC.10 | Operability: `/healthz` and backups via NFR-HRS-06, NFR-HRS-14 |
| BR-HRS-21 | FR-HRS-01.10, FR-HRS-02.4, FR-HRS-02.10 | Hypotheses INT-1 and INT-2 (2.5) |
| BR-HRS-22 | (no FR; fulfilled by 11.1 and 11.3 of this document) | Status reporting consistent with code |

### 11.3 Contradictions between status documents and code

This section records claims of status documents, ADRs, and `CLAUDE.md` that **contradict the code**. These claims are not used as fact elsewhere in this document. The `Evidence` column points to the code or file that shows reality; the source document of the claim is named in the first column.

#### 11.3.1 Technical defects verified from code

| Item | Reality according to code | Evidence |
|---|---|---|
| Attendance migration 001 is invalid | Line 25 `CONSTRAINT attendance_unique_per_day UNIQUE (tenant_id, employee_id, date) WHERE NOT is_override` is not valid PostgreSQL syntax (the `WHERE` clause is only for unique indexes). Lines 45-46 `GRANT ... TO app` use role `app`, not `hris_app`; the init script only creates `hris_app` and `hris_admin`. RLS `ENABLE` without `FORCE` and policy without `WITH CHECK`. Role `app` is also used by the leave and notification migrations. | `services/attendance-service/migrations/001_create_attendance_schema.up.sql`, `deploy/docker/init-postgres.sql` |
| Duplicate indexes | Attendance 002 (`idx_attendance_employee_date`) and leave 002 (`idx_leave_requests_employee_id`, `_status`, `_approved_by`) create indexes with the same name as migration 001 without `IF NOT EXISTS` (`Pending Verification`, not yet run) | `services/attendance-service/migrations/002_add_performance_indexes.up.sql`, `services/leave-service/migrations/002_add_performance_indexes.up.sql` |
| No HTTP mux is mounted | There is no `runtime.NewServeMux` or `ListenAndServe` in `services/**` outside `*.pb.gw.go` files. The grpc-gateway HTTP path called by the Laravel gateway is never served; ADR-0001 states the gateway runs in each service. The only HTTP that exists is the metrics server `:8081` (`StartMetricsServer`). The `*.pb.gw.go` files are in `gen/go`, which is ignored by git. | `services/auth-service/cmd/server/main.go`, `services/_shared/observability/metrics.go`, `docs/adr/0001-grpc-gateway-over-grpc-php.md` |
| The web login is a mock | `apps/web/src/routes/(auth)/login/+page.server.ts` lines 38-40 call `http://localhost:3000/api/auth/login`, not the gateway/services; `setSession` in `src/lib/stores/auth.ts` line 58 has no caller | `apps/web/src/routes/(auth)/login/+page.server.ts`, `apps/web/src/lib/stores/auth.ts` |
| The gateway client does not pass the token | `getBearerToken()` returns an empty string; `AuthServiceClient` calls the real gRPC path `/hris.auth.v1.AuthService/Login`, not `/v1/auth/login` | `apps/api-gateway/app/Services/BaseGrpcGatewayClient.php`, `apps/api-gateway/app/Services/AuthServiceClient.php` |
| Documents: empty RPCs | `GetPresignedUrl`, `RollbackDocumentVersion` are empty (`document_service.go` lines 215-221) and `DownloadDocument` is also a placeholder; `UploadDocument` always returns `Unauthenticated` because `actor_id` is never filled | `services/document-service/internal/interfaces/grpc/document_service.go`, `services/document-service/internal/interfaces/grpc/mapping.go` |
| Notifications are not delivered | `cmd/server/main.go` lines 112-132 set `_ = emailAdapter` and `_ = inAppAdapter` | `services/notification-service/cmd/server/main.go` |
| Empty event payloads | Leave events carry private fields without `MarshalJSON`, so `payload` is `{}`; `ApplyLeave` does not fill `DaysCount` | `services/leave-service/internal/infrastructure/nats/event_publisher.go`, `services/leave-service/internal/domain/events.go` |
| gRPC port collision | leave-service and audit-service both bind `50054` by default | `services/leave-service/cmd/server/main.go`, `services/audit-service/cmd/server/main.go` |
| OTel/metrics not active | `InitTracer` and `NewMetricsRegistry` are only defined in `_shared`, not called by services; `StartMetricsServer` binds `:8081` for all services | `services/_shared/observability/otel.go`, `services/_shared/observability/metrics.go` |
| Go versions not uniform | `go.work` and all `go.mod` Go 1.25.0, CI `GO_VERSION: '1.24'` | `go.work`, `.github/workflows/ci-go.yml` |

#### 11.3.2 Status documents that contradict the code

| Document (claim) | Claim | Reality according to code | Evidence |
|---|---|---|---|
| `PROJECT_PROGRESS.md` | "Phase 1 Foundation Nearly Complete" (line 4) and "100% COMPLETE / Production-ready" blocks per service (lines 10-68), including notifications | The services cannot run end-to-end: attendance migration invalid, no HTTP mux, mock web login, notifications not wired | `PROJECT_PROGRESS.md`, `services/attendance-service/migrations/001_create_attendance_schema.up.sql` |
| `PHASE1_RELEASE.md` | "PRODUCTION READY" (line 4) and release certification 2026-06-02 (line 346) | same; also no service authentication and plain PII | `PHASE1_RELEASE.md`, `services/_shared/server/grpc_interceptors.go` |
| `IMPLEMENTATION_COMPLETE.md` | "PRODUCTION READY", "Services (5/5 Complete)", "Production Hardening (8/8 Complete)", "Real email delivery with 5 templates", OpenTelemetry | The email adapter is not wired and templates are not loaded; there are six HTML templates; OTel is never initialised | `IMPLEMENTATION_COMPLETE.md`, `services/notification-service/cmd/server/main.go` |
| `PHASE2_STATUS.md` | Audit "Completion: 50% (Layers 1-3 Complete)" | The audit code has domain, application, infrastructure, interface, and consumer layers | `PHASE2_STATUS.md`, `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| `PHASE2_FOUNDATION.md` | Phase 2 foundation document | Read only as a plan; there is no phase 2 (payroll) code in the repo, consistent with ADR-0004 | `PHASE2_FOUNDATION.md`, `docs/adr/0004-defer-payroll-to-phase2.md` |
| `docs/PHASE5_SUMMARY.md` (2026-06-08) | Readiness 45/100; leave "not implemented"; gateway and frontend empty | Now outdated in the opposite direction: leave handlers exist, the gateway has auth and leave routes, the web has four pages. The 45/100 score itself is not a verified measure | `docs/PHASE5_SUMMARY.md`, `apps/api-gateway/routes/api.php` |
| `CLAUDE.md` section 2 | "sqlc + pgx for all DB access" | sqlc is only used in auth-service | `CLAUDE.md`, `services/auth-service/cmd/server/main.go` |
| `CLAUDE.md` section 3 | Gateway authentication: Laravel Sanctum token mode with RSA JWT, "see ADR-0005" | Sanctum and spatie/permission are installed in `composer.json` but unused; ADR-0005 does not exist (only `docs/adr/0001..0004`); the token is not forwarded to services (`getBearerToken()` is empty) | `CLAUDE.md`, `apps/api-gateway/composer.json`, `apps/api-gateway/app/Services/BaseGrpcGatewayClient.php` |
| `CLAUDE.md` section 6 | All events use an envelope schema with `event_id` UUID v7; publish after commit (outbox if needed) | There is no outbox; events are published directly to NATS; the id type has not been verified (`Pending Verification`) | `CLAUDE.md`, `services/leave-service/internal/infrastructure/nats/event_publisher.go` |
| `CLAUDE.md` section 9 | OpenTelemetry span per service | `InitTracer` is never called | `CLAUDE.md`, `services/_shared/observability/otel.go` |
| `docs/adr/0001-grpc-gateway-over-grpc-php.md` | grpc-gateway runs in every service | there is no HTTP mux in any service | `docs/adr/0001-grpc-gateway-over-grpc-php.md`, `services/auth-service/cmd/server/main.go` |
| `docs/adr/0004-defer-payroll-to-phase2.md` | Phase 2 routes in the gateway return 501 | Those routes do not exist in `api.php` | `docs/adr/0004-defer-payroll-to-phase2.md`, `apps/api-gateway/routes/api.php` |

Note: ADR-0004 mentions PDPA once (line 47). The statement "no mention of UU PDP" is softened to: there is no UU PDP handling (classification, retention, consent, data subject rights) in the repo.

#### 11.3.3 Corrections to the initial claims of the assistant scan

The following claims came from the initial scan and were corrected after re-reading against the code. Suite documents that use them must follow these corrections.

| Initial claim | Correction | Evidence |
|---|---|---|
| `server.exe` is committed in the repo | Not tracked by git: `*.exe` is ignored (`.gitignore` line 7) and `git ls-files` does not include `server.exe`; the file only exists in the working tree. Cleanup applies to the working tree and build artefacts | `.gitignore` |
| Token stored twice in cookie and localStorage | Only cookies are filled (`auth_token` HttpOnly and `auth_user` not HttpOnly); `setSession` has no caller | `apps/web/src/lib/stores/auth.ts`, `apps/web/src/routes/(auth)/login/+page.server.ts` |
| Document 5 of 7 RPCs real | 4 of 7 are real (`UploadDocument`, `ListDocuments`, `GetDocumentMetadata`, `DeleteDocument`); `DownloadDocument` is also a placeholder | `services/document-service/internal/interfaces/grpc/document_service.go` |
| The `user.registered` event creates an employee shell | The consumer exists but cannot work (payload without `tenant_id`, random department/position ids, error acked then lost) | `services/employee-service/internal/infrastructure/nats_consumer.go` |
| Leave balance and overlap are not enforced | Balance is enforced in application code; overlap is not | `services/leave-service/internal/application/commands/create_leave_request.go` |
| `deploy/k8s/base` is empty | `deploy/k8s` only contains empty directories (`base`, `overlays/production`, `overlays/staging`); `git ls-files deploy` only includes four `deploy/docker` files | `deploy/docker/docker-compose.yml` |
| CD ignores some services | CD references Dockerfiles that do not exist (only `Dockerfile.base`, `services/auth-service/Dockerfile`, `services/employee-service/Dockerfile` exist); `document-service` is not in the CI/CD matrix; the smoke test calls `/health`, not `/api/health`; the CI build matrix includes `ai-service`, which has no entry point | `.github/workflows/cd-staging.yml`, `.github/workflows/ci-go.yml` |
| No mention of UU PDP at all | ADR-0004 mentions PDPA once; there is no UU PDP handling | `docs/adr/0004-defer-payroll-to-phase2.md` |
| CI does not test migrations | CI runs `up -> down -> up` and integration tests for auth, employee, leave only, as `hris_admin` | `.github/workflows/ci-go.yml` |

---

## 12. Closing Notes

1. **Document changes.** This document is the first version. Every HRIS code change that affects an FR in section 4 must update 11.1 and, if there is a new contradiction, 11.3 (STACK-SPEC 18 convention: FSD updated together with code).
2. **Estimates.** All effort estimates in 2.7 are `[INFERENCE]`. This document contains no readiness percentage.
3. **`tbd` values.** Values not found in the repo and awaiting an owner decision: tier/price/limit `hrs.employees.max` (Q-HRS-02), session TTL, account lockout threshold, and entitlement tolerance (Q-HRS-11), default work rules and leave policy (Q-HRS-05, Q-HRS-06), UU PDP retention and legal basis (Q-HRS-04, Q-HRS-14), availability/latency/RPO/RTO targets (Q-HRS-07), OTel observability (Q-HRS-08), LLM provider (Q-HRS-09), payroll legal review schedule (Q-HRS-10), production hosting (Q-HRS-12), SMTP provider/additional channels (Q-HRS-13).
4. **Work order.** Fixing the defects in 11.3.1 and the BRD 12 stabilisation gate precede stack alignment (D-HRS-09); HRIS alignment is scheduled after POS (stage S3) and one alignment is done at a time.
5. **Deferred modules.** Payroll, tax, recruitment, and performance have no FR; new FRs may only be written after the prerequisites of 4.09-4.12 are met and the related ADR is created.
6. **Limits of evidence.** All evidence comes from static reading; runtime failure claims are marked `Pending Verification` until run at the stabilisation gate.
