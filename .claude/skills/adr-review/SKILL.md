# ADR Review Skill

**Trigger:** Before writing code that modifies a service or domain area

**Purpose:** Ensure architectural decisions are documented and understood before implementation

## Instructions

1. Identify the service package being changed (e.g., `services/employee`, `services/auth`, `services/leave`)
2. Check if relevant ADR exists in `docs/adr/`
3. If ADR exists:
   - Read the ADR
   - Confirm the proposed change aligns with the ADR
   - Flag any deviations
4. If no ADR exists:
   - Suggest creating one if the change is architectural
   - Proceed with caution and document assumptions in code comments

## ADR Checklist

- [ ] ADR document exists in docs/adr/
- [ ] ADR covers the service/domain being modified
- [ ] Proposed change aligns with ADR decisions
- [ ] No conflicts with in-force decisions (ADR-0002, ADR-0004, ADR-0005, ADR-0006)

## In-force ADRs (Do Not Override)

- ADR-0002: PostgreSQL RLS for tenant isolation (amended 2026-11-08)
- ADR-0004: Defer payroll/recruitment/performance to Phase 2
- ADR-0005: Single-module Go API on the SELA stack (supersedes ADR-0001)
- ADR-0006: Dragonfly Streams replace the old event backbone (supersedes ADR-0003)

## Examples

**Trigger:** "Add a new field to the Employee model"
→ Read ADR-0005 for package layout and ADR-0002 for the RLS rules

**Trigger:** "Implement leave request approval workflow"
→ Read ADR-0006 (event design) and the leave package in `apps/api/services/leave`

**Trigger:** "Add cross-service database query"
→ Check ADR-0002 and CLAUDE.md — this violates RLS and service boundaries
