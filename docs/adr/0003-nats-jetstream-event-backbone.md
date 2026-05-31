# ADR-0003: NATS JetStream as the Async Event Backbone

**Status:** Accepted  
**Date:** 2026-05-30  
**Deciders:** Architecture Team  
**Supersedes:** —

---

## Context

The architecture requires an async event backbone for:
1. Decoupled inter-service communication (e.g., `leave.approved` → notification-service)
2. Audit event ingestion (all services → audit-service)
3. AI pipeline triggers (talent events → ai-service)
4. Real-time notifications to the frontend

The alternatives evaluated were Apache Kafka, RabbitMQ, and NATS JetStream.

### Option A: Apache Kafka

Kafka is the industry standard for high-throughput event streaming. It is the right choice for systems processing millions of events per day with strict ordering guarantees across partitions.

Problems for this project:
- Operational complexity is significant: ZooKeeper (or KRaft), brokers, topic management, consumer group coordination.
- Minimum viable Kafka cluster requires 3 brokers for production-grade durability.
- Kafka's Kubernetes footprint is large (StatefulSets, persistent volumes, JVM memory).
- Kafka's Go client (`confluent-kafka-go`) requires CGo. `sarama` or `franz-go` are pure Go but have different trade-offs.
- At the scale of this system (target: 100k users, multi-tenant HRIS), Kafka's operational overhead is unjustified.

### Option B: RabbitMQ

RabbitMQ is a mature message broker well-suited for work queues and request/reply patterns.

Problems:
- RabbitMQ is not a log/stream — messages are consumed and gone. Replay is not supported.
- The AMQP protocol is more complex to model for event-driven architectures.
- RabbitMQ's Go support is good but the mental model differs from pub/sub streams.
- Does not provide built-in key-value store or object store (NATS does, via JetStream KV and Object Store).

### Option C: NATS JetStream (selected)

NATS is a high-performance, Go-native messaging system. JetStream adds persistence, replay, at-least-once delivery, and consumer management on top of the core NATS pub/sub.

Advantages:
- Single binary. No external dependencies (no ZooKeeper, no Erlang runtime).
- Native Go client (`nats.go`) — first-class support, idiomatic API.
- JetStream provides: persistent streams, durable consumers, message replay, deduplication window, key-value store, object store.
- NATS JetStream clustering for production HA requires 3 nodes — operationally simpler than Kafka.
- Subject-based addressing matches the domain event taxonomy naturally.
- NATS also handles real-time pub/sub (core NATS) for low-latency notification delivery.
- Built-in ACL support for service-level publish/subscribe permissions.

---

## Decision

**Use NATS JetStream as the sole async event backbone.**

NATS serves two roles:
1. **JetStream** — durable, persistent event stream for domain events (at-least-once delivery)
2. **Core NATS** — ephemeral pub/sub for real-time frontend notifications (fire-and-forget)

---

## Stream Configuration

```
Stream Name:    HRIS_EVENTS
Subjects:       hris.>
Storage:        File
Retention:      Limits (7 days default, configurable per tenant)
Max Messages:   -1 (unlimited within retention)
Replicas:       1 (dev) / 3 (staging, production)
Dedup Window:   2 minutes (prevents JetStream-level duplicates)
```

---

## Subject Taxonomy

```
hris.{domain}.{entity}.{verb}

Domain values:   identity | workforce | operations | compensation | talent | intelligence | audit
Entity values:   user | tenant | employee | department | attendance | leave | payroll | job | candidate | analysis
Verb values:     created | updated | deleted | requested | approved | rejected | completed | failed

Examples:
  hris.identity.user.registered
  hris.workforce.employee.created
  hris.workforce.employee.terminated
  hris.operations.leave.requested
  hris.operations.leave.approved
  hris.operations.attendance.checked_in
  hris.audit.event.recorded
```

---

## Event Envelope (Mandatory Schema)

All events published to `hris.>` must conform to this envelope. Schema version is used for forward-compatibility.

