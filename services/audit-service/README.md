# Audit Service

Immutable, append-only audit trail for HRIS operations.

## Purpose

Captures WHO, WHAT, WHEN, WHERE for all business operations:
- Authentication events (login, logout, password change)
- Employee lifecycle (hired, terminated, department changes)
- Leave actions (requested, approved, rejected)
- Notification delivery

## Architecture

- **Domain:** Immutable AuditEntry aggregate
- **Application:** CQRS (RecordAudit command, QueryAuditTrail query)
- **Infrastructure:** PostgreSQL (INSERT-ONLY) + NATS consumers (event-driven)
- **Interfaces:** gRPC with grpc-gateway

## Key Properties

- ✅ **Immutable:** No update/delete
- ✅ **Append-Only:** INSERT-only database schema
- ✅ **Event-Driven:** NATS consumers populate entries
- ✅ **Idempotent:** Duplicate events = single entry
- ✅ **Queryable:** Filter by actor, action, resource, date range
- ✅ **Compliant:** Full audit trail for audits and forensics

## Development

See AUDIT_SERVICE_IMPLEMENTATION.md for implementation roadmap.
