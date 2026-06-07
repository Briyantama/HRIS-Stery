# Critical Fixes Implementation Plan

**Date:** 2026-06-07  
**Scope:** Week 1 - Fix 5 critical issues  
**Priority:** CRITICAL - Must fix before production

---

## ISSUE 1: NATS Idempotency Silent Failures

**Problem:** `markProcessed()` errors ignored → duplicate event processing

**Files Affected:**

- `services/audit-service/internal/infrastructure/nats/workforce_consumer.go:132`
- `services/audit-service/internal/infrastructure/nats/identity_consumer.go:138`
- `services/audit-service/internal/infrastructure/nats/operations_consumer.go` (similar pattern)
- `services/audit-service/internal/infrastructure/nats/notification_consumer.go` (similar pattern)

**Current Code (Line 132 in workforce_consumer.go):**

```go
func (c *WorkforceConsumer) markProcessed(ctx context.Context, eventID string, subject string) {
    query := `INSERT INTO audit.processed_events (event_id, subject)
             VALUES ($1, $2) ON CONFLICT DO NOTHING`
    _, _ = c.pool.Exec(ctx, query, eventID, subject)  // ← ERRORS IGNORED!
}
```

**Fix Strategy:**

1. **markProcessed() must return error** (not swallow silently)
2. In handleMessage(), check error before ACK
3. If markProcessed() fails → NAK message (requeue for retry)
4. Only ACK if marking succeeded
5. Apply same pattern to all 4 audit-service consumers

**Correct Implementation:**

```go
func (c *WorkforceConsumer) markProcessed(ctx context.Context, eventID string, subject string) error {
    query := `INSERT INTO audit.processed_events (event_id, subject)
             VALUES ($1, $2) ON CONFLICT DO NOTHING`
    if err := c.pool.Exec(ctx, query, eventID, subject); err != nil {
        c.logger.Error("failed to mark event as processed",
            zap.String("event_id", eventID),
            zap.Error(err))
        return err  // ← MUST RETURN ERROR
    }
    return nil
}

// In handleMessage():
result, err := c.handler.Handle(ctx, cmd)
if err != nil {
    c.logger.Error("failed to record audit entry", zap.Error(err))
    sharednats.NakMessage(c.logger, msg)
    return
}

// CRITICAL: Check if marking succeeded before ACK
if err := c.markProcessed(ctx, envelope.EventID, msg.Subject); err != nil {
    c.logger.Error("failed to mark processed, NAKing message", zap.Error(err))
    sharednats.NakMessage(c.logger, msg)  // ← REQUEUE IF MARKING FAILS
    return
}

sharednats.AckMessage(c.logger, msg)  // ← ONLY ACK AFTER SUCCESSFUL MARK
```

**Expected Outcome:**

- Errors logged with event_id
- Failed idempotency records cause NAK/requeue
- No silent failures → guaranteed idempotency

---

## ISSUE 2: Context Not Propagated in NATS Handlers

**Problem:** Using `context.Background()` instead of shutdown context → message loss on restart

**Files Affected:**

- `services/audit-service/internal/infrastructure/nats/workforce_consumer.go:41`
- `services/audit-service/internal/infrastructure/nats/identity_consumer.go:52`
- `services/audit-service/internal/infrastructure/nats/operations_consumer.go`
- `services/audit-service/internal/infrastructure/nats/notification_consumer.go`
- `services/notification-service/internal/infrastructure/nats/leave_consumer.go:52`
- `services/notification-service/internal/infrastructure/nats/employee_consumer.go`
- `services/notification-service/internal/infrastructure/nats/auth_consumer.go`

**Current Code (audit workforce_consumer.go:41):**

```go
func (c *WorkforceConsumer) Subscribe(ctx context.Context) error {
    _, err := c.js.Subscribe("hris.workforce.>", func(msg *nats.Msg) {
        c.handleMessage(context.Background(), msg)  // ← IGNORES PARENT CTX!
    }, nats.Durable("audit-workforce-consumer"))
    return err
}
```

**Fix Strategy:**

1. Capture parent context from Subscribe() method
2. Create message-scoped context with timeout
3. Pass message context to handleMessage()
4. Graceful shutdown will cancel parent context → all handlers stop
5. Apply to all 7 consumer files

