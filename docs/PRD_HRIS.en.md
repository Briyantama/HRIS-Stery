# PRD HRIS-Stery

*Product Requirements Document (PRD): Multi-tenant HRIS for companies in Indonesia: tidy, secure, auditable employee data, attendance, and leave*

|                         |                                                                                                                                                                                                              |
|-------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Document Name**       | PRD - HRIS-Stery (SKU `HRS`)                                                                                                                                                                                 |
| **Version**             | 1.0 (Draft)                                                                                                                                                                                                  |
| **Date**                | 6 November 2026                                                                                                                                                                                              |
| **Prepared by**         | Product & Engineering Team                                                                                                                                                                                   |
| **Status**              | Draft for review                                                                                                                                                                                             |
| **Related Documents**   | BRD HRIS-Stery (`BRD_HRIS.id.md`), FSD HRIS-Stery (`FSD_HRIS.id.md`), `00_Analisis_Dampak_Risiko`, BRD/PRD/FSD Stery Suite (folder `stery-suite/docs`), `SELA/CLAUDE.md` (stack standard), ADR-0001 to ADR-0004 |
| **Changes from previous** | First issue                                                                                                                                                                                                |

## 1. Vision and Summary

HRIS-Stery makes the daily HR work of companies in Indonesia tidy and accountable: a single source of truth for employee data, attendance that is recorded and can be corrected with a trail, and leave with clear balances and approvals. Each company (tenant) sees only its own data; employees' personal data is protected in line with UU PDP No. 27/2022.

The product is sold as SKU `HRS`: standalone (mode `standalone`, default) or connected to Stery Suite (mode `suite`) together with POS and SELA. Payroll, tax, recruitment, and performance are deliberately deferred (ADR-0004) until a legal review is done; this must be stated clearly in the positioning because ADR-0004 records the consequence: tenants cannot yet run payroll through the system.

Product principles:

- **Correct data first**: no feature is claimed done without code evidence (BR-HRS-22).
- **Secure by default**: tenant isolation in the database, authorisation in the backend, encrypted PII (BR-HRS-02, BR-HRS-03, BR-HRS-12).
- **Standalone and suite**: standalone behaviour is not changed by suite mode (BR-HRS-15).
- **No invented rules**: business values not yet decided (working hours, leave entitlement, price) are `tbd` with an open question.

## 2. Problem and Target Users

### 2.1 Core Problem

Hypotheses (not yet validated by user research `[INFERENCE]`; Q-HRS-01, Q-HRS-03):

- Employee data is scattered across spreadsheets and messages; there is no definitive organisation structure.
- Leave is requested via chat; balances are calculated manually; decisions are hard to trace.
- Attendance is recorded inconsistently; corrections leave no trail.
- Employee documents and identity data (NIK, NPWP) are stored without access control and without a UU PDP policy.
- Companies that use other Stery products (POS, SELA) must log in to each product with separate accounts (suite context).

### 2.2 User Personas

| **Persona**                        | **Description**                                                                                                 | **Main Needs**                                                                                                                     | **Price Sensitivity** |
|------------------------------------|-----------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------|------------------------|
| HR Admin (`hr_admin`)              | HR staff who register the company, manage employees, leave, attendance corrections, documents; also the buyer/decision maker   | Fast onboarding, employee import, correct leave rules, corrections with a trail, reports, UU PDP compliance                        | `tbd` (Q-HRS-02)       |
| Manager (`manager`)                | Direct supervisor; approves leave and views their team's attendance                                              | A clear approval queue, notifications, seeing who is on leave, team chart                                                          | Does not pay directly |
| Employee (`employee`)              | End user; checks in/out, requests leave, views balances                                                          | Few steps, clear request status, Indonesian language, usable on a phone                                                            | Does not pay directly |
| Platform Admin (Stery operator)    | Stery staff who manage tenants, entitlements, support, and view service health (suite mode via console)          | See tenants and usage, set the `hrs.employees.max` limit, no access to employee data without an audit trail                       | Not relevant          |

## 3. User Stories

