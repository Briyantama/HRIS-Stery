# BRD HRIS-Stery

*Business Requirements Document (BRD): multi-tenant human resource information system (HRIS) for companies in Indonesia, sold as the `HRS` product on its own or as part of Stery Suite*

|                         |                                                                                                                                                                                                                                                                  |
|-------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Document Name**       | BRD - HRIS-Stery (SKU `HRS`)                                                                                                                                                                                                                                     |
| **Version**             | 1.0 (Draft)                                                                                                                                                                                                                                                      |
| **Date**                | 6 November 2026                                                                                                                                                                                                                                                  |
| **Prepared by**         | Product & Engineering Team                                                                                                                                                                                                                                       |
| **Status**              | Draft for review                                                                                                                                                                                                                                                 |
| **Related Documents**   | PRD HRIS-Stery (`PRD_HRIS.id.md`), FSD HRIS-Stery (`FSD_HRIS.id.md`), `00_Analisis_Dampak_Risiko`, BRD/PRD/FSD Stery Suite (folder `stery-suite/docs`), `SELA/CLAUDE.md` (stack standard), ADR-0001 through ADR-0004 (`HRIS-Stery/docs/adr`)                       |
| **Changes from previous** | First issue                                                                                                                                                                                                                                                    |

## 1. Executive Summary

HRIS-Stery is a multi-tenant HRIS application: a company (tenant) registers, manages employee data, daily attendance, and leave, and then HR and supervisors work with approval flows and an audit trail. This repository has never had a BRD, PRD, or FSD; this document is the first BRD and is written from reading the code (truth = code), not from the old status documents.

Position in the portfolio: HRIS-Stery is sold as SKU `HRS`, can stand alone (`STERY_MODE=standalone`, default) or connect to the Stery Suite platform (`STERY_MODE=suite`) together with POS-Stery (SKU `POS`) and SELA (SKU `SEL`). The entire stack, patterns, and architecture of HRIS-Stery will be aligned with the SELA standard (owner decision, see 9.3 and FSD 2.1) so that one team can operate all three products in the same way.

Core finding from reading the code (status per module in section 13): the HRIS-Stery code already has a layered architecture (domain, application, infrastructure, interfaces) across seven Go services and 39 of 47 proto RPCs have substantive handlers, but **not a single end-to-end flow has been proven to work**: the attendance module migration contains invalid syntax (`services/attendance-service/migrations/001_create_attendance_schema.up.sql:25`), no HTTP server serves the grpc-gateway path that the Laravel gateway calls, the web login page is mocked to `http://localhost:3000/api/auth/login`, no service checks authentication or authorization, and NIK/NPWP/date of birth are stored as plain text. The old status documents (`PROJECT_PROGRESS.md`, `PHASE1_RELEASE.md`, `IMPLEMENTATION_COMPLETE.md`) claim "100% / production ready" and contradict the code (list in FSD 11.3). For that reason the work before sale is split into three: (1) the HRIS stabilisation gate, (2) stack alignment to the SELA standard, (3) suite mode readiness.

The payroll, tax rules, recruitment, and performance modules are **deferred** per ADR-0004 (accepted 2026-05-30) and may only start after a legal review of PPh21, BPJS, and THR. The AI assistant module has no code yet (proto contract only).

Prices, tiers, SLAs, market size, and competitor data do not exist in any repository and are recorded as `tbd` with open questions (9.4).

## 2. Background and Problem Statement

Companies in Indonesia manage employees, attendance, and leave with a mix of spreadsheets, instant messaging, and separate systems. The following problem statements are product hypotheses and have **not been validated by user research** `[INFERENCE]` (research questions: Q-HRS-01, Q-HRS-03):

- Employee master data is scattered and has no single source of truth; changes of position and supervisor are not documented.
- Leave requests and approvals via chat are hard to track, leave balances are calculated manually, and there is no decision history.
- Attendance recording is not uniform and corrections leave no trail.
- Employee documents (contracts, identity) are stored in shared folders without access control and versioning.
- Employee personal data (NIK, NPWP, date of birth) requires protection in accordance with Personal Data Protection Law No. 27 of 2022 (UU PDP), while this repository does not yet contain a single policy or protection mechanism (FSD 8).

From the product owner's side, an additional problem: the owner sells three products (HRIS, POS, SELA) and wants to market them as individual packages or bundles; HRIS-Stery currently uses a different stack from SELA (Laravel gateway, SvelteKit SSR, JWT, NATS, golang-migrate, sqlc, zap, MinIO, Redis) so it cannot be operated uniformly.

The design scale target recorded in ADR-0002 is "hundreds to low thousands of tenants" (`docs/adr/0002-postgres-rls-for-tenant-isolation.md`); that figure is a design assumption, not a business target.

## 3. Business Objectives

