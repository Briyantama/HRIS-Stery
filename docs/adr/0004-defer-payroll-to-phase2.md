# ADR-0004: Defer Payroll, Tax Rules, Recruitment, and Performance to Phase 2

**Status:** Accepted  
**Date:** 2026-05-30  
**Deciders:** Architecture Team  
**Supersedes:** —

---

## Context

The original research document lists payroll, recruitment, and performance reviews as Phase 1 MVP features. After domain analysis, all three are deferred to Phase 2.

This ADR explains why, defines what "deferred" means in practice, and establishes the guard rails that prevent premature implementation.

---

## Why Payroll Must Be Deferred

Indonesian payroll is a legally regulated computation. Errors are not bugs — they are regulatory violations with financial and legal consequences for the tenant (employer).

The computation involves:

| Component | Complexity |
|---|---|
| PPh 21 (income tax) | Progressive brackets, effective rates vs. gross rates, non-taxable income (PTKP), varies by employment type (permanent, contract, foreign worker) |
| BPJS Kesehatan | Employer 4% + Employee 1% of basic salary, capped at salary ceiling, changes annually |
| BPJS Ketenagakerjaan | JKK, JKM, JHT, JP — each with different employer/employee split percentages |
| THR (holiday allowance) | Pro-rated by tenure, legally mandated deadline, penalty for late payment |
| Uang Pesangon | Severance calculation based on tenure, reason for termination |
| UMP/UMR compliance | Minimum wage varies by province, updated annually |
| Payslip format | Must meet Kemenaker requirements |

These rules change every year via government regulation. The system requires:
- A versioned `tax-rules-service` that stores rule tables by effective date
- A `PayrollRuleSet` aggregate that knows which rules apply to a given pay period
- A separate admin interface for finance teams to configure allowances, deductions, and tax method
- An auditable, immutable `PayrollRun` that records the exact rules used for each calculation
- Legal review of the tax calculation logic before any real payroll is processed

**Building this correctly takes a dedicated design sprint before any code is written.** Building it incorrectly means tenants process incorrect payroll — a situation that cannot be quietly fixed in a patch release.

---

## Why Recruitment Must Be Deferred

Recruitment introduces a new bounded context: `Candidate`. A Candidate is not an Employee. Candidate data may be retained beyond a failed application (for PDPA compliance this is actually a liability). The recruitment pipeline involves:
- Job posting lifecycle (draft → published → closed → archived)
- Multi-stage interview scheduling (requires a scheduling domain)
- Resume parsing via ai-service (this is Phase 1 in the AI service, but the recruitment orchestration is not)
- Offer letter generation (document-service dependency)
- Onboarding trigger (creates an Employee — cross-domain event)

The `Candidate → Employee` transition is a critical domain event that requires careful design. Rushing this risks data model decisions that are difficult to reverse.

---

## Why Performance Reviews Must Be Deferred

Performance reviews require:
- Review cycle configuration (quarterly, annual, custom)
- Multi-rater support (self, peer, manager, skip-level)
- Competency framework management
- Calibration workflows
- Integration with compensation (performance → salary adjustment)

The last point — performance feeding into payroll — creates a dependency on the payroll domain. Since payroll is deferred, performance reviews are also deferred.

---

## Decision

**Payroll, tax-rules, recruitment, and performance are Phase 2. They are excluded from MVP implementation.**

---

## What "Deferred" Means in Practice

### In the codebase

Phase 2 service directories are **not created** in Sprint 0. No scaffold, no placeholder, no `TODO` services.

Phase 2 API routes exist in Laravel gateway but return `501 Not Implemented`:

```php
// routes/api.php
Route::middleware('auth:sanctum')->group(function () {
    // Phase 2 — not implemented
    Route::any('/payroll/{any}', fn() => response()->json([
        'code'    => 'NOT_IMPLEMENTED',
        'message' => 'Payroll module is not available in this release.',
    ], 501))->where('any', '.*');

    Route::any('/recruitment/{any}', fn() => response()->json([
        'code'    => 'NOT_IMPLEMENTED',
        'message' => 'Recruitment module is not available in this release.',
    ], 501))->where('any', '.*');
});
```

### In the frontend

Phase 2 navigation items exist in the sidebar but are marked `coming soon` and are non-clickable. No route files are created for Phase 2 modules.

### Proto files

`proto/hris/payroll/v1/payroll.proto` and similar are **not created** in Phase 1. Creating proto files signals intent to implement. Phase 2 protos are written at the start of the Phase 2 design sprint.

---

## Phase 2 Prerequisites (Before Any Payroll Code Is Written)

1. **Tax Rules Design Sprint** — a dedicated sprint to model PPh21, BPJS, THR as versioned domain objects. Output: a domain model document, not code.
2. **Legal Review** — tax calculation logic reviewed by a qualified Indonesian tax professional.
3. **`tax-rules-service` ADR** — defines how tax tables are stored, versioned, and queried.
4. **`PayrollRun` Immutability ADR** — defines how completed payroll runs are locked and audited.
5. **Saga Design** — payroll calculation spans attendance, employee, and tax-rules services. The compensating transaction design must be documented before implementation.

---

## Consequences

### Positive
- MVP ships faster. Phase 1 team is not blocked on payroll domain complexity.
- Payroll correctness is treated as a first-class requirement, not an afterthought.
- No incorrect payroll calculations reach production.
- The domain model can be designed with full knowledge of Phase 1 patterns.

### Negative
- Tenants cannot run payroll through the system in Phase 1. This must be communicated clearly in product positioning.
- The 501 routes must be maintained and not silently removed.

### Neutral
- Phase 2 kickoff is targeted for Week 15 (see milestone plan). The payroll design sprint is not blocked by Phase 1 completion — it can begin in parallel during the hardening sprint.

---

## Compliance

- No payroll calculation logic may be merged into any branch until Phase 2 prerequisites are complete.
- PRs that contain payroll, tax, recruitment, or performance business logic will be rejected at review with a reference to this ADR.
- Claude Code must refuse to generate payroll business logic and must cite ADR-0004.
