# ADR Review Skill

**Trigger:** Before writing code that modifies a service or domain area

**Purpose:** Ensure architectural decisions are documented and understood before implementation

## Instructions

1. Identify the domain/service being changed (e.g., employee-service, auth-service, leave-service)
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
- [ ] No conflicts with locked decisions (ADR-0001, ADR-0002, ADR-0003, ADR-0004)

## Locked ADRs (Do Not Override)

- ADR-0001: grpc-gateway over gRPC-PHP
- ADR-0002: PostgreSQL RLS for tenant isolation
- ADR-0003: NATS JetStream event backbone
- ADR-0004: Defer payroll/recruitment/performance to Phase 2

## Examples

**Trigger:** "Add a new field to the Employee model"
→ Read `docs/adr/employee-service.md` (if exists) to confirm domain design

**Trigger:** "Implement leave request approval workflow"
→ Check `docs/adr/leave-service.md` and ADR-0003 (event design)

**Trigger:** "Add cross-service database query"
→ Check ADR-0002 and CLAUDE.md — this violates RLS and service boundaries
