# Code Review Decision Tree

**Use this to decide which Skill or Subagent to apply to a task.**

---

## Step 1: What are you doing?

### 🔵 Writing new code (feature or change)

→ Go to **Step 2: Which service/layer?**

### 🟢 Reviewing existing code

→ Go to **Step 3: What type of file?**

### 🟡 Checking architecture/design

→ Use **architecture-review** skill (Go)  
or discuss with team

---

## Step 2: Which service/layer? (For New Code)

### Go Service (services/*/internal/)

1. **First:** Read relevant ADR in `docs/adr/`
   - Trigger: **adr-review** skill
   
2. **Then:** Check proto contract
   - Trigger: **proto-review** skill
   - Verify `proto/hris/{domain}/v1/{domain}.proto` exists
   
3. **Implement:** Domain → Application → Infrastructure → Interfaces
   - Domain: No I/O, pure logic
   - Application: Orchestration, use cases
   - Infrastructure: Repositories, NATS, caching
   - Interfaces: gRPC handlers (thin)
   
4. **If touching NATS events:**
   - Trigger: **event-review** skill
   - Verify event envelope and idempotency
   
5. **Before commit:** Code review
   - Trigger: **go-backend-reviewer** subagent

### Laravel Gateway (gateway/app/)

1. **Controller:** Keep <20 lines of logic
   - Trigger: **laravel-gateway-reviewer** subagent
   - Check: No DB access, no domain logic
   
2. **Service class:** Calls Go services via HTTP/JSON
   - Check: No gRPC-PHP, only grpc-gateway endpoints
   
3. **Form Request:** Validation logic
   - Check: Form Request used for every POST/PUT

### SvelteKit Frontend (frontend/src/)

1. **Route:** Use TanStack Query for server state
   - Check: No raw fetch(), use createQuery/createMutation
   
2. **Component:** UI-only logic
   - Check: No business logic

### Proto (proto/hris/{domain}/v1/)

1. **New RPC:** Add request/response messages first
   - Trigger: **proto-review** skill
   - Run: `buf lint` and `buf breaking`
   
2. **Field addition:** Safe (backward compatible)
   
3. **Field removal:** BREAKING — needs v2 package

### Database Migrations (services/*/migrations/)

1. **Create both:** `NNN_description.up.sql` + `NNN_description.down.sql`
   - Trigger: **migration-review** skill
   - Verify: RLS on tenant-scoped tables
   - Verify: Down fully reverses up

---

## Step 3: What type of file? (For Code Review)

```
┌─────────────────────────────────────┐
│ Type of File Being Changed          │
└─────────────────────────────────────┘
         │
         ├─ .go (Go code)
         │    └─ Subagent: go-backend-reviewer
         │
         ├─ .php (Laravel)
         │    └─ Subagent: laravel-gateway-reviewer
         │
         ├─ .svelte / .ts (Frontend)
         │    └─ Manual review (or frontend-reviewer if available)
         │
         ├─ .sql (Migrations)
         │    └─ Subagent: migration-reviewer
         │
         ├─ .proto (gRPC proto)
         │    └─ Subagent: proto-reviewer
         │
         └─ .md (Documentation)
              └─ Manual review
```

---

## Step 4: Specific Review Paths

### "I'm adding a new RPC endpoint"

1. Proto first:
   - Create `{Verb}{Entity}Request` + `{Verb}{Entity}Response`
   - Add HTTP annotation
   - Trigger: **proto-review** skill
   - Run: `cd proto && buf lint`

2. Handler:
   - Implement gRPC handler (thin, delegates to application)
   - Trigger: **go-backend-reviewer** subagent
   - Check: <20 lines, uses application handler

3. Application:
   - Implement command/query handler (business logic)
   - Trigger: **go-backend-reviewer** subagent
   - Check: Uses repositories, not raw DB

4. Infrastructure:
   - Implement repository methods (data access)
   - Trigger: **go-backend-reviewer** subagent
   - Check: Uses WithTenantTx for RLS

### "I'm adding a database table"

1. Write migration:
   - `NNN_description.up.sql` + `NNN_description.down.sql`
   - Trigger: **migration-review** skill
   - Check: RLS on tenant-scoped tables, down reverses up

2. Create repositories:
   - Implement repository interface
   - Trigger: **go-backend-reviewer** subagent
   - Check: RLS context, error handling

### "I'm publishing an event"

1. Check event structure:
   - Trigger: **event-review** skill
   - Verify: UUID v7, tenant_id, subject taxonomy

2. Implement publisher:
   - Trigger: **go-backend-reviewer** subagent
   - Check: Publishes AFTER transaction

3. Test:
   - Verify event in NATS
   - Verify idempotency (publish same event twice)

### "I'm consuming an event"

1. Check consumer structure:
   - Trigger: **event-review** skill
   - Verify: processed_events table, ON CONFLICT DO NOTHING

2. Implement consumer:
   - Trigger: **go-backend-reviewer** subagent
   - Check: Idempotency, error logging, event_id tracking

3. Test:
   - Publish duplicate events
   - Verify only one processing

---

## Step 5: Decision Summary

| Task | Primary Skill | Primary Subagent | Check First |
|------|---------------|------------------|-------------|
| New service feature | adr-review + proto-review | go-backend-reviewer | ADR + proto contract |
| New database table | migration-review | - | Both up & down migrations |
| New RPC method | proto-review | go-backend-reviewer + proto-reviewer | Proto file exists |
| Event publisher | event-review | go-backend-reviewer | Event envelope structure |
| Event consumer | event-review | go-backend-reviewer | Idempotency table |
| Laravel controller | - | laravel-gateway-reviewer | <20 lines, no logic |
| Laravel service | - | laravel-gateway-reviewer | Calls Go via HTTP only |
| SvelteKit component | - | Manual or frontend-reviewer | No business logic |
| Migration | migration-review | migration-reviewer | Both up & down files |

---

## Step 6: Quick Command Reference

```bash
# Before coding a new service
adr-review        # Read relevant ADR
proto-review      # Check proto exists

# Before committing Go code
go-backend-reviewer  # Full architecture review

# Before committing migrations
migration-review     # RLS + reversibility

# Before committing events
event-review         # Envelope + idempotency

# Before committing Laravel
laravel-gateway-reviewer  # Controller size + logic separation

# Before merging proto
cd proto && buf lint
cd proto && buf breaking --against .git#branch=main
```

---

## Repository Rules That Block Changes

These violate CLAUDE.md and **cannot be merged**:

- ❌ Business logic in gRPC handlers
- ❌ Business logic in Laravel controllers (>20 lines total)
- ❌ `interface{}` or `any` in Go domain types
- ❌ Direct database access from another service
- ❌ Migration without reversible down file
- ❌ Event without UUID v7 event_id
- ❌ Consumer without idempotency table
- ❌ RLS missing on tenant-scoped table
- ❌ No ADR for architectural decision
- ❌ No proto contract before handler
- ❌ Phase 2 features (payroll, recruitment, performance)

---

## When in Doubt

1. **Ask:** "Does this change cross service boundaries?" → Verify database isolation
2. **Ask:** "Is there business logic?" → Check it's in application layer, not controller/handler
3. **Ask:** "Does this need a migration?" → Verify both up & down files
4. **Ask:** "Does this publish events?" → Verify envelope and idempotency
5. **Ask:** "Is there an ADR for this?" → Read it before implementing
6. **Ask:** "Is this Phase 2?" → Check CLAUDE.md forbidden list

---

**Default path:** ADR-review → Proto-review → Architecture-review → Subagent-specific-review
