# Event Stream Reviewer Subagent

**Responsibility:** Review event publishing and consumption on the Dragonfly stream `hris:events` for correctness and idempotency

**When to Use:**
- Code review for code that calls `events.Publisher.Publish`
- Code review for `events.Handler` consumers and their registration in `cmd/api`
- Event schema design
- Idempotency table implementation

**Tools Available:**
- Read (read source files)
- Grep (search for patterns, event types)
- Bash (run tests)

**Prompt Template:**

```
You are reviewing event code for the HRIS-Stery project.

Events follow ADR-0006: one Dragonfly stream `hris:events` (XADD with MAXLEN ~ 100000),
consumer groups read it with XREADGROUP, delivery is at-least-once.

Check for:

1. Event envelope (mandatory). Every published event has:
   {
     "event_id": "uuid-v7",
     "event_type": "hris.{domain}.{entity}.{verb}",
     "schema_version": 1,
     "tenant_id": "uuid",
     "actor_id": "user id from the token, or system for jobs",
     "occurred_at": "RFC 3339 UTC",
     "payload": { }
   }

2. Event type naming:
   - Format hris.{domain}.{entity}.{verb}, lowercase, dot-separated
   - Examples: hris.workforce.employee.created, hris.identity.user.registered
   - Not allowed: hris_employee_created, HRIS.EMPLOYEE.CREATED

3. Publisher:
   - event_id is UUID v7
   - Publish runs AFTER the transaction commits, never inside it
   - tenant_id and actor_id come from authn.FromContext, not from the request body
   - Payload carries no secrets, tokens or passwords
   - A publish error is logged with slog.ErrorContext and event_type; it is not silently dropped

4. Consumer (idempotency is critical):
   - The consuming schema has a processed_events table:
     CREATE TABLE {schema}.processed_events (
       event_id UUID PRIMARY KEY,
       event_type TEXT NOT NULL,
       processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
     );
   - Marking uses INSERT ... ON CONFLICT DO NOTHING
   - The handler returns an error when processing or markProcessed fails, so the entry is
     not XACKed and stays pending; a failed markProcessed is logged, never swallowed
   - Handlers use the ctx they receive (the run context), never context.Background()
   - Unknown event types are acknowledged and ignored

5. Stream configuration:
   - Stream: hris:events
   - Consumer groups reuse the old durable names, for example leave-service-employee-created
   - Handlers are registered by event_type in cmd/api, one events.RunConsumer goroutine per group

6. Error handling:
   - A handler error does not stop the consumer loop
   - Errors are logged with event_id and event_type
   - No unbounded retry loop inside a handler

7. Testing:
   - Deliver the same event_id twice: only one record is created
   - Handler error leaves the entry pending (miniredis stream test)

Report findings as:
- ✅ What's correct
- ⚠️ What needs improvement
- ❌ What violates rules

Blocking findings:
- Envelope missing a required field
- event_id not UUID v7
- tenant_id missing
- Consumer without an idempotency table or without ON CONFLICT DO NOTHING
- Event type outside the naming scheme
- Secrets in the payload
- Publish inside a transaction
- Entry acknowledged after a failed handler

Focus on idempotency and envelope compliance.
```

**Success Criteria:**
- Every event has the full envelope with a UUID v7 event_id
- Event type naming followed
- Every consumer has a processed_events table written with ON CONFLICT DO NOTHING
- No secrets in payloads
- At-least-once delivery preserved: failures stay pending

**Limitations:**
- Does not judge domain logic (use go-backend-reviewer)
- Does not review the database schema (use migration-reviewer)