| **ID**    | **As a...**                 | **I want to...**                                                           | **So that...**                                                                           | **Priority** |
|-----------|-----------------------------|----------------------------------------------------------------------------|------------------------------------------------------------------------------------------|---------------|
| US-HRS-01 | HR Admin                    | register the company and the first admin account myself                    | I can use the product immediately without waiting for technical help                     | Must          |
| US-HRS-02 | HR Admin                    | log in with email and password with a secure session                       | company data cannot be accessed by others                                                | Must          |
| US-HRS-03 | HR Admin                    | add users, assign roles, and deactivate them                               | access rights match duties and departed employees' accounts cannot be used               | Must          |
| US-HRS-04 | HR Admin                    | create hierarchical departments and positions                              | the company's organisation structure is represented                                      | Must          |
| US-HRS-05 | HR Admin                    | add employees one by one with automatic employee numbers                   | employee master data is complete and unique                                              | Must          |
| US-HRS-06 | HR Admin                    | import many employees from a CSV file                                      | onboarding a large company does not take days                                            | Should        |
| US-HRS-07 | HR Admin                    | edit employee data and terminate departing employees                       | data is always current and departed employees' access is closed automatically            | Must          |
| US-HRS-08 | HR Admin / Manager          | view the organisation chart                                                | I know the reporting lines and who reports to whom                                       | Must          |
| US-HRS-09 | Employee                    | check in and check out with a single tap                                   | my attendance is recorded without paper                                                  | Must          |
| US-HRS-10 | HR Admin                    | correct attendance records with a reason                                   | mistakes are fixed and the change can be audited                                         | Must          |
| US-HRS-11 | Manager                     | view my team's attendance and summary                                      | I know who is present, late, or on leave                                                 | Should        |
| US-HRS-12 | HR Admin                    | set the tenant's working hours, lateness threshold, and working days       | attendance status is calculated according to company policy (values `tbd`, Q-HRS-05)     | Should        |
| US-HRS-13 | Employee                    | request leave with a type, dates, and a reason                             | my request is recorded and processed                                                     | Must          |
| US-HRS-14 | Employee                    | view my leave balance per type and request history                         | I know my remaining leave and request status                                             | Must          |
| US-HRS-15 | Manager                     | approve or reject team members' leave with a reason                        | decisions are fast and recorded; the balance is updated automatically                    | Must          |
| US-HRS-16 | Employee                    | cancel a leave request                                                     | the balance is restored if plans change                                                  | Must          |
| US-HRS-17 | HR Admin                    | configure leave types and annual entitlement                               | the company's leave policy is represented (default values `tbd`, Q-HRS-06)               | Must          |
| US-HRS-18 | Manager / Employee          | receive email and in-app notifications for requests and decisions          | I do not have to keep checking                                                           | Should        |
| US-HRS-19 | HR Admin                    | trace the audit trail (who, what, when)                                    | I can prove the history of actions in case of a dispute or inspection                    | Must          |
| US-HRS-20 | HR Admin                    | upload and download employee documents securely                            | sensitive documents are not leaked and can only be opened by authorised parties          | Should        |
| US-HRS-21 | Employee                    | request access to or correction of my personal data                        | my rights as a UU PDP data subject are fulfilled                                         | Must          |
| US-HRS-22 | Platform Admin              | view tenants and set the `hrs.employees.max` limit                         | per-tier sales can be enforced                                                           | Should        |
| US-HRS-23 | Organisation Owner (suite)  | open HRIS from the product switcher without logging in again               | I use several Stery products with one account                                            | Should        |
| US-HRS-24 | HR Admin                    | be warned when approaching the plan's employee limit                       | I can upgrade before employee creation is rejected                                       | Could         |
| US-HRS-25 | HR Admin                    | request AI-assisted resume analysis                                        | candidate screening is faster (after the LLM PII policy is decided, Q-HRS-09)            | Could         |
| US-HRS-26 | HR Admin                    | run payroll (PPh21, BPJS, THR) in the system                               | salaries and taxes are calculated automatically                                          | Won't         |

## 4. Product Features (MoSCoW Priority)

Features refer to FSD functional requirements (FR). The implementation status per FR is in FSD 11.1.

### 4.1 Must Have