1. Provide a multi-tenant HRIS that is safe for real companies to use: employee master data, attendance, leave, notifications, audit, and documents (BR-HRS-01 through BR-HRS-10).
2. Guarantee data isolation between tenants and personal data protection in accordance with UU PDP as a condition of sale, not an add-on feature (BR-HRS-02, BR-HRS-12).
3. Make HRIS-Stery a product ready to be sold standalone or in a suite bundle without changing standalone behaviour (BR-HRS-15, BR-HRS-16).
4. Bring the code from "partially wired" to "proven to work" through a stabilisation gate before real customers (BR-HRS-17).
5. Align the stack to the SELA standard to keep the operating and development cost of the three products low (BR-HRS-18).
6. Manage legally risky features (payroll, tax) with discipline: not released before legal review (BR-HRS-13, BR-HRS-14).
7. Keep status honest: progress reports must match the code (BR-HRS-22).

## 4. Scope

### 4.1 Segments and Packages

| **Aspect**                 | **Value**                                                                                                                        |
|----------------------------|----------------------------------------------------------------------------------------------------------------------------------|
| Target segment             | `tbd` (no segment statement in the repository; Q-HRS-01)                                                                         |
| Product SKU                | `HRS`                                                                                                                            |
| Bundles (defined in the Stery Suite BRD) | `HRS+POS`, `HRS+SEL`, `SUITE` (all three)                                                                          |
| Operating mode             | `standalone` (default; own users, sessions, and data) or `suite` (identity, entitlement, billing, notification, audit delegated to the platform) |
| Limit dimension            | `hrs.employees.max` (value per tier `tbd`; Q-HRS-02)                                                                             |
| Tier                       | `tbd` (number and names of tiers; Q-HRS-02)                                                                                      |

### 4.2 In Scope

In scope for the first wave (catalogue modules 01-08, see FSD 4):

- 01 AUTH: tenant registration, login, session, RBAC (roles `hr_admin`, `manager`, `employee` and `resource:action` permissions).
- 02 EMP: employees, departments, positions, org chart, termination.
- 03 ATT: check-in/check-out, correction by HR, attendance list and summary.
- 04 LEV: leave types, balances, requests, approval, rejection, cancellation.
- 05 NTF: event-based email and in-app notifications.
- 06 AUD: append-only audit trail.
- 07 DOC: employee documents (upload via signed URL, versions, classification, retention).
- 08 AI: AI assistant (resume analysis, attrition prediction, chat); in scope as a contract, not yet worked on (status Not started).
- UU PDP personal data protection for all modules (FSD 8).
- HRIS stabilisation gate, stack alignment to the SELA standard, and suite mode adapters (section 12).

### 4.3 Deferred

| **Module** | **Reason for deferral**                                                                                                                                                                                          | **Prerequisites before starting**                                                                                                         |
|-----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------|
| 09 PAY payroll | Indonesian payroll calculation is regulated by law and changes every year (PPh21, BPJS Kesehatan, BPJS Ketenagakerjaan, THR, severance, UMP). An error is a regulatory violation, not a bug (ADR-0004)         | Legal review by an Indonesian tax professional of the PPh21/BPJS/THR logic (Q-HRS-10); tax rules design sprint; `tax-rules-service` ADR; `PayrollRun` immutability ADR; cross-module saga design |
| 10 TAX tax rules | Rules versioned by effective date; part of payroll (ADR-0004)                                                                                                                                          | Same as 09                                                                                                                                |
| 11 REC recruitment | Introduces a new `Candidate` context; retention of failed-candidate data is a UU PDP liability; the `Candidate -> Employee` transition needs design (ADR-0004)                                    | Candidate domain design, candidate retention policy with legal review, LLM provider decision for resume parsing (Q-HRS-09)                |
| 12 PRF performance | Requires appraisal cycles, multiple raters, calibration, and is tied to compensation, so it depends on payroll (ADR-0004)                                                                            | Payroll done and legal review; competency framework design                                                                                |

Rule applying during deferral (ADR-0004 Compliance section): no payroll, tax, recruitment, or performance logic may be merged into any branch; pull requests containing it are rejected with reference to ADR-0004.

### 4.4 Out of Scope

The following items are proposed as out of scope for this phase and await product owner approval (section 14):

- Calculation of salary, tax, BPJS, THR, severance (see 4.3).
- Native mobile app; web access (SPA) only.
- Integration with government systems (BPJS online, tax e-filing, Kemnaker) while payroll is deferred.
- Automatic contract-based standalone billing inside HRIS; suite billing is done by the platform (Stery Suite BRD).
- Enterprise-specific features (third-party corporate SSO, annual contracts) until demand is proven (Q-HRS-01).

## 5. Stakeholders

