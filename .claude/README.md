# Claude Code Project Configuration — HRIS-Stery

**Status:** ✅ Complete and Ready to Use  
**Date:** 2026-06-01  
**Scope:** Skills, Subagents, and Hooks for HRIS-Stery Repository

---

## What Has Been Set Up

A **complete code review and architecture enforcement system** for the HRIS-Stery project, consisting of:

### 🎯 **6 Skills** (Reusable Workflows)

Step-by-step guides for common development tasks:

1. **adr-review** — Verify ADR exists before implementing architectural changes
2. **proto-review** — Verify proto contracts exist before writing gRPC handlers
3. **migration-review** — Ensure migrations are reversible with proper RLS
4. **event-review** — Validate event envelope and idempotency patterns
5. **architecture-review** — Enforce clean architecture boundaries
6. **go-style-review** — Check Go type safety, RLS context, error handling

### 🤖 **5 Subagents** (Specialized AI Reviewers)

AI agents with domain expertise for code review:

1. **go-backend-reviewer** — Reviews Go architecture, type safety, RLS enforcement
2. **laravel-gateway-reviewer** — Reviews Laravel API gateway patterns
3. **migration-reviewer** — Reviews database migrations for correctness
4. **event-nats-reviewer** — Reviews NATS event publishing/consumption
5. **proto-reviewer** — Reviews proto files and gRPC standards

### 🔒 **7 Hooks** (Automated Enforcement)

Deterministic rules that trigger automatically:

1. **phase-2-block** — Prevent Phase 2 features (payroll, recruitment, performance)
2. **adr-reminder** — Remind to check ADR when modifying services
3. **rls-reminder** — Remind about RLS when touching migrations
4. **migration-down-check** — Require both up and down migration files
5. **proto-before-handler** — Remind to verify proto before implementing handler
6. **controller-logic-limit** — Warn if Laravel controller logic exceeds 20 lines
7. **no-global-state** — Warn about package-level globals in Go

### 📚 **3 Guidance Documents**

1. **SETUP_GUIDE.md** — How to use the skills, subagents, and hooks
2. **REVIEW_DECISION_TREE.md** — Workflow guidance for different scenarios
3. **settings.json** — Configuration for all hooks and skills

---

## How to Use This System

### **Scenario 1: You're Starting a New Feature**

1. Read: `.claude/REVIEW_DECISION_TREE.md` to plan your workflow
2. Skill: **adr-review** — Check for relevant ADR in `docs/adr/`
3. Skill: **proto-review** — Verify proto contract exists
4. Implement: Domain → Application → Infrastructure → Interfaces
5. Subagent: **go-backend-reviewer** (or appropriate subagent) — Request code review
6. Commit: Follow the workflow's output

### **Scenario 2: You're Creating a Migration**

1. Skill: **migration-review** — Read migration checklist
2. Create: Both `NNN_description.up.sql` and `NNN_description.down.sql`
3. Subagent: **migration-reviewer** — Request code review
4. Commit: After review passes

### **Scenario 3: You're Publishing an Event**

1. Skill: **event-review** — Check event envelope requirements
2. Implement: NATS publisher with proper envelope
3. Subagent: **event-nats-reviewer** — Request code review
4. Commit: After review passes

### **Scenario 4: You're Writing a gRPC Handler**

1. Skill: **proto-review** — Verify proto contract exists
2. Skill: **architecture-review** — Keep handler thin (<20 lines)
3. Implement: Handler (thin) → Command/Query (logic) → Repository (DB)
4. Subagent: **go-backend-reviewer** — Request architecture review
5. Commit: After review passes

---

## Quick Reference: How to Trigger Each Component

### **Using Skills**

In your request, mention the skill:

```
Use the proto-review skill to verify the CreateEmployee proto exists.
```

Claude will read the skill and guide you through it.

### **Using Subagents**

Mention the subagent in your request:

```
Review this CreateEmployeeHandler code with go-backend-reviewer.
```

The subagent will review according to its guidelines.

### **Using Hooks**

Hooks activate automatically. No action needed. They will:
- **Block:** Prevent edits to Phase 2 files
- **Warn:** Remind you of requirements before committing

### **Using Decision Tree**

Read `.claude/REVIEW_DECISION_TREE.md` before starting:

```
I'm implementing a new RPC endpoint. Which skill/subagent should I use?
→ Read REVIEW_DECISION_TREE.md for the workflow
```

---

## The System in Action

### Example: Adding CreateLeaveRequest Endpoint

```
1. Check Decision Tree
   "New RPC endpoint" → Requires ADR → Proto → Architecture → Subagent review

2. Use adr-review Skill
   Read docs/adr/leave-service.md (or create if missing)

3. Use proto-review Skill
   Verify proto/hris/leave/v1/leave.proto has CreateLeaveRequestRequest/Response

4. Implement in Order:
   - Handler: internal/interfaces/grpc/leave_service.go (thin, delegates)
   - Command: internal/application/commands/create_leave_request.go (logic)
   - Repository: internal/infrastructure/postgres/leave_repository.go (DB)

5. Request Subagent Review
   "Review this with go-backend-reviewer"
   → Checks: Architecture, type safety, RLS, error handling

6. Commit After Review
   Hooks activate automatically to verify rules are followed
```

---

## File Structure