- Self-service tenant registration with a first admin and default roles; password login with a secure opaque session; user and role management (FR-HRS-01.1 to FR-HRS-01.7).
- Master data: employees (create, view, edit, terminate), hierarchical departments, positions, organisation chart, encrypted PII (FR-HRS-02.1 to FR-HRS-02.8).
- Attendance: check-in, check-out, HR correction with a reason, attendance list (FR-HRS-03.1 to FR-HRS-03.4).
- Leave: leave types, balances, requests with automatic working days and no overlap, approval, rejection, cancellation, list and balance (FR-HRS-04.1 to FR-HRS-04.7).
- Persisted multi-channel notifications and real email delivery (FR-HRS-05.1, FR-HRS-05.2).
- Append-only immutable audit trail, filtered queries, and tenant isolation (FR-HRS-06.1, FR-HRS-06.2, FR-HRS-06.4, FR-HRS-06.6, FR-HRS-06.7).
- Documents: upload via signed URL, fail-closed malware scanning, short-lived download, size and type limits (FR-HRS-07.1 to FR-HRS-07.4, FR-HRS-07.8).
- Cross-module security: full RLS, cross-tenant tests, encrypted PII, opaque sessions, deny-by-default, rate limiting, secrets, UU PDP data subject rights (FR-HRS-SEC.1 to FR-HRS-SEC.8, FR-HRS-SEC.10).
- HRIS stabilisation gate (section 11) and stack alignment (BR-HRS-17, BR-HRS-18).
- Complete `standalone` mode and `suite` mode through adapters (FR-HRS-01.10; section 10).

### 4.2 Should Have

- Bulk CSV employee import (FR-HRS-02.9); automatic employee shell from registration (FR-HRS-02.10); `hrs.employees.max` limit validation (FR-HRS-02.11).
- Attendance summary, working-hours and lateness rules, tenant time zone (FR-HRS-03.5, FR-HRS-03.6, FR-HRS-03.8).
- National holiday and collective leave calendar (FR-HRS-04.8).
- Event consumption for notifications, list and mark as read (FR-HRS-05.3, FR-HRS-05.4).
- Fetch a single audit entry, audit retention and partitioning (FR-HRS-06.3, FR-HRS-06.5).
- Document versions, rollback, delete and retention, classification (FR-HRS-07.5 to FR-HRS-07.7).
- Email verification and password recovery (FR-HRS-01.8); client IP logging (FR-HRS-01.9).
- Security headers, CORS, CSRF; data retention and deletion (FR-HRS-SEC.9, FR-HRS-SEC.11).
- Bilingual interface (BR-HRS-19).

### 4.3 Could Have

- Attendance-leave integration (leave days mark `ON_LEAVE`) (FR-HRS-03.7); manual balance adjustment, carry-over, pro-rata (FR-HRS-04.9).
- Per-tenant channel configuration and notification opt-out preferences (FR-HRS-05.5, FR-HRS-05.7).
- AI assistant: resume analysis, attrition prediction, HR chat (FR-HRS-08.1 to FR-HRS-08.4).
- Suite integration hypotheses INT-1 and INT-2 (BR-HRS-21; "hypothesis, validate demand").
- Warning when approaching the plan limit (US-HRS-24).

### 4.4 Won't Have (this phase)

- Payroll, tax rules, PPh21, BPJS, THR, severance (modules 09, 10; ADR-0004).
- Recruitment and performance (modules 11, 12; ADR-0004).
- WhatsApp and push notifications (FR-HRS-05.6; deferred).
- Native mobile app (proposal, BRD 4.4).

## 5. User Flow

### 5.1 Main Flow: Tenant Onboarding to Attendance

1. **Tenant registration**: the HR Admin enters the company name, slug (3-50 characters `[a-z0-9-]`), email, and password (8-128 characters, uppercase, lowercase, digit, symbol). The system creates the tenant, three system roles (`hr_admin`, `manager`, `employee`), default permissions, and the first admin.
2. **First login**: the admin logs in; the opaque session is stored in an HttpOnly cookie; the landing screen shows the onboarding steps.
3. **Structure**: the admin creates departments and positions (FR-HRS-02.5, FR-HRS-02.6).
4. **Employees**: the admin adds employees one by one or imports a CSV (FR-HRS-02.1, FR-HRS-02.9). Each new employee triggers creation of an annual leave balance and a welcome notification.
5. **Accounts**: the admin invites employees to become users with a role (FR-HRS-01.7).
6. **Attendance**: employees check in and check out every working day (FR-HRS-03.1, FR-HRS-03.2); HR corrects when needed (FR-HRS-03.3).

### 5.2 Leave Approval Flow