| **Role**                                    | **Description of Interest**                                                                                   | **Level of Influence** |
|---------------------------------------------|---------------------------------------------------------------------------------------------------------------|----------------------|
| Product Owner / Business Sponsor            | Sets segment, price, tier, and phase priorities; approves this document                                       | High                 |
| Tenant HR admin (`hr_admin`)                | Manages employees, leave, attendance corrections, documents; heaviest daily user                               | High                 |
| Manager / direct supervisor (`manager`)     | Approves leave and views team attendance                                                                       | Medium               |
| Employee (`employee`)                       | Check-in/out, applies for leave, views balance and notifications                                               | Medium               |
| Platform admin (Stery operator)             | Manages tenants, entitlements, and support; in suite mode acts through the platform console                    | High                 |
| Engineering Team                            | Builds, stabilises, and aligns the stack; bears operations                                                     | High                 |
| Legal / Compliance Team                     | Sets processing basis, retention, data subject rights (UU PDP) and approves before payroll starts              | High                 |
| Personal data owner (tenant employee)       | UU PDP data subject: entitled to access, correction, and deletion as provided                                  | Medium               |
| Business / Marketing Team                   | Pricing per tier and bundle, competitor research                                                               | Medium               |

## 6. Business Requirements

Priority: M = Must have, S = Should have, C = Could have. Deferred items are marked (Deferred) and remain recorded so the decision trail exists. The Module column refers to the FSD 4 module catalogue (SEC = cross-module, FSD section 8).

| **ID**    | **Business Requirement**                                                                                                                                                                                              | **Priority**  | **Module**      |
|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------|-----------------|
| BR-HRS-01 | A company (tenant) can register itself and immediately has a first HR admin with default roles and permissions.                                                                                                       | M             | 01 AUTH         |
| BR-HRS-02 | One tenant's data can never be read or written by another tenant; isolation is enforced in the database (RLS) and proven by cross-tenant tests in every module.                                                       | M             | SEC, 01-07      |
| BR-HRS-03 | Role-based access (HR admin, manager, employee) is enforced in the backend on every endpoint, not merely hidden in the interface.                                                                                     | M             | 01 AUTH         |
| BR-HRS-04 | HR admin can manage master data: create, change, and terminate employees; manage departments (hierarchical) and positions; view the org chart.                                                                       | M             | 02 EMP          |
| BR-HRS-05 | Employees can record daily attendance (check-in/check-out); HR can correct records with a reason and audit trail; HR and supervisors view attendance lists and summaries.                                             | M             | 03 ATT          |
| BR-HRS-06 | Working hours rules, lateness threshold, working days, tenant time zone, and (optionally) attendance location can be configured per tenant; default values `tbd` (Q-HRS-05).                                         | S             | 03 ATT          |
| BR-HRS-07 | Employees can apply for leave; supervisor or HR approves or rejects; working days are calculated by the system; balances are always consistent; overlapping requests are rejected; cancellation restores the balance. | M             | 04 LEV          |
| BR-HRS-08 | Relevant parties receive email and in-app notifications for important events (leave request and decision, new employee, account activation) and can mark them as read.                                                | S             | 05 NTF          |
| BR-HRS-09 | Sensitive actions (login, employee changes, attendance corrections, leave decisions, access to confidential documents, PII changes) are recorded in an append-only audit trail that HR or auditors can query and that cannot be altered. | M             | 06 AUD          |
| BR-HRS-10 | Employee documents are stored securely: upload and download via short-lived signed URLs, malware scanning, versions and version restore, classification, and retention.                                               | S             | 07 DOC          |
| BR-HRS-11 | The AI assistant (resume analysis, attrition prediction, HR chat) can be offered once the PII processing policy toward the LLM provider, per-tenant cost limits, and consent are set (Q-HRS-09).                       | C             | 08 AI           |
| BR-HRS-12 | Employee personal data (NIK, NPWP, date of birth, attendance location, leave reason) is protected in accordance with UU PDP: classification, encryption, processing basis and consent, retention, data subject rights, incident procedure. | M             | SEC             |
| BR-HRS-13 | Payroll, tax rules (PPh21), BPJS, and THR are deferred until legal review and ADR-0004 prerequisites are complete; no payroll logic is released before then. (Deferred)                                              | C             | 09 PAY, 10 TAX  |
| BR-HRS-14 | Recruitment and performance appraisal are deferred per ADR-0004. (Deferred)                                                                                                                                           | C             | 11 REC, 12 PRF  |
| BR-HRS-15 | HRIS can be sold and run standalone or as part of Stery Suite; `standalone` behaviour is not changed by the existence of `suite` mode.                                                                                | M             | SEC, 01         |
| BR-HRS-16 | Sales are limited by the `hrs.employees.max` dimension; tiers and prices `tbd` (Q-HRS-02); exceeding the limit or entitlement expiry is handled without data loss (only creation of new employees is rejected).      | S             | 02 EMP          |
| BR-HRS-17 | The product must pass the HRIS stabilisation gate (section 12) before being sold or used by real customers.                                                                                                           | M             | SEC             |
| BR-HRS-18 | The product is aligned to the standard suite stack (SELA) so operations, security, and development are uniform across products.                                                                                       | M             | SEC             |
| BR-HRS-19 | The interface is available in Indonesian (default) and English; WIB time zone; IDR currency where monetary values exist.                                                                                              | S             | SEC             |
| BR-HRS-20 | The service can be operated and recovered: `/healthz`, structured logs without PII, rate limiting and account lockout, data backup (RPO/RTO `tbd`, Q-HRS-07).                                                        | M             | SEC             |
| BR-HRS-21 | Suite integration hypotheses (validate demand first): INT-1 link HRIS employee staff with POS users (terminated employee deactivates the POS user); INT-2 corporate gathering event creates a SELA event.            | C             | 02 EMP          |
| BR-HRS-22 | Implementation status reports must match the code and contain evidence; old status documents that contradict the code are consolidated or withdrawn.                                                                  | S             | SEC             |

