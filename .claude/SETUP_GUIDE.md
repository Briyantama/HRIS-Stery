# Claude Code Setup Guide — HRIS-Stery

This guide explains the Skills, Subagents, and Hooks configured for this repository.

**Goal:** Automate enforcement of architecture rules, prevent common mistakes, and guide code review.

---

## What Has Been Set Up

### 1. **Skills** (Reusable Workflows)

Located in `.claude/skills/`, these are step-by-step guides for repetitive tasks:

| Skill | Purpose | Trigger |
|-------|---------|---------|
| **adr-review** | Check if ADR exists before implementing | Before writing code that modifies a service |
| **proto-review** | Verify proto contract exists before handler | Before implementing gRPC handlers |
| **migration-review** | Ensure migrations are reversible + RLS | Before creating `.sql` files |
| **event-review** | Check event envelope + idempotency | Before publishing/consuming events |
| **architecture-review** | Verify clean architecture boundaries | Before writing handlers, controllers, repos |
| **go-style-review** | Check type safety, RLS, error handling | Before committing Go code |

**How to Use:** When facing a task that matches a skill's trigger, read the skill file and follow its checklist.

### 2. **Subagents** (Specialized Reviewers)

Located in `.claude/agents/`, these are AI agents with specific expertise:

| Subagent | Specialization | When to Use |
|----------|---|-----------|
| **go-backend-reviewer** | Go service architecture and correctness | Code review for any `.go` file in services/ |
| **laravel-gateway-reviewer** | API gateway and business logic separation | Code review for Laravel files |
| **migration-reviewer** | Database schema and reversibility | Code review for `.sql` migration files |
| **event-nats-reviewer** | Event envelope and idempotency | Code review for NATS publisher/consumer |
| **proto-reviewer** | Proto correctness and gRPC standards | Code review for `.proto` files |

**How to Use:** Mention the subagent name in your request, e.g., "Review this Go code with go-backend-reviewer"

Example:
```
Use go-backend-reviewer to review this CreateEmployeeHandler implementation
```

### 3. **Hooks** (Deterministic Enforcement)

Configured in `.claude/settings.json`, hooks are automated checks:

| Hook | Action | Trigger |
|------|--------|---------|
| **phase-2-block** | Block edits to Phase 2 services | Touching payroll-service, recruitment-service, etc. |
| **adr-reminder** | Remind to check ADR | Modifying services/*/internal/** |
| **rls-reminder** | Remind about RLS on tenant tables | Touching migration files |
| **migration-down-check** | Require both up and down files | Creating `.up.sql` file |
| **proto-before-handler** | Remind to check proto contract | Modifying gRPC handlers |
| **controller-logic-limit** | Warn if controller gets too large | Modifying Laravel controllers |
| **no-global-state** | Warn about package-level globals | Modifying Go service files |

**How They Work:** Hooks automatically trigger based on file paths. No manual action needed — they just remind or block.

---

## Quick Start: Decision Tree

**Before you start coding, ask yourself:**

1. **Am I creating a new service or feature?**
   - Read the **adr-review** skill
   - Check **REVIEW_DECISION_TREE.md** for the right sequence

2. **Am I implementing a gRPC handler?**
   - Use **proto-review** skill to confirm proto exists
   - Delegate to **go-backend-reviewer** subagent for review

3. **Am I writing a database migration?**
   - Use **migration-review** skill
   - Delegate to **migration-reviewer** subagent for review
   - Ensure `.down.sql` file exists and reverses `.up.sql`

4. **Am I publishing/consuming events?**
   - Use **event-review** skill to check envelope
   - Verify idempotency table for consumers
   - Delegate to **event-nats-reviewer** subagent for review

5. **I'm stuck on architecture decisions**
   - Read **REVIEW_DECISION_TREE.md**
   - Check **CLAUDE.md** for hard rules
   - Read relevant ADR in **docs/adr/**

---

## How to Trigger Skills & Subagents

### Manually Triggering a Skill

**In your request:**
```
Use the adr-review skill before implementing this.

Here's my situation: I'm adding a new endpoint to employee-service.
What should I check?
```

Claude will read the skill file and guide you through it.

### Invoking a Subagent

**In your request:**
```
Review this CreateEmployeeHandler code with go-backend-reviewer.

Here's the code: [paste code]
```

Claude will invoke the subagent, which will review the code according to its guidelines.

### Chaining Workflows

**Example for a new RPC endpoint:**
```
I'm implementing a new RPC endpoint. Help me:

1. Use adr-review to check the ADR
2. Use proto-review to verify proto contract exists
3. Implement the handler, command, and repository
4. Review with go-backend-reviewer

Here's my task: [describe what you're building]
```

---

## File Structure

```
.claude/
├── settings.json                  # Hook and configuration definitions
├── SETUP_GUIDE.md                # This file
├── REVIEW_DECISION_TREE.md       # Decision tree for choosing workflows
│
├── skills/
│   ├── adr-review/SKILL.md
│   ├── proto-review/SKILL.md
│   ├── migration-review/SKILL.md
│   ├── event-review/SKILL.md
│   ├── architecture-review/SKILL.md
│   └── go-style-review/SKILL.md
│
└── agents/
    ├── go-backend-reviewer.md
    ├── laravel-gateway-reviewer.md
    ├── migration-reviewer.md
    ├── event-nats-reviewer.md
    └── proto-reviewer.md
```

---

## Common Scenarios

### Scenario 1: Adding a New Employee Endpoint

```
1. Skill: adr-review
   → Check docs/adr/ for employee-service architecture

2. Skill: proto-review
   → Verify proto/hris/employee/v1/employee.proto exists
   → Add CreateEmployeeRequest/Response if not present

3. Implement:
   - Handler: internal/interfaces/grpc/employee_service.go (thin)
   - Command: internal/application/commands/create_employee.go (logic)
   - Repository: internal/infrastructure/postgres/employee_repository.go (DB)

4. Subagent: go-backend-reviewer
   → Review handler, command, repository for architectural correctness
```

### Scenario 2: Creating a Database Migration

```
1. Skill: migration-review
   → Check migration requirements and RLS patterns

2. Create:
   - services/{service}/migrations/NNN_description.up.sql
   - services/{service}/migrations/NNN_description.down.sql

3. Subagent: migration-reviewer
   → Review RLS policies, reversibility, grants
```

### Scenario 3: Publishing an Event

```
1. Skill: event-review
   → Check event envelope structure
   → Verify UUID v7, tenant_id, subject taxonomy

2. Implement: NATS publisher in infrastructure layer

3. Subagent: event-nats-reviewer
   → Review event envelope, subject naming, transaction handling
```

### Scenario 4: Implementing Event Consumer

```
1. Skill: event-review
   → Check idempotency pattern requirements
   → Verify processed_events table design

2. Implement: NATS consumer
   - Subscription handler
   - Idempotency check
   - Processing logic
   - Mark processed (ON CONFLICT DO NOTHING)

3. Subagent: event-nats-reviewer
   → Review idempotency, error handling, event envelope parsing
```

---

## Rules That Are Automatically Enforced

### Hard Blocks (Will prevent edits)

- ❌ **Phase 2 services:** payroll-service, recruitment-service, performance-service
  - Hook: phase-2-block
  - Message: Return 501 Not Implemented

### Warnings (Will remind you)

- ⚠️ **ADR existence:** Modifying a service
  - Hook: adr-reminder
  - Action: Check docs/adr/ before implementing

- ⚠️ **RLS enforcement:** Creating/modifying tenant-scoped tables
  - Hook: rls-reminder
  - Action: Verify RLS policy and FORCE ROW LEVEL SECURITY

- ⚠️ **Migration reversibility:** Creating up migration
  - Hook: migration-down-check
  - Action: Create corresponding down migration

- ⚠️ **Proto contract:** Implementing gRPC handler
  - Hook: proto-before-handler
  - Action: Verify proto/hris/{domain}/v1/{domain}.proto exists

- ⚠️ **Controller size:** Modifying Laravel controller
  - Hook: controller-logic-limit
  - Action: Keep controller ≤20 lines, delegate to Service

- ⚠️ **Global state:** Go services
  - Hook: no-global-state
  - Action: Use constructor injection, no package-level vars

---

## Using the Decision Tree

Open `.claude/REVIEW_DECISION_TREE.md` to answer:

**"What should I do before writing code?"**

The decision tree will guide you through:
1. What to read first (ADR, proto, etc.)
2. Which skill to use
3. Which subagent to invoke for review
4. What commands to run

---

## Repository Hard Rules

These cannot be overridden. Violations will be caught by hooks or subagents:

1. **No business logic in controllers/handlers** → Clean Architecture
2. **No `interface{}` or `any` in domain types** → Type Safety
3. **No direct cross-service DB access** → Service Boundaries
4. **Every migration needs `.up.sql` and `.down.sql`** → Reversibility
5. **Tenant-scoped tables need RLS** → Multi-Tenancy
6. **All events have UUID v7 event_id** → Audit Trail
7. **All consumers have idempotency table** → At-Least-Once Delivery
8. **No Phase 2 features (payroll, recruitment, performance)** → MVP Scope

See **CLAUDE.md** for the full set of rules.

---

## Getting Help

1. **"I'm not sure which skill to use"**
   → Read `.claude/REVIEW_DECISION_TREE.md`

2. **"I need to understand a rule"**
   → Read `CLAUDE.md` (project rules)

3. **"I need architectural guidance"**
   → Read relevant ADR in `docs/adr/`

4. **"I want to understand a specific pattern"**
   → Read the relevant skill file (e.g., `skills/architecture-review/SKILL.md`)

5. **"I need code review"**
   → Mention the appropriate subagent:
   - Go code → go-backend-reviewer
   - Laravel → laravel-gateway-reviewer
   - Migrations → migration-reviewer
   - Events → event-nats-reviewer
   - Proto → proto-reviewer

---

## Next Steps

1. **Review the skills** in `.claude/skills/` to understand what guidance is available
2. **Read the decision tree** at `.claude/REVIEW_DECISION_TREE.md` to understand workflow ordering
3. **Familiarize yourself with** `CLAUDE.md` — it's the source of truth for all rules
4. **When starting a task:**
   - Check the decision tree
   - Apply relevant skills
   - Request subagent review before committing

---

## Summary

| Component | Purpose | How to Use |
|-----------|---------|-----------|
| **Skills** | Step-by-step checklists for common tasks | Read the skill, follow its checklist |
| **Subagents** | AI experts for specialized code review | Mention in your request (e.g., "Review with go-backend-reviewer") |
| **Hooks** | Automated reminders and blocks | They activate automatically based on file changes |
| **Decision Tree** | Workflow guidance | Read before starting a task |
| **CLAUDE.md** | Source of truth for all rules | Reference when unsure about a rule |

---

**Created:** 2026-06-01  
**For:** HRIS-Stery Project  
**Status:** Ready to Use  