1. The employee chooses the leave type, start and end dates, and a reason; the system calculates working days and checks the balance and overlap.
2. The request has status `PENDING`; the pending balance (`pending_days`) increases; the supervisor receives a notification.
3. The supervisor or HR opens the queue and approves or rejects (a reason is required when rejecting); the requester cannot approve their own request.
4. When approved, the pending balance moves to used; when rejected, the pending balance is restored; the requester receives a notification.
5. The requester or HR can cancel; cancellation restores the balance according to the previous status.

### 5.3 Manager Flow

The dashboard shows the team's leave approval queue, the team's attendance today, and the organisation chart (FR-HRS-02.7, FR-HRS-03.4, FR-HRS-04.6).

### 5.4 Platform and Organisation Admin Flow (suite mode)

1. The organisation owner chooses a package on the platform; the platform activates the `hrs.*` entitlement.
2. The user opens HRIS from the product switcher; the `stery_session` session is used without logging in again.
3. The platform admin views tenants and sets `hrs.employees.max`; when the limit is reached or the entitlement expires, only creation of new employees is rejected and data remains readable (BR-HRS-16).
4. Migrating a standalone customer to suite: link the suite organisation with the existing HRS tenant (FSD Stery Suite 7.5).

## 6. Non-Functional Requirements

All numeric targets are `tbd` (Q-HRS-07, Q-HRS-15) until the owner decides. The technical realisation is in FSD 9.

| **ID**       | **Category**                  | **Requirement**                                                                                                                                           | **Target** |
|--------------|-------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|------------|
| NFR-HRS-01   | Tenant isolation              | Every table with `tenant_id` uses RLS ENABLE+FORCE with `USING` and `WITH CHECK`; cross-tenant tests are mandatory in every module                         | 100% tenant tables |
| NFR-HRS-02   | Session security              | Opaque session in an HttpOnly Secure cookie; only the hash is stored server-side; no JWT and no tokens in browser storage                                  | `tbd` (session TTL)  |
| NFR-HRS-03   | Authorisation                 | Deny-by-default: every endpoint has an explicit permission checked in the backend; tenant and actor come from the session, not from the request body       | 100% endpoints       |
| NFR-HRS-04   | PII protection                | NIK, NPWP, birth date are encrypted with AES-256-GCM with a blind index for search; separate keys per environment                                          | 100% class 2 PII columns |
| NFR-HRS-05   | Auditability                  | Append-only audit trail with UPDATE/DELETE/TRUNCATE rejected at the database level                                                                          | 100% sensitive actions |
| NFR-HRS-06   | Availability                  | Service availability target                                                                                                                                 | `tbd`                |
| NFR-HRS-07   | Performance                   | p95 latency of main read/write operations (login, employee list, check-in, leave request)                                                                   | `tbd`                |
| NFR-HRS-08   | Rate limiting                 | Rate limiting of login and sensitive endpoints per address and per account; account lockout after repeated failed attempts                                  | `tbd` (threshold)    |
| NFR-HRS-09   | Migration                     | All migrations are reversible; CI runs `up -> down -> up` on a fresh DB                                                                                     | 100% migrations      |
| NFR-HRS-10   | Code quality                  | Go and web test coverage >= 80% (SELA floor); `-race` in CI; a single `scripts/check.sh` script the same as SELA                                            | >= 80%               |
| NFR-HRS-11   | Observability                 | `slog` JSON logs without PII and secrets; `/healthz`; OpenTelemetry/Prometheus is an owner decision `tbd` (Q-HRS-08)                                         | baseline D11         |
| NFR-HRS-12   | Localisation                  | `{id,en}` interface with a catalogue parity test; WIB; IDR                                                                                                  | 100% strings         |
| NFR-HRS-13   | Configuration and secrets     | Configuration from environment variables with fail-fast validation; no default secrets in code or git-tracked files                                         | 0 secrets in repo    |
| NFR-HRS-14   | Backup and recovery           | Scheduled database backups and restore tests                                                                                                                | RPO/RTO `tbd`        |
| NFR-HRS-15   | Retention and deletion        | Retention of each data class and deletion/anonymisation at the deadline (UU PDP)                                                                            | `tbd` (Q-HRS-04)     |
| NFR-HRS-16   | Mode portability              | `standalone` and `suite` are selected through `STERY_MODE` without changing standalone behaviour                                                            | 0 standalone regressions |
| NFR-HRS-17   | Accessibility                 | The interface is usable with keyboard and screen reader in the main flows                                                                                   | `tbd` (WCAG level)   |
| NFR-HRS-18   | Event reliability             | At-least-once events with a transactional outbox and idempotent consumers (`processed_events`)                                                              | 0 lost events        |