**Expected Outcome:**

- Graceful shutdown works properly
- No "hung" message handlers on restart
- Messages get NAK'd and retried

---

## ISSUE 3: Missing Durable Consumer Names

**Problem:** Some consumers missing Durable() → messages not replayed on restart

**Files Affected:**

- `services/notification-service/internal/infrastructure/nats/leave_consumer.go:51` ← NO DURABLE!
- Possibly employee_consumer.go, auth_consumer.go

**Current Code (notification leave_consumer.go:51):**

```go
_, err := c.js.Subscribe("hris.operations.leave.>", func(msg *nats.Msg) {
    c.handleMessage(context.Background(), msg)
})  // ← NO Durable()!
```

**Fix Strategy:**

1. Add `nats.Durable("notification-leave-consumer")` to all subscribers
2. Add other JetStream options for robustness: MaxAckPending, AckWait
3. Apply to all 7 consumer files

**Expected Outcome:**

- Durable consumer state preserved across restarts
- No message loss on deployment
- NATS tracks delivery state per consumer

---

## ISSUE 4: N+1 Query Pattern in GetPermissions

**Problem:** 6 queries instead of 1 for permission check (impacts 80% of requests)

**Files Affected:**

- `services/auth-service/internal/application/queries/get_permissions.go`
- `services/auth-service/internal/infrastructure/postgres/permission_repository.go`

**Current Pattern:**

```go
// Query 1: Get all roles for user
roles, err := h.userRepo.GetRoles(ctx, query.TenantID, query.UserID)
// Returns: [role_admin, user_manager, employee]

// Queries 2-N: Get permissions for EACH role
for _, role := range roles {
    perms, err := h.permissionRepo.GetForRole(ctx, role.ID())
    // 3 additional queries = 6 total instead of 1
}
```

**Fix Strategy:**

1. Implement new method: `permissionRepo.GetForUserByRoles(roleIDs []string)`
2. Single query with JOIN + ANY() clause
3. Update handler to call new method
4. Benchmark: 6 queries → 1 query

**Expected Outcome:**

- Auth latency: 250ms → 30ms
- Database load reduced 6x
- Query cache hit rate improves

---

## ISSUE 5: Connection Pool Unconfigured

**Problem:** Default 4 connections → exhaustion at ~200 RPS

**Files Affected:**

- `services/auth-service/cmd/server/main.go:36-40`
- `services/employee-service/cmd/server/main.go`
- `services/attendance-service/cmd/server/main.go`
- `services/leave-service/cmd/server/main.go`
- `services/notification-service/cmd/server/main.go`
- `services/audit-service/cmd/server/main.go`

**Current Code:**

```go
poolConfig, err := pgxpool.ParseConfig(dbURL)
pool, err := pgxpool.NewWithConfig(ctx, poolConfig)  // ← Uses defaults!
```

**Fix Strategy:**

1. Create shared utility: `services/_shared/database/pool.go`
2. Set: MaxConns=25, MinConns=5, MaxConnLifetime=15min, MaxConnIdleTime=5min
3. Add health check: HealthCheckPeriod=1min
4. Apply to all 6 services

**Expected Outcome:**

- No connection exhaustion under load
- Warm connections reduce latency
- Stale connections cleaned up

---

## ISSUE 6: Non-Atomic Multi-Step Operations

**Problem:** Multi-step operations not coordinated → data corruption possible

**Files Affected:**

- `services/leave-service/internal/application/commands/approve_leave_request.go`
- `services/leave-service/internal/application/commands/request_leave.go`

**Current Pattern:**

```go
// Step 1: Get leave request
leaveRequest, _ := h.leaveRequestRepo.GetByID(ctx, ...)

// Step 2: Update balance (SEPARATE TRANSACTION)
balance.RemovePending(...)
h.employeeRepo.UpdateBalance(ctx, balance)  // ← Transaction 1

// Step 3: Update leave request (SEPARATE TRANSACTION)
leaveRequest.Approve()
h.leaveRequestRepo.Update(ctx, leaveRequest)  // ← Transaction 2 (fails!)

// Result: Balance corrupted!
```

**Fix Strategy:**

1. Wrap all updates in single `WithTenantTx()` transaction
2. Create repository methods accepting `tx` parameter
3. All updates happen atomically
4. Either all succeed or all rollback