## 7. Business Model and Monetisation

All HRS monetary values are `tbd` (no price, tier, or billing mechanism in the repository; Q-HRS-02). Proposed model, following suite decision D5:

| **Tier**                       | **Employee limit (`hrs.employees.max`)** | **Price** | **Billing cycle** | **Trial** | **Notes**                                                                                     |
|--------------------------------|------------------------------------------|-----------|------------------|-----------|-----------------------------------------------------------------------------------------------|
| `tbd` (number and names of tiers)   | `tbd`                                    | `tbd`     | `tbd`            | `tbd`     | One limit dimension; values and tiers decided by the owner (Q-HRS-02)                         |
| Bundles `HRS+POS`, `HRS+SEL`, `SUITE` | follows the HRS tier                | `tbd`     | `tbd`            | `tbd`     | Bundle discount rule `tbd` (managed in the Stery Suite BRD)                                   |

- Billing in `suite` mode is done by the platform using SELA's `payment.Provider` abstraction (Midtrans QRIS); billing in `standalone` mode does not exist yet and its mechanism is `tbd` (manual contract or platform billing; Q-HRS-02).
- HRS feature rights are applied through entitlement: module feature keys and the `hrs.employees.max` dimension; when the limit is exceeded, creation of new employees is rejected while existing data remains readable and exportable.
- The payroll, tax, recruitment, and performance modules (deferred) may become paid add-ons; decision `tbd`.

## 8. Alternatives and Competitor Analysis

### 8.1 Competitors

All content in this section is `tbd`: there is no competitor research in the HRIS-Stery repository (`deep-research-report.md` is a technical plan, not market research).

| **Competitor**     | **Market scale evidence** | **Pricing model** | **Segment focus** | **Payroll/PPh21 features** | **Indonesian localisation** |
|--------------------|-----------------------|-----------------|------------------|-------------------------|--------------------------|
| `tbd` (list from research) | `tbd`          | `tbd`           | `tbd`            | `tbd`                   | `tbd`                    |

Research questions (Q-HRS-03): (1) Which HRIS does the target segment use now and how much does it cost per employee per month? (2) Do buyers demand payroll/PPh21 in the first stage or are master data, attendance, and leave enough? (3) Is attendance with geofence or selfie a purchase requirement? (4) How do buyers assess cloud storage of employee data against UU PDP? (5) Is a bundle with POS or SELA valuable to HRS buyers?

### 8.2 Architecture Alternatives

Portfolio-level architecture alternatives (A polyrepo + platform, B monorepo, C single codebase) are discussed in Stery Suite BRD section 8 and `00_Analisis_Dampak_Risiko` section 5; decision: option A (D-HRS-01). Product-level alternatives considered for HRIS:

| **Alternative**                                             | **Assessment**                                                                                                                                                    |
|-------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Keep the current stack (Laravel + many gRPC services + NATS) | Rejected: conflicts with the owner decision "stack = SELA"; the gateway does not work (FSD 11.3); high operating cost for a small team                           |
| Rewrite from scratch on the SELA stack                      | Rejected as default: the domain and application layers are reusable; what is rewritten is infrastructure and interfaces (FSD 2.7)                                  |
| Align incrementally: keep domain/application, rewrite infrastructure and interfaces | Chosen (D-HRS-11, D-HRS-12)                                                                                                              |
| Buy/integrate a third-party HRIS                            | `tbd` depending on research result Q-HRS-03; not evaluated in this document                                                                                       |

## 9. Assumptions and Constraints

### 9.1 Assumptions

- Users access via modern browsers; no native app need in this phase (proposal, 4.4).
- One tenant = one company; the relation of one suite organisation to one HRS tenant is described in FSD Stery Suite 7.5.
- The tenant acts as data controller and Stery as data processor for employee data; the division of roles and the data processing agreement are `tbd` (Q-HRS-14) `[INFERENCE]`.
- The development team is small, so one alignment is done at a time (suite decision D12).

### 9.2 Constraints

- Payroll, tax, recruitment, performance must not be worked on before the ADR-0004 prerequisites (4.3).
- The target stack follows SELA: Go >= 1.25.1, PostgreSQL 18, pgx/v5 without ORM, goose, Dragonfly, RustFS, SvelteKit SPA, no JWT and no NATS (FSD 2.1).
- `standalone` behaviour must remain intact when the platform is absent (decision D2).
- No price, SLA, quota, KPI, or competitor figures may be invented; all are `tbd` until the owner decides.
- This document does not change any code; the stack migration is specified as roadmap workstreams.