## 7. Monetisation Model (Product Summary)

Refers to BRD HRIS 7: a single limit dimension `hrs.employees.max`, tiers and prices `tbd` (Q-HRS-02). Suite mode billing by the platform (D5); standalone billing `tbd`. Payroll and deferred modules are potential add-ons.

## 8. Positioning

`tbd` (awaiting segment and competitor research, Q-HRS-01, Q-HRS-03). Differentiator hypotheses to be tested: (1) employee data protected in line with UU PDP by design; (2) standalone or bundled with POS and SELA under one account; (3) an immutable audit trail; (4) Indonesian language and WIB by default. No positioning statement is written until there is data.

## 9. Success Metrics

KPI definitions are in BRD HRIS 11; all targets `tbd` (Q-HRS-15). Product metrics measured: active tenants, tenant onboarding time, leave decision time, attendance coverage, notification delivery success, cross-tenant incidents, and stabilisation gate pass.

## 10. Operating Modes

| **Aspect**                | **`standalone` (default)**                                             | **`suite`**                                                                                                 |
|---------------------------|------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------|
| Identity and session      | HRIS's own users and passwords (bcrypt); opaque HRIS cookie session    | The platform issues the `stery_session` session; HRIS validates through `SessionService/Introspect` (30-second cache) |
| Tenant                    | HRIS-owned `tenants`                                                   | The platform organisation is linked to the HRS tenant (`external_ref`)                                      |
| Entitlement               | No local commercial restriction (`tbd`; Q-HRS-02)                      | `EntitlementService/Check` and a signed snapshot; fail-closed for paid features after 24 hours              |
| Billing                   | None (`tbd`)                                                           | Platform (Midtrans QRIS through `payment.Provider`)                                                         |
| Notification              | HRIS's own notification service                                        | Platform notification hub through the `Notification` port                                                   |
| Audit                     | HRIS's own audit trail                                                 | Still written locally and forwarded to the platform audit store through the `Audit` port                    |
| Cross-product events      | None                                                                   | Outbox -> `IntegrationService/Publish`; opt-in per organisation through features `integration.hrs_pos`, `integration.hrs_sel` (proposal) |
| Web                       | HRIS SPA on port 5173 (standalone profile)                             | The same SPA; product switcher and console in `stery-suite/apps/web`                                        |
| Default ports             | API 8080, gRPC 9090, web 5173                                          | API 8280, gRPC 9290, web 5373                                                                               |

Product Contract (what HRIS fulfils to become a suite product): manifest (`product_code` `hrs`, module feature keys, dimension `hrs.employees.max`), six ports (`Identity`, `Tenant`, `Entitlement`, `Notification`, `Audit`, `Events`) with two adapters each, `STERY_MODE`, `/healthz`, outbox publishing, a STACK-SPEC conformance checklist, and the BRD/PRD/FSD triple (this document is part of that triple). Details in FSD Stery Suite 4.08.

## 11. Release Roadmap (Indicative)

Dates `tbd`. The order follows D12.

| **Release** | **Focus**                          | **Main features / exit criteria**                                                                                                                                                                                                                                                   |
|-----------|------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| R0        | Documents (S0)                     | HRIS BRD/PRD/FSD approved                                                                                                                                                                                                                                                          |
| R1        | HRIS stabilisation (S3)            | Gate: (1) all migrations `up -> down -> up` on a fresh DB in CI; (2) real login through the API; (3) `UpdateEmployee` and org chart implemented; (4) document presign and rollback implemented with tests; (5) NIK/NPWP/birth date encrypted with `kit/piicrypt`; (6) rate limiting enforced; (7) secrets and binaries cleaned up |
| R2        | Stack alignment (S3)               | Go HTTP adapter `/api/v1` replaces the Laravel gateway; SPA; opaque sessions; outbox; goose; pgx; slog; RustFS; Dragonfly; modules in a single process; cross-tenant tests per module                                                                                              |
| R3        | Suite mode (S4)                    | Six-port suite adapters; `hrs.employees.max` entitlement; SSO; entitlement degradation                                                                                                                                                                                             |
| R4        | Suite integration (S5)             | INT-1 and INT-2 if demand is validated                                                                                                                                                                                                                                             |
| R5        | Phase 2 (outside this plan)        | Payroll, tax, recruitment, performance after the ADR-0004 prerequisites                                                                                                                                                                                                            |