```
.claude/
├── README.md                         # This file
├── SETUP_GUIDE.md                    # Detailed usage guide
├── REVIEW_DECISION_TREE.md          # Workflow guidance
├── settings.json                     # Hook and config definitions
│
├── skills/                           # Reusable workflow checklists
│   ├── adr-review/SKILL.md
│   ├── proto-review/SKILL.md
│   ├── migration-review/SKILL.md
│   ├── event-review/SKILL.md
│   ├── architecture-review/SKILL.md
│   └── go-style-review/SKILL.md
│
└── agents/                           # Specialized AI reviewers
    ├── go-backend-reviewer.md
    ├── laravel-gateway-reviewer.md
    ├── migration-reviewer.md
    ├── event-nats-reviewer.md
    └── proto-reviewer.md
```

---

## Key Principles

### 1. **Repository Rules Are Non-Negotiable**

Defined in `CLAUDE.md`, enforced by hooks:
- ❌ Phase 2 features blocked
- ❌ No business logic in handlers/controllers
- ❌ No `interface{}` in domain types
- ❌ No cross-service DB access
- ✅ All migrations reversible (up + down)
- ✅ RLS enforced on tenant-scoped tables
- ✅ Events have UUID v7 event_id
- ✅ Consumers idempotent

### 2. **Clean Architecture Always**

Domain → Application → Infrastructure → Interfaces (no inversion)

**Enforced by:** architecture-review skill + go-backend-reviewer subagent

### 3. **Verify First, Code Second**

Before implementing:
1. Check ADR (adr-review skill)
2. Check proto contract (proto-review skill)
3. Plan architecture (architecture-review skill)
4. Implement
5. Request specialized review (appropriate subagent)

**Enforced by:** adr-reminder, proto-before-handler hooks

### 4. **Reversibility Always Matters**

Every migration must have a down file that fully reverses the up file.

**Enforced by:** migration-review skill + migration-reviewer subagent + migration-down-check hook

---

## Common Workflows

| Task | Skills to Use | Subagent | Hooks |
|------|---------------|----------|-------|
| New service feature | adr-review, proto-review, architecture-review | go-backend-reviewer | adr-reminder, proto-before-handler |
| New database table | migration-review | migration-reviewer | rls-reminder, migration-down-check |
| New event publisher | event-review | event-nats-reviewer | (none specific) |
| Event consumer | event-review | event-nats-reviewer | (none specific) |
| Laravel endpoint | architecture-review | laravel-gateway-reviewer | controller-logic-limit |
| Proto update | proto-review | proto-reviewer | (none specific) |

---

## When to Use Each Component

### **Skills** — Use When:
- You're starting a task and want a checklist
- You need to understand requirements before implementing
- You want to learn the project's patterns

### **Subagents** — Use When:
- You've finished implementing and want expert review
- You want to verify architectural correctness
- You're uncertain if your code follows patterns

### **Hooks** — Use When:
- They automatically trigger (no action needed)
- They block your edit (Phase 2 features)
- They remind you before committing

### **Decision Tree** — Use When:
- You're unsure which skill/subagent to use
- You're starting a new task
- You want to understand the workflow sequence

### **SETUP_GUIDE** — Use When:
- You're new to the project
- You want detailed instructions on using the system
- You need to understand how skills and subagents work

---

## Architecture Principles Enforced

### Domain Layer
- Pure business logic
- No I/O (no DB, HTTP, NATS)
- No external dependencies except stdlib
- Aggregates, value objects, domain events
- Invariants enforced

### Application Layer
- Orchestrates domain and infrastructure
- Commands (write) and Queries (read) — CQRS
- Uses domain methods and repository interfaces
- Publishes domain events
- Business logic lives here

### Infrastructure Layer
- Implements repository interfaces
- Handles I/O (DB, cache, NATS, HTTP)
- No business logic
- RLS context injection via WithTenantTx()

### Interface Layer
- gRPC handlers (thin, max 20 lines)
- Laravel controllers (thin, max 20 lines)
- Request validation only
- Delegates to application layer
- Error mapping to protocol format

---

## Getting Help

1. **"I'm not sure what to do"**
   → Read `.claude/REVIEW_DECISION_TREE.md`

2. **"I want to understand a rule"**
   → Read `CLAUDE.md` (source of truth)

3. **"I need architectural guidance"**
   → Use architecture-review skill or read relevant ADR

4. **"I need code review"**
   → Mention appropriate subagent (go-backend-reviewer, etc.)

5. **"I don't know if my code is correct"**
   → Request review with subagent specific to your code type

---

## Summary

| Component | Purpose | How to Use |
|-----------|---------|-----------|
| **Skills** | Step-by-step guidance for tasks | Read skill file before implementing |
| **Subagents** | Expert code review | Mention in request: "Review with X" |
| **Hooks** | Automated enforcement | They trigger automatically |
| **Decision Tree** | Workflow guidance | Read before starting task |
| **SETUP_GUIDE** | Detailed instructions | Reference when confused |
| **CLAUDE.md** | Source of truth for rules | Check when unsure about rule |

---

## Next Steps

1. **Read** `.claude/SETUP_GUIDE.md` for detailed instructions
2. **Bookmark** `.claude/REVIEW_DECISION_TREE.md` for quick workflow reference
3. **Remember** `.claude/CLAUDE.md` is the source of truth for rules
4. **Start your next task** by following the decision tree

---

**Created:** 2026-06-01  
**For:** HRIS-Stery Project  
**Maintainer:** Claude Code  
**Status:** ✅ Ready to Use

This system ensures code quality, architectural consistency, and rule compliance across the entire HRIS-Stery project.
