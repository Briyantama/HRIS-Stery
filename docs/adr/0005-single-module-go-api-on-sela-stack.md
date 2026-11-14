# ADR-0005: Single-Module Go API on the SELA Stack

**Status:** Accepted  
**Date:** 2026-11-08  
**Deciders:** Owner (decisions D1, D2, D5 of intent `261108-stack-align-sela`)  
**Supersedes:** ADR-0001

---

## Context

Until 2026-11-08 HRIS-Stery ran as a Laravel 11 gateway in front of seven Go services. Each service had its own `go.mod`, a gRPC server, grpc-gateway HTTP handlers generated from `proto/` with buf, golang-migrate files, zap logging and OpenTelemetry. The services shared code through `services/_shared` and a `go.work` file. PgBouncer sat in front of Postgres.

The owner runs a second product, SELA, on a smaller stack: one Go module, one binary, `net/http` ServeMux routes, `log/slog`, goose migrations embedded in the binary, Dragonfly, an S3 SDK client and a `{success, data, error}` envelope. POS-Stery is moving to the same stack. Keeping three stacks means three sets of tooling, CI jobs and review rules for one developer.

Nothing outside the repo calls the HRIS services over gRPC. The frontend only talks to the gateway over REST.

### Option A: Keep the microservice layout

Independent deployables and scaling per service. Costs: a PHP runtime, buf and protoc in the toolchain, eight images in CI and Kubernetes, network hops for calls that never leave one machine.

### Option B: One Go module, services as packages (selected)

`apps/api` becomes the only Go module (`github.com/Briyantama/HRIS-Stery`). Each former service becomes a package under `apps/api/services/<name>` that keeps its domain, application and infrastructure layers and adds an `http.go` adapter. `cmd/api` wires every package into one `http.Server`; `cmd/migrate` applies the goose files.

### Option C: One module, keep gRPC between packages

Keeps the proto contracts but adds a network boundary inside one process. No consumer needs it.

---

## Decision

**Run HRIS-Stery as one Go module with one API binary on the SELA stack.**

- The Laravel gateway, gRPC, protobuf and buf, grpc-gateway, `go.work` and PgBouncer are removed.
- Routes are Go 1.22 `net/http` ServeMux patterns (`"VERB /api/v1/..."`) registered by each package's `HTTPHandler.Register`. The old gateway paths are kept; other paths come from the former proto `google.api.http` annotations, prefixed with `/api`.
- Authentication moves into the API: RS256 JWT verified locally with the public key (claims `sub`, `tid`, `roles`), roles checked per route. Tenant and actor ids come from the token, never from the request body.
- Calls between services are in-process function calls to the other package's application handlers.
- Logging is `log/slog`. OpenTelemetry, Prometheus and Jaeger are removed.
- Migrations are goose files in `apps/api/internal/db/migrations`, one number range per service, embedded in the binary.
- Object storage uses the AWS S3 SDK (RustFS locally) instead of minio-go.
- Postgres 18, Dragonfly for Redis-protocol state and the event stream (ADR-0006).

### Shared kit (decision D5)

Infrastructure that HRIS-Stery and POS-Stery both need lives in a sibling repository, `stery-kit` (`git@github.com:Briyantama/stery-kit.git`), checked out next to this repo:

- Go module `github.com/Briyantama/stery-kit/go`: `httpx`, `health`, `objstore`, `migrate`, `testdb`, `pgtenant`, `events`, `authn`, `ratelimit`.
- npm package `@stery/kit` in `stery-kit/web`.

`apps/api/go.mod` uses `replace github.com/Briyantama/stery-kit/go => ../../../stery-kit/go`; the frontend uses `file:../../../stery-kit/web`. Docker builds of `apps/api` use the parent directory of the repo as build context so both trees are visible, and CI checks out `stery-kit` into a sibling directory. Product-only code (`internal/config`, migrations, service packages) stays in this repo.

---

## Consequences

### Positive
- One binary, one image, one CI job for Go. Local start is `make up`.
- No PHP, buf or protoc in the toolchain.
- Calls between services cannot fail on the network and need no service tokens.
- The HRIS gateway TODOs that read the tenant from the request body and used `ActorID: "system"` are fixed, because every handler reads both from the token.
- Fixes to shared infrastructure land once in `stery-kit` for both products.

### Negative
- Services can no longer be scaled or deployed independently. A crash or a slow handler affects every route.
- No distributed tracing or Prometheus metrics. Diagnosis relies on `slog` logs with `request_id` and `tenant_id`.
- A build needs the `stery-kit` checkout next to the repo. Until the kit repo is pushed, CI fails at checkout.
- A boundary between packages is a code-review rule, not a process boundary. A package can import another package's infrastructure if review misses it.

### Neutral
- Each service keeps its own Postgres schema and RLS policies (ADR-0002).
- Repositories keep `pgx/v5`, `pgxpool` and the existing sqlc output.

---

## Compliance

- No new Go module, gRPC server or proto file without a new ADR.
- A package never queries another package's schema; it calls the other package's exported application handler or consumes its events.
- Every route is either listed as public (`POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`, tenant registration, `GET /healthz`) or wrapped by `authn.Authenticate`.
- Handlers answer only through `httpx.WriteSuccess` / `httpx.WriteError`.