## 12. Out of Scope

- Payroll, tax, BPJS, THR, severance (ADR-0004).
- Recruitment and performance (ADR-0004).
- WhatsApp and push.
- Native mobile app; integration with government systems.
- Automatic standalone billing inside HRIS.

## 13. Dependencies and Risks

| **Dependency**                                               | **Impact**                                                            | **Related Risk** |
|--------------------------------------------------------------|-----------------------------------------------------------------------|--------------------|
| `kit` suite (tenantdb, session, outbox, piicrypt, audit, httpx, config, db, testdb, objstore, health) | Alignment cannot start before S1 is done | RISK-HRS-05, RISK-HRS-06 |
| `stery-suite` platform (IAM, ENT, BIL)                       | Suite mode can only be tested once the platform is available           | RISK-HRS-10        |
| Legal review (UU PDP, PPh21/BPJS/THR)                        | Blocks the full PII module and the payroll phase                       | RISK-HRS-04, RISK-HRS-07 |
| Owner decisions: price, segment, hosting, SMTP, LLM          | Blocks sales and modules 05 and 08                                     | RISK-HRS-11, RISK-HRS-08, RISK-HRS-12 |
| Production SMTP provider                                      | Email notification delivery                                            | RISK-HRS-14        |
| Stabilisation gate and migration fixes                       | All end-to-end flows                                                   | RISK-HRS-02, RISK-HRS-09 |

Full risk list: BRD HRIS 10.

## 14. Current Status

Based on a static read of the code as of 6 November 2026 (without running builds/tests). Status per requirement: FSD 11.1; contradictions between status documents and code: FSD 11.3.

| **Area**                         | **Status** | **Summary**                                                                                                                                    | **Evidence**                                                                                                                   |
|----------------------------------|------------|------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------|
| 01 AUTH                          | Partial    | Registration, login, refresh, revoke, permissions exist; no account lockout, no user management, no authentication interceptor                 | `services/auth-service/internal/interfaces/grpc/auth_service.go`                                                                |
| 02 EMP                           | Partial    | Create/view/list/terminate exist; update and org chart not yet; risk of SQL write failure                                                      | `services/employee-service/internal/interfaces/grpc/employee_service.go`                                                        |
| 03 ATT                           | Partial    | Check-in/out exist; migration invalid; no working-hours rules                                                                                  | `services/attendance-service/migrations/001_create_attendance_schema.up.sql`                                                    |
| 04 LEV                           | Partial    | State machine and balances exist; requests do not send `days_count`; overlap not enforced                                                      | `services/leave-service/internal/interfaces/grpc/leave_service.go`                                                              |
| 05 NTF                           | Partial    | Stored as PENDING; never sent                                                                                                                  | `services/notification-service/cmd/server/main.go`                                                                              |
| 06 AUD                           | Partial    | Consumer and queries exist; RLS disabled; no immutability trigger                                                                              | `services/audit-service/migrations/001_create_audit_schema.up.sql`                                                              |
| 07 DOC                           | Partial    | Metadata upload and presigned PUT; download/presign/rollback are placeholders                                                                  | `services/document-service/internal/interfaces/grpc/document_service.go`                                                        |
| 08 AI                            | Not started | Proto only                                                                                                                                    | `proto/hris/ai/v1/ai.proto`                                                                                                     |
| 09-12 PAY, TAX, REC, PRF         | Deferred   | ADR-0004                                                                                                                                       | `docs/adr/0004-defer-payroll-to-phase2.md`                                                                                      |
| Web                              | Partial    | Four pages; login is mocked to `http://localhost:3000/api/auth/login`                                                                          | `apps/web/src/routes/(auth)/login/+page.server.ts`                                                                              |
| Gateway                          | Partial    | `auth/*` and leave routes; the HTTP path to the services is not served                                                                         | `apps/api-gateway/routes/api.php`                                                                                               |
