# ADR-0006: Dragonfly Streams Replace NATS JetStream

**Status:** Accepted  
**Date:** 2026-11-08  
**Deciders:** Owner (decision D3 of intent `261108-stack-align-sela`)  
**Supersedes:** ADR-0003

---

## Context

ADR-0003 chose NATS JetStream for domain events between services. After ADR-0005 every service runs inside one binary, but events are still needed: the employee shell is created from `hris.identity.user.registered`, leave balances are seeded from `hris.workforce.employee.created`, and notification and audit react to events from several domains. These reactions must survive a restart and must not run inside the publisher's transaction.

The stack already runs Dragonfly for refresh tokens and revocation keys. Dragonfly speaks the Redis protocol, including streams (`XADD`, `XREADGROUP`, `XACK`), and `go-redis/v9` is already a dependency. SELA uses the same store.

### Option A: Keep NATS JetStream

Durable file storage and replay. Costs one more server in every environment, the `nats.go` client and per-service ACL configuration for a single process.

### Option B: In-process channels

No extra infrastructure. Events are lost on a crash between commit and handling, and there is no redelivery.

### Option C: Dragonfly Streams (selected)

One stream, consumer groups with pending-entry tracking, no new server.

---

## Decision

**Publish domain events to the Dragonfly stream `hris:events` and consume them with consumer groups.**

- Publisher: `XADD hris:events MAXLEN ~ 100000 * type <event_type> data <envelope JSON>`, through `events.Publisher.Publish` from `stery-kit`.
- Consumer: `events.RunConsumer` creates the group (`XGROUP CREATE ... $ MKSTREAM`), first re-reads its own pending entries (ID `0`), then reads new entries (`>`, `BLOCK 5s`, `COUNT 100`) and dispatches on the `type` field.
- An entry is acknowledged only when the handler returns nil. A handler error is logged with `slog.ErrorContext` and the entry stays pending until the next restart. Unknown types are acknowledged.
- Consumer groups keep the old JetStream durable names (for example `leave-service-employee-created`, `notification-auth-consumer`, `audit-identity-consumer`). Each group runs as one goroutine started by `cmd/api` with the root context.

Unchanged from ADR-0003:

- Envelope fields `event_id` (UUID v7), `event_type`, `schema_version`, `tenant_id`, `actor_id`, `occurred_at`, `payload`.
- Event type strings `hris.{domain}.{entity}.{verb}`; the old subjects become the `type` field as they were.
- Every consumer writes its schema's `processed_events` table with `INSERT ... ON CONFLICT DO NOTHING`.
- Events are published after the database commit, never inside the transaction.

---

## Consequences

### Positive
- One less server to run, configure and secure.
- No new dependency: `go-redis/v9` was already used.
- A failed handler keeps its entry pending, so the at-least-once contract of ADR-0003 holds.

### Negative
- Durability is weaker. JetStream wrote to disk with replicas; Dragonfly holds the stream in memory and persists it through snapshots. Entries written after the last snapshot are lost if Dragonfly crashes. `MAXLEN ~ 100000` trims old entries, so replay is limited to the most recent ones.
- Pending entries are retried only after a restart; there is no ack timeout or dead-letter subject.
- No per-service ACLs on publish or subscribe. All packages share one client.

### Neutral
- Moving back to a broker later stays possible: envelope and type strings are transport-agnostic, and only `stery-kit/go/events` would change.

---

## Compliance

- Every published event carries the full envelope with a UUID v7 `event_id`.
- Every consumer has a `processed_events` table and returns an error when processing or marking fails.
- Handlers use the context they receive, never `context.Background()`.
- New event types follow `hris.{domain}.{entity}.{verb}` and are added to the list in `.claude/skills/event-review/SKILL.md`.
