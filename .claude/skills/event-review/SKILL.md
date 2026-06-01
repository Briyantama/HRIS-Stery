# Event Review Skill

**Trigger:** Before publishing or consuming NATS events

**Purpose:** Ensure events follow the standard envelope and idempotency patterns

## Event Envelope (Mandatory)

Every event published to NATS must have this structure:

```json
{
  "event_id": "uuid-v7",
  "event_type": "hris.{domain}.{entity}.{verb}",
  "schema_version": 1,
  "tenant_id": "uuid",
  "actor_id": "user-id or system-id",
  "occurred_at": "ISO-8601 timestamp",
  "payload": { /* domain-specific data */ }
}
```

## Publishing Checklist

- [ ] Event has unique `event_id` (UUID v7, not v4)
- [ ] `event_type` follows taxonomy: `hris.{domain}.{entity}.{verb}`
- [ ] `schema_version` is documented (for compatibility)
- [ ] `tenant_id` included (for tenant isolation)
- [ ] `actor_id` identifies who triggered the event
- [ ] `occurred_at` is ISO-8601 UTC timestamp
- [ ] `payload` contains only necessary data
- [ ] Event published AFTER transaction commits (transactional outbox pattern if needed)
- [ ] No secrets in payload

## Consumer Checklist (Idempotency)

- [ ] Consumer has `idempotency_processed_events` table:
  ```sql
  CREATE TABLE {schema}.processed_events (
    event_id UUID PRIMARY KEY,
    event_type TEXT NOT NULL,
    processed_at TIMESTAMPTZ DEFAULT now()
  );
  ```
- [ ] Before processing, check: `SELECT 1 FROM processed_events WHERE event_id = $1`
- [ ] After processing, insert: `INSERT INTO processed_events (event_id, event_type) VALUES ($1, $2) ON CONFLICT DO NOTHING`
- [ ] Consumer handles JetStream at-least-once delivery
- [ ] Errors logged with event_id for audit

## Subject Taxonomy

```
hris.{domain}.{entity}.{verb}

Examples:
- hris.identity.user.registered      (auth-service publishes)
- hris.workforce.employee.created    (employee-service publishes)
- hris.workforce.employee.terminated (employee-service publishes)
- hris.attendance.record.clocked-in  (attendance-service publishes)
- hris.leave.request.submitted       (leave-service publishes)
```

## JetStream Configuration

- Stream name: `HRIS_EVENTS`
- Consumer names: `{service-name}-{subject-slug}`
- Retention: Long-term (for audit)
- Delivery policy: At-least-once (requires idempotency)

## Examples

**Trigger:** "Publish employee.created event"
→ Ensure UUID v7 event_id, hris.workforce.employee.created subject, tenant_id in payload

**Trigger:** "Consume user.registered events"
→ Implement idempotency table, check event_id before processing

**Trigger:** "Add new event type"
→ Update subject taxonomy and document in ADR-0003
