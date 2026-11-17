---
description: System-wide architecture rules for HRIS-Stery. Read before touching any file.
globs: ["**/*"]
---

# HRIS-Stery System Rules

## Architecture

```text
SvelteKit (apps/web) -> REST/JSON /api/v1 -> Go API (apps/api, one binary, port 8000)
                                              services are packages, called in-process
                                              -> PostgreSQL 18 (one schema per service, RLS)
                                              -> Dragonfly (refresh tokens, stream hris:events)
                                              -> S3-compatible storage (RustFS locally)
```

## Non-negotiable rules

- Layers: domain -> application -> infrastructure -> http.go. Never invert.
- A service never queries another service's schema. It calls the other package's exported application handler or consumes its events.
- Every table with `tenant_id` has PostgreSQL RLS enabled (ADR-0002).
- Every event consumer has a processed-events idempotency table, and events are published after the DB commit (ADR-0006).
- Business logic lives in `services/<name>/application/`. Never in `http.go` or repositories.
- Phase 2 features (payroll, tax rules, recruitment, performance) are blocked (ADR-0004). Their routes, if added, return 501.

## ADR index

- ADR-0001: grpc-gateway over gRPC-PHP (superseded by ADR-0005)
- ADR-0002: PostgreSQL RLS for tenant isolation (amended 2026-11-08)
- ADR-0003: NATS JetStream event backbone (superseded by ADR-0006)
- ADR-0004: Defer payroll, recruitment and performance to Phase 2
- ADR-0005: Single-module Go API on the SELA stack
- ADR-0006: Dragonfly Streams as the event backbone (ADR-0003 superseded)

## Event envelope (every publish to `hris:events`)

```json
{ "event_id": "uuid-v7", "event_type": "hris.domain.entity.verb",
  "schema_version": 1, "tenant_id": "...", "actor_id": "...",
  "occurred_at": "RFC 3339", "payload": {} }
```

## Comment policy

Write comments only when the reason is non-obvious. Never explain what the code does.
