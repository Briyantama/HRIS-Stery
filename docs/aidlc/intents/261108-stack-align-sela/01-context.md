# 01 - Context

## Starting point

HRIS-Stery before 2026-11-08: Laravel 11 gateway (Sanctum) calling seven Go gRPC services through grpc-gateway, NATS JetStream, golang-migrate per service, zap, OpenTelemetry, Prometheus, MinIO, PgBouncer. Frontends: `apps/web`, behind Laravel gateway `apps/api-gateway`.

Reference stack: `SELA/apps/api` (one module, `cmd/api`, `cmd/migrate`, `internal/{config,httpx,health,db,testdb,objstore}`, `services/<name>`).

## Invariants that must not weaken

- RLS on every `tenant_id` table; `hris_app` has no BYPASSRLS
- `audit.audit_events` is insert-only
- every consumer has a processed-events table written with `INSERT ... ON CONFLICT DO NOTHING`; a failed mark is logged and the entry stays un-acked
- publish after commit, never inside the transaction
- JWT RS256 claims `sub`, `tid`, `roles[]`, verified locally with the public key
- Phase 2 services (payroll, tax-rules, recruitment, performance) stay unimplemented
- no tokens, passwords or PII in logs

## Behavior to preserve

Every invariant above, every existing domain and application unit test, the event subject strings and envelope fields, the JWT claims, and every route path the frontends call (with `/v1` inserted where the route table in `05-units-of-work.md` says so).
