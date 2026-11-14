# Event Review Skill

**Trigger:** Before publishing or consuming events on the Dragonfly stream `hris:events`

**Purpose:** Ensure events follow the standard envelope and idempotency patterns

## Event Envelope (Mandatory)

Every event published with `events.Publisher.Publish` must have this structure:

```json
{
  "event_id": "uuid-v7",
  "event_type": "hris.{domain}.{entity}.{verb}",
  "schema_version": 1,
  "tenant_id": "uuid",
  "actor_id": "user-id or system-id",
  "occurred_at": "RFC 3339 UTC timestamp",
  "payload": { /* domain-specific data */ }
}
```

## Publishing Checklist

- [ ] Event has unique `event_id` (UUID v7, not v4)
- [ ] `event_type` follows taxonomy: `hris.{domain}.{entity}.{verb}`
- [ ] `schema_version` is documented (for compatibility)
- [ ] `tenant_id` included (for tenant isolation)
- [ ] `actor_id` identifies who triggered the event
- [ ] `occurred_at` is an RFC 3339 UTC timestamp
- [ ] `payload` contains only necessary data
- [ ] Event published AFTER transaction commits (transactional outbox pattern if needed)
- [ ] No secrets in payload

## Consumer Checklist (Idempotency)

- [ ] Consumer schema has a `processed_events` table:
  ```sql
  CREATE TABLE {schema}.processed_events (
    event_id UUID PRIMARY KEY,
    event_type TEXT NOT NULL,
    processed_at TIMESTAMPTZ DEFAULT now()
  );
  ```
- [ ] Before processing, check: `SELECT 1 FROM processed_events WHERE event_id = $1`
- [ ] After processing, insert: `INSERT INTO processed_events (event_id, event_type) VALUES ($1, $2) ON CONFLICT DO NOTHING`
- [ ] Handler returns an error when processing or marking fails, so the entry is not acknowledged and stays pending
- [ ] Handler uses the ctx it receives, never `context.Background()`
- [ ] Errors logged with `slog.ErrorContext` and event_id

## Subject Taxonomy

```
hris.{domain}.{entity}.{verb}

Examples:
- hris.identity.user.registered          (auth publishes)
- hris.workforce.employee.created        (employee publishes)
- hris.workforce.employee.terminated     (employee publishes)
- hris.attendance.record.checked_in      (attendance publishes)
- hris.operations.leave.requested        (leave publishes)
```

## Stream configuration

- Stream: `hris:events`, `XADD ... MAXLEN ~ 100000`
- Consumer groups keep the old durable names (e.g. `leave-service-employee-created`)
- Delivery: at-least-once (requires idempotency); pending entries are re-read after a restart

## Examples

**Trigger:** "Publish employee.created event"
→ Ensure UUID v7 event_id, hris.workforce.employee.created subject, tenant_id in payload

**Trigger:** "Consume user.registered events"
→ Implement idempotency table, check event_id before processing

**Trigger:** "Add new event type"
→ Follow the naming scheme and list the type in ADR-0006