```json
{
  "event_id":       "019547f2-3a1b-7e2c-8d4f-0a1b2c3d4e5f",
  "event_type":     "hris.workforce.employee.created",
  "schema_version": 1,
  "tenant_id":      "550e8400-e29b-41d4-a716-446655440000",
  "actor_id":       "user-uuid-or-system",
  "occurred_at":    "2026-05-30T10:00:00Z",
  "payload":        { ... }
}
```

`event_id` must be UUID v7 (time-ordered). This provides natural chronological ordering and enables deduplication.

---

## Idempotency Contract (Mandatory for All Consumers)

JetStream guarantees at-least-once delivery. Consumers must be idempotent.

Every consumer service maintains:

```sql
CREATE TABLE {schema}.processed_events (
    event_id   UUID PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Consumer processing pattern:

```go
func (c *Consumer) handle(msg *nats.Msg) {
    var envelope EventEnvelope
    json.Unmarshal(msg.Data, &envelope)

    _, err := db.Exec(
        `INSERT INTO processed_events (event_id) VALUES ($1) ON CONFLICT DO NOTHING`,
        envelope.EventID,
    )
    if err != nil || rowsAffected == 0 {
        msg.Ack() // already processed — ack and skip
        return
    }

    // process event
    if err := c.usecase.Handle(ctx, envelope); err != nil {
        msg.Nak() // will be redelivered
        return
    }
    msg.Ack()
}
```

---

## NATS ACL Configuration

Each service has a dedicated NATS user with minimal permissions:

| Service | Publish | Subscribe |
|---|---|---|
| auth-service | `hris.identity.>` | — |
| employee-service | `hris.workforce.>` | `hris.identity.user.registered` |
| attendance-service | `hris.operations.attendance.>` | `hris.workforce.employee.created` |
| leave-service | `hris.operations.leave.>` | `hris.workforce.employee.created` |
| notification-service | — | `hris.operations.leave.>`, `hris.compensation.payroll.>` |
| audit-service | `hris.audit.>` | `hris.>` (all) |
| ai-service | `hris.intelligence.>` | `hris.talent.>` |

---

## Real-Time Notifications

For real-time frontend notifications (e.g., "Your leave was approved"), the notification-service:
1. Subscribes to relevant JetStream subjects.
2. Publishes to a core NATS ephemeral subject: `realtime.{tenant_id}.{user_id}`.
3. The SvelteKit frontend connects to a Server-Sent Events (SSE) endpoint on the Laravel gateway.
4. Laravel gateway subscribes to `realtime.{tenant_id}.{user_id}` for the duration of the SSE connection.

This avoids WebSocket infrastructure complexity in the MVP.

---

## Consequences

### Positive
- Single binary, simple deployment. NATS server in Docker Compose is one line.
- Native Go client is idiomatic and well-maintained.
- JetStream deduplication window provides a first line of defense against duplicate events.
- Subject ACLs enforce service boundaries at the messaging layer.
- Message replay enables audit-service to rebuild state from events if needed.

### Negative
- NATS JetStream is less battle-tested at extreme scale than Kafka. Acceptable for this use case.
- NATS does not have a managed cloud offering comparable to Confluent Cloud or Amazon MSK. Self-hosted only (or NGS for cloud-hosted NATS).
- Consumer lag monitoring requires NATS monitoring endpoints or the `nats` CLI — less mature tooling than Kafka's ecosystem.

### Neutral
- Migrating from NATS to Kafka later (if scale requires it) is possible: the event envelope schema and subject taxonomy are transport-agnostic. A migration adapter can bridge the two systems during cutover.

---

## Compliance

- All events published to `hris.>` must include the mandatory envelope fields.
- All consumer services must implement the idempotency table pattern.
- NATS ACLs must be reviewed in every PR that adds a new consumer or publisher.
- `event_id` must be UUID v7. Reject events at the consumer level if `event_id` is missing.
- Never publish events inside a database transaction. Use the transactional outbox pattern if ordering between DB write and event publish is required.