**Expected Outcome:**

- Data consistency guaranteed
- No orphaned state
- Audit trail accurate

---

## Implementation Order (Days 1-6)

### Day 1: NATS Idempotency + Context (Issues 1 & 2)

**Effort:** 1 day (6 hours)

- Fix audit-service 4 consumers
- Fix notification-service 3 consumers
- Add error handling + context propagation
- Add/verify Durable() names

**Files to Change:** 7 files
**Test:** Publish event twice, verify only one processing

### Day 2: Connection Pool Configuration (Issue 5)

**Effort:** 1 day (4 hours)

- Create shared/database/pool.go
- Apply to all 6 services
- Test: Verify pool size in logs

**Files to Change:** 7 files (1 shared + 6 services)

### Day 3: N+1 Query Fix (Issue 4)

**Effort:** 1 day (5 hours)

- Implement GetForUserByRoles()
- Update handler
- Benchmark before/after
- Add test case

**Files to Change:** 2 files
**Test:** Measure query count reduction

### Days 4-6: Transaction Atomicity (Issue 6)

**Effort:** 2 days (8 hours)

- Refactor repositories to accept tx param
- Update 2 commands
- Add transaction-level tests
- Verify rollback behavior

**Files to Change:** 5 files

---

## Verification Strategy

### Day 1 Verification (NATS)

```bash
# 1. Verify error logging
- Publish event with bad data
- Check logs: "failed to mark event as processed"

# 2. Verify durable consumer
nats stream info HRIS_EVENTS | grep Consumers
# Should show "audit-workforce-consumer", "notification-leave-consumer", etc.

# 3. Verify idempotency
- Publish same event twice
- Verify: only one processing, no duplicates
- Check: processed_events table has only 1 entry
```

### Day 2 Verification (Connection Pool)

```bash
# 1. Verify pool configuration
SELECT datname, usename, count(*) FROM pg_stat_activity
WHERE datname = 'hris_db' GROUP BY datname, usename;
# Should show ≤25 connections per service

# 2. Verify no exhaustion under load
ab -n 10000 -c 100 http://localhost:50051/api/test
# Should complete without timeout
```

### Day 3 Verification (Query Optimization)

```bash
# 1. Measure query reduction
# Enable query logging in PostgreSQL
- Run: SELECT * FROM auth.GetPermissions(...) - OLD
- Count: 6 separate queries
- Run: WITH new implementation
- Count: 1 query with JOIN
- Measure latency: 250ms → 30ms

# 2. Test permission caching
- Login user
- Check logs: permission query executed
- Login same user again
- Check logs: cache hit (0 queries)
```

### Days 4-6 Verification (Transactions)

```bash
# 1. Verify atomicity
- Inject failure at step 2 of ApproveLeaveRequest
- Verify: Both updates rolled back or both succeeded
- Check: Balance not corrupted

# 2. Verify concurrent updates
- 10 parallel approval requests
- Verify: No negative balances
# Check: Audit trail consistent
```

---

## Success Criteria

By end of Week 1:

- ✅ All 7 NATS consumers have proper error handling + context + Durable()
- ✅ Connection pool configured on all 6 services
- ✅ N+1 query fixed (1 query instead of 6)
- ✅ Transaction atomicity implemented on leave-service
- ✅ All changes committed with verification
- ✅ No regressions in unit tests
- ✅ Ready for High-Impact Optimizations (Week 2)

---

## Risk Assessment

### Low Risk ✅

- Adding error logging (no logic change)
- Adding context propagation (standard pattern)
- Adding Durable() (JetStream feature)
- Configuration change (MaxConns, timeouts)

### Medium Risk ⚠️

- N+1 query fix (behavior change, needs test)
- Transaction refactoring (critical path, needs extensive testing)

### Mitigation

- All changes tested locally first
- Unit tests run before commit
- Integration tests added
- Gradual rollout (one service at a time)

---

## Next Steps

1. ✅ **Validate plan** with event-nats-reviewer agent
2. ✅ **Start Day 1** implementation (NATS fixes)
3. ✅ **Verify each change** before moving to next
4. ✅ **Create PRs** for team review
5. ✅ **Commit changes** in logical chunks