### 9.3 Applicable Decisions

Portfolio-level decisions refer to the suite plan (D1-D12); product-level decisions refer to the HRIS ADRs.

| **ID**     | **Decision**                                                                                                                                                                            | **Source**                                              |
|------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------|
| D-HRS-01   | Composition, not consolidation: HRIS remains a separate deployable and database; the platform (`stery-suite`) owns organisation, identity, entitlement, billing, notification hub, audit sink, console | Suite D1; ADR-0001 superseded (see D-HRS-12)             |
| D-HRS-02   | Two modes per product via `STERY_MODE=standalone\|suite` (default `standalone`); suite mode delegates identity, entitlement, billing, notification, audit through six Go ports with two adapters each | Suite D2                                  |
| D-HRS-03   | Suite session = opaque cookie `stery_session` (parent domain `tbd`), validated through `SessionService/Introspect` with a 30-second local cache; HRIS keeps password login (bcrypt, `services/auth-service/internal/domain/user.go:244-250`) but the session becomes opaque, JWT removed | Suite D3, D10 |
| D-HRS-04   | Entitlement through `EntitlementService/Check` and a signed snapshot; when the platform is down, the last snapshot is used for up to 24 hours (`tbd`) then paid features fail closed, never in standalone mode | Suite D4                              |
| D-HRS-05   | Events: no NATS; in-process domain events + Postgres transactional outbox + idempotent consumers; cross-product via outbox -> platform `IntegrationService/Publish` | Suite D6; ADR-0003 superseded                |
| D-HRS-06   | Data: one `hris` database and one `hris` application role per profile; no cross-database FK; RLS remains mandatory (ENABLE+FORCE, `USING` and `WITH CHECK`, NOBYPASSRLS role) through `kit/tenantdb.WithTenant` | Suite D7, D8; ADR-0002 retained      |
| D-HRS-07   | Web: one `apps/web` SvelteKit SPA (`ssr=false`, `adapter-static`)                                                                                                                         | Suite D9                                                |
| D-HRS-08   | Observability baseline = `slog` JSON + `/healthz`; OpenTelemetry/Prometheus not in the baseline, owner decision `tbd` (Q-HRS-08)                                                           | Suite D11                                               |
| D-HRS-09   | Work order: HRIS is aligned at stage S3 after POS (S2); one alignment runs at a time; ends at the HRIS stabilisation gate                                                                 | Suite D12                                               |
| D-HRS-10   | Payroll, tax rules, recruitment, performance deferred to phase 2                                                                                                                          | ADR-0004 (`docs/adr/0004-defer-payroll-to-phase2.md`)   |
| D-HRS-11   | The stack, patterns, system design, and architecture of the whole suite follow SELA; HRIS is re-specified onto that stack                                                                  | Owner decision (suite plan)                             |
| D-HRS-12   | The Laravel gateway tier is removed: Go serves REST `/api/v1` itself (stdlib `net/http`) and gRPC for machine clients; the ADR-0001 decision (grpc-gateway + Laravel) is superseded        | STACK-SPEC row 4                                        |

### 9.4 Open Questions

| **ID**    | **Question**                                                                                                                                | **Decision owner**     |
|-----------|---------------------------------------------------------------------------------------------------------------------------------------------|------------------------|
| Q-HRS-01  | Target segment and typical company size (number of employees) for HRS.                                                                      | Product owner          |
| Q-HRS-02  | Tiers, value of `hrs.employees.max` per tier, price, billing cycle, trial, and billing mechanism for standalone mode.                        | Product owner          |
| Q-HRS-03  | List of competitors and research results (8.1).                                                                                              | Business/Marketing     |
| Q-HRS-04  | Legal basis and retention period for employee data (attendance, leave, documents, audit) and mapping of NIK/NPWP classification to UU PDP categories. | Legal                  |
| Q-HRS-05  | Working hours/shift rules, lateness threshold, working days per week, geofence, time zone per tenant.                                        | Product owner          |
| Q-HRS-06  | Default leave policy (types, entitlement, carry-over, pro-rata) and the source of the national holiday and collective leave calendar.        | Product owner + Legal  |
| Q-HRS-07  | Availability/SLA target, p95 latency, RPO/RTO.                                                                                               | Engineering + Business |
| Q-HRS-08  | Should OpenTelemetry/Prometheus become an optional `kit` module (D11)?                                                                       | Product owner          |
| Q-HRS-09  | LLM provider, processing location, and policy for sending PII to third parties for module 08.                                                | Owner + Legal          |
| Q-HRS-10  | Schedule of the PPh21/BPJS/THR legal review and the tax rules design sprint (ADR-0004 prerequisite).                                         | Owner + Legal          |
| Q-HRS-11  | Parent domain of the `stery_session` cookie (D3) and entitlement grace period when the platform is down (D4, default 24 hours).             | Product owner          |
| Q-HRS-12  | Place and method of production hosting (`deploy/k8s` is empty; no working deploy target yet).                                                | Owner + Engineering    |
| Q-HRS-13  | Production SMTP provider and need for WhatsApp/push channels.                                                                                | Product owner          |
| Q-HRS-14  | Division of controller/processor roles and data processing agreement template with tenants.                                                  | Legal                  |
| Q-HRS-15  | Numeric KPI targets (section 11).                                                                                                            | Product owner          |

