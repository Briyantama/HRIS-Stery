---
description: System-wide architecture rules for HRIS-Stery. Read before touching any file.
globs: ["**/*"]
---

# HRIS-Stery System Rules

## Architecture Hierarchy

```
SvelteKit → REST/JSON → Laravel Gateway → HTTP/JSON (grpc-gateway) → Go Services
                                                                   ↔ gRPC (service-to-service)
                                                                   → NATS JetStream (events)
```

## Non-Negotiable Rules

- Clean Architecture: domain → application → infrastructure → interfaces. Never invert.
- Services never query another service's database schema. Cross-service data goes through gRPC only.
- Every table with `tenant_id` must have PostgreSQL RLS enabled (see ADR-0002).
- Every NATS consumer must implement the idempotency table pattern (see ADR-0003).
- Business logic lives in `internal/application/`. Never in handlers, controllers, or repositories.
- Phase 2 features (payroll, recruitment, performance) are blocked — see ADR-0004.

## ADR Index

- ADR-0001: grpc-gateway over gRPC-PHP
- ADR-0002: PostgreSQL RLS for tenant isolation
- ADR-0003: NATS JetStream event backbone
- ADR-0004: Defer payroll/recruitment/performance to Phase 2

## Event Envelope (mandatory for all NATS publishes)

```json
{ "event_id": "uuid-v7", "event_type": "hris.domain.entity.verb",
  "schema_version": 1, "tenant_id": "...", "actor_id": "...",
  "occurred_at": "ISO-8601", "payload": {} }
```

## Comment Policy

Write comments only when the WHY is non-obvious. Never explain WHAT the code does.