## 10. Risks and Mitigation

| **ID**       | **Risk**                                                                                                                                                                  | **Impact**                                            | **Mitigation**                                                                                                                                            | **Evidence**                                                                                               |
|--------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------|
| RISK-HRS-01  | Status documents claim 100% and production ready although the end-to-end flow is not proven                                                                               | Schedule and sales decisions based on wrong data      | BR-HRS-22; FSD 11.3 records the contradictions; status reported only with code evidence                                                                   | `PROJECT_PROGRESS.md`, `PHASE1_RELEASE.md`, `IMPLEMENTATION_COMPLETE.md`, `docs/PHASE5_SUMMARY.md`          |
| RISK-HRS-02  | Migrations fail from an empty database (attendance 001 invalid syntax, attendance 002 and leave 002 duplicate index names)                                                | Service cannot be installed; customers cannot onboard | Gate: all migrations `up -> down -> up` in CI on a fresh DB                                                                                               | `services/attendance-service/migrations/001_create_attendance_schema.up.sql`, `services/leave-service/migrations/002_add_performance_indexes.up.sql` |
| RISK-HRS-03  | No authentication/authorization in the services; tenant and actor are trusted from the request body                                                                       | Cross-tenant leak, forged approvals                   | Deny-by-default interceptor, `RequirePermission`, tenant from session not from body (FR-HRS-SEC.4, FR-HRS-SEC.5)                                          | `services/_shared/server/grpc_interceptors.go`, `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| RISK-HRS-04  | NIK, NPWP, date of birth stored as plain text; no UU PDP policy                                                                                                           | UU PDP violation, sanctions, loss of trust            | `kit/piicrypt` encryption, data classification, retention, data subject rights (FSD 8; BR-HRS-12)                                                         | `services/employee-service/migrations/001_create_employee_schema.up.sql`                                    |
| RISK-HRS-05  | Large rewrite scope during stack alignment (Laravel gateway, SSR, JWT, NATS, golang-migrate, sqlc, zap, MinIO, Redis)                                                     | Schedule slips, RLS regression                        | Domain and application reused; only infrastructure and interfaces rewritten; cross-tenant tests mandatory per module (FSD 2.7)                            | `go.work`, `apps/api-gateway/composer.json`                                                                 |
| RISK-HRS-06  | Small team: parallel work (stabilisation, alignment, suite) exceeds capacity                                                                                              | Long queues, quality drops                            | One alignment at a time (D12); HRIS after POS                                                                                                              | Impact and Risk Analysis, RISK-SUI-08 (suite document)                                                      |
| RISK-HRS-07  | Payroll/tax logic released in haste                                                                                                                                       | Regulatory violation for tenants                      | ADR-0004: merge prohibition; legal review mandatory; 501 routes until ready                                                                               | `docs/adr/0004-defer-payroll-to-phase2.md`                                                                  |
| RISK-HRS-08  | No working deploy target: `deploy/k8s` is empty, CD references Dockerfiles that do not exist, only two services have a Dockerfile                                         | Cannot release to staging/production                  | Deploy follows the SELA pattern (Caddy + compose) at alignment; hosting decision Q-HRS-12                                                                 | `.github/workflows/cd-staging.yml`, `Dockerfile.base`                                                       |
| RISK-HRS-09  | Domain events have empty payloads and consumers do not match publishers so the onboarding and notification chain does not work                                            | Cross-module features look done but do not run        | Outbox + versioned payload contract + publisher/consumer contract tests (FSD 2.5, 7.2)                                                                    | `services/leave-service/internal/infrastructure/nats/event_publisher.go`, `services/employee-service/internal/infrastructure/nats_consumer.go` |
| RISK-HRS-10  | Suite mode makes the platform a single point of failure for identity and entitlement                                                                                      | All HRS logins disrupted                              | 30-second Introspect cache, 24-hour entitlement snapshot, standalone unaffected (D-HRS-04)                                                                | Impact and Risk Analysis, RISK-SUI-11 (suite document)                                                      |
| RISK-HRS-11  | Prices, tiers, and billing model undefined                                                                                                                                | Cannot be sold                                        | Q-HRS-02 decided before S4                                                                                                                                | No evidence in the repository (value `tbd`)                                                                 |
| RISK-HRS-12  | The AI module processes PII to a third-party LLM without a policy                                                                                                         | Data leak, UU PDP violation                           | Module 08 not worked on before Q-HRS-09; per-tenant cost and rate limits                                                                                  | `proto/hris/ai/v1/ai.proto`                                                                                 |
| RISK-HRS-13  | Dev credentials and default passwords written in git-tracked files and as default values in code                                                                          | Misuse if carried into real environments             | Rotation and history cleanup; fail-fast configuration without secret defaults (FR-HRS-SEC.7)                                                              | `deploy/docker/docker-compose.yml`, `deploy/docker/nats.conf`, `deploy/docker/init-postgres.sql`, `docker-compose.production.yml` |
| RISK-HRS-14  | Notifications never sent (email adapter not wired) so users do not know about leave requests                                                                              | Leave approvals delayed, trust drops                  | FR-HRS-05.2 implemented and tested; queue + retry                                                                                                         | `services/notification-service/cmd/server/main.go`                                                          |

## 11. KPIs

Definitions only; all targets `tbd` (Q-HRS-15). Data source: the product database and audit trail.

| **KPI**                               | **Definition**                                                                                            | **Target** |
|---------------------------------------|-----------------------------------------------------------------------------------------------------------|------------|
| Active tenants                        | Number of tenants with >= 1 active employee and >= 1 activity (check-in or leave request) within 30 days  | `tbd`      |
| Tenant onboarding time                | Time from tenant registration until the first employee other than the admin is created                    | `tbd`      |
| Managed employees per tenant          | Average active employees per active tenant                                                                 | `tbd`      |
| Leave decision time                   | Median time from leave request until approved or rejected                                                 | `tbd`      |
| Attendance coverage                   | Percentage of active employees who have an attendance record on a working day                              | `tbd`      |
| Notification delivery success         | Percentage of notifications with delivered status versus those created                                     | `tbd`      |
| Cross-tenant incidents                | Number of cross-tenant access findings per release (reasonable target: zero)                              | `tbd`      |
| Stabilisation gate pass rate          | Number of HRIS gate items (section 12) passed versus total                                                 | `tbd`      |
| Trial-to-paid conversion              | Percentage of trial tenants that become paying customers                                                   | `tbd`      |
| Tenant churn                          | Percentage of tenants cancelling their subscription per period                                             | `tbd`      |

## 12. Timeline and Phases

All durations `tbd` (small team; estimates made after the owner sets priorities). Order follows decision D12.

| **Phase**                                          | **Scope**                                                                                                                                                | **Exit criteria (gate)**                                                                                                                                                                                                                  | **Duration** |
|----------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------|
| H0 Documentation (S0)                              | HRIS BRD/PRD/FSD; impact analysis                                                                                                                        | Documents approved by the owner                                                                                                                                                                                                           | `tbd`      |
| H1 HRIS stabilisation (S3, first part)             | Fix blocking defects in the current stack and close the following gates                                                                                  | (1) all migrations `up -> down -> up` on a fresh DB in CI; (2) real login through the API; (3) `UpdateEmployee` and org chart implemented; (4) document presign and rollback implemented with tests; (5) NIK/NPWP/date of birth encrypted with `kit/piicrypt`; (6) rate limits enforced; (7) secrets and binaries cleaned from the repository | `tbd`      |
| H2 Stack alignment (S3, second part)               | Laravel gateway -> Go HTTP adapter `/api/v1`; SSR -> SPA; JWT -> opaque session; NATS -> outbox; golang-migrate -> goose; sqlc -> pgx; zap -> slog; MinIO -> RustFS; Redis -> Dragonfly; services -> modules in one process | Cross-tenant tests pass in every module; `scripts/check.sh` green on the SELA pattern; coverage >= 80% (SELA target)                                                                                                       | `tbd`      |
| H3 Suite mode (S4)                                 | Suite adapters for the six ports; entitlement `hrs.employees.max`; platform console                                                                      | `standalone` mode unchanged; `suite` mode passes SSO and entitlement degradation tests                                                                                                                                                     | `tbd`      |
| H4 Suite integration (S5)                          | INT-1 and INT-2 if demand is validated                                                                                                                   | Market demand validation; versioned event contract                                                                                                                                                                                        | `tbd`      |
| H5 Phase 2 risky domains (outside this plan)       | Payroll, tax, recruitment, performance                                                                                                                   | ADR-0004 prerequisites complete (legal review, ADR, saga design)                                                                                                                                                                          | `tbd`      |

Correction note on gate (7): the `server.exe` file in the working tree is not tracked by git (`.gitignore:7` ignores `*.exe`), so what needs cleaning is git-tracked secrets and local artefacts; gate verification still checks that no binaries are tracked.

## 13. Current Implementation Status

Status based on static reading of the code (no build or tests run; see FSD 1.4). Branch `development`, 33 commits, last commit 10 June 2026. Detailed defects per FR are in FSD sections 4 and 11.

| **Module**           | **Description**                                                                                                                                              | **Status**  | **Evidence**                                                                                                                                                        |
|----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 01 AUTH              | 6 RPCs with handlers (login, refresh, validate, revoke, permissions, tenant registration); JWT RS256; no account lockout, interceptor, or user management     | Partial     | `services/auth-service/internal/interfaces/grpc/auth_service.go`, `services/auth-service/cmd/server/main.go`                                                        |
| 02 EMP               | 8 of 10 RPCs with handlers; `UpdateEmployee` and `GetOrgChart` not yet; SQL mismatches (employee number, `description` column, enum case) may make writes fail | Partial     | `services/employee-service/internal/interfaces/grpc/employee_service.go`, `services/employee-service/internal/infrastructure/postgres/employee_repository.go`      |
| 03 ATT               | 5 RPCs with handlers; migration 001 invalid; override fails for new records; no working hours rules                                                          | Partial     | `services/attendance-service/migrations/001_create_attendance_schema.up.sql`, `services/attendance-service/internal/application/commands/override_attendance.go`  |
| 04 LEV               | 8 RPCs with handlers; `ApplyLeave` does not send `days_count`; overlap and working days not enforced; no leave type creator                                 | Partial     | `services/leave-service/internal/interfaces/grpc/leave_service.go`, `services/leave-service/internal/domain/leave_request.go`                                       |
| 05 NTF               | 5 RPCs with handlers; notifications only stored (PENDING); email adapter not wired                                                                           | Partial     | `services/notification-service/cmd/server/main.go`, `services/notification-service/internal/infrastructure/channels/email_adapter.go`                               |
| 06 AUD               | 3 RPCs with handlers (`GetEntry` wrong); four consumers; no RLS and no immutability trigger                                                                  | Partial     | `services/audit-service/internal/interfaces/grpc/audit_service.go`, `services/audit-service/migrations/001_create_audit_schema.up.sql`                              |
| 07 DOC               | 4 of 7 RPCs are filled in; `DownloadDocument`, `GetPresignedUrl` and `RollbackDocumentVersion` are still placeholders returning empty responses; 0 test files; no malware scanning | Partial     | `services/document-service/internal/interfaces/grpc/document_service.go`, `services/document-service/internal/interfaces/grpc/mapping.go`                          |
| 08 AI                | Proto contract only (3 RPCs); no service code                                                                                                                | Not started | `proto/hris/ai/v1/ai.proto`, `go.work`                                                                                                                              |
| 09 PAY               | Deferred (ADR-0004)                                                                                                                                          | Deferred    | `docs/adr/0004-defer-payroll-to-phase2.md`                                                                                                                          |
| 10 TAX               | Deferred (ADR-0004)                                                                                                                                          | Deferred    | `docs/adr/0004-defer-payroll-to-phase2.md`                                                                                                                          |
| 11 REC               | Deferred (ADR-0004)                                                                                                                                          | Deferred    | `docs/adr/0004-defer-payroll-to-phase2.md`                                                                                                                          |
| 12 PRF               | Deferred (ADR-0004)                                                                                                                                          | Deferred    | `docs/adr/0004-defer-payroll-to-phase2.md`                                                                                                                          |

Supporting components:

| **Component**                   | **Description**                                                                                                                                       | **Status**  | **Evidence**                                                                                           |
|---------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------|-------------|--------------------------------------------------------------------------------------------------------|
| Laravel 11 gateway              | Only `auth/*` and leave routes; employee, attendance, document, audit routes are only comments; calls an HTTP path the Go services do not serve       | Partial     | `apps/api-gateway/routes/api.php`, `apps/api-gateway/app/Services/BaseGrpcGatewayClient.php`            |
| SvelteKit web (SSR)             | Four pages (login, dashboard, apply leave, admin leave); login mocked; no i18n                                                                         | Partial     | `apps/web/src/routes/(auth)/login/+page.server.ts`, `apps/web/src/hooks.server.ts`                      |
| CI/CD                           | Go CI: lint 7 services, unit and integration tests only auth, employee, leave; CD references Dockerfiles that do not exist and an empty `deploy/k8s`  | Partial     | `.github/workflows/ci-go.yml`, `.github/workflows/cd-staging.yml`                                       |
| Test files                      | 22 `_test.go` files in 6 services (document-service: 0); integration tests tagged with build tag `integration`                                        | Partial     | `services/leave-service/internal/domain/leave_request_test.go`, `services/employee-service/internal/integration/employee_integration_test.go` |
| Local infrastructure            | Compose for Postgres 16, Redis 7, NATS 2.10, MinIO, Jaeger, Prometheus, PgBouncer                                                                      | Done        | `deploy/docker/docker-compose.yml`                                                                      |

Summary: 7 of 8 service directories contain code (`ai-service` is empty); 39 of 47 proto RPCs have substantive handlers (the other eight: `UpdateEmployee`, `GetOrgChart`, three document RPCs, three AI RPCs). "Has a handler" does not mean "works": see FSD 4 and 11.

## 14. Approval

This document needs to be reviewed and approved before the HRIS-Stery PRD and FSD are used as the basis for work, including the out-of-scope items (4.4) that are still proposals.

| **Name** | **Title**        | **Signature**    | **Date**    |
|----------|------------------|------------------|-------------|
|          | Product Owner    |                  |             |
|          | Business Sponsor |                  |             |
|          | Engineering Lead |                  |             |
|          | Legal/Compliance |                  |             |
