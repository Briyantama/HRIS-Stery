# Product

<!-- impeccable:product-schema 1 -->

> Provenance: written by a delegated worker with no interview channel. Every fact below is taken from `docs/PRD_HRIS.en.md` (v1.0 Draft, 6 Nov 2026) or the code under `apps/web`; nothing was confirmed with a human. Items marked `tbd` are undecided in the PRD and stay undecided here.

## Platform

web

## Users

- **HR Admin (`hr_admin`)**: HR staff who register the company and manage employees, leave, attendance corrections and documents. Also the buyer. Needs fast onboarding, correct leave rules, corrections with a trail, UU PDP compliance.
- **Manager (`manager`)**: direct supervisor. Approves or rejects leave and views the team's attendance. Needs a clear approval queue and to see who is away.
- **Employee (`employee`)**: end user. Checks in/out, requests leave, views balances. Needs few steps, clear request status, Indonesian language, and a layout usable on a phone.
- **Platform Admin (Stery operator)**: manages tenants and the `hrs.employees.max` limit; not an `apps/web` end user.

Employee and manager flows are the ones the web app currently serves: leave balance, leave request, leave approval (PRD 5.2, 5.3).

## Product Purpose

HRIS-Stery (SKU `HRS`) is a multi-tenant HRIS for companies in Indonesia: one source of truth for employee data, attendance that can be corrected with a trail, and leave with clear balances and approvals. Each tenant sees only its own data; personal data is protected in line with UU PDP No. 27/2022. Success is tidy, accountable daily HR work (PRD 1, 9: onboarding time, leave decision time, attendance coverage; all targets `tbd`).

## Positioning

`tbd` per PRD section 8 (awaiting segment and competitor research). Differentiator hypotheses only, not claims: UU PDP-by-design employee data, standalone or bundled with POS and SELA under one account, an immutable audit trail, Indonesian language and WIB by default. Do not write positioning copy until the PRD does.

## Operating Context

- Two modes: `standalone` (default; own login, SPA on port 5173) and `suite` (platform-issued session, product switcher in `stery-suite/apps/web`, SPA on 5373). The same SPA serves both; standalone behaviour must not change in suite mode (BR-HRS-15).
- Login takes organization (tenant slug), email and password; the session is an opaque HttpOnly cookie. No tokens in browser storage.
- Deny-by-default authorisation lives in the backend; the web app's role checks (`manager`, `hr_admin`) are presentation only.
- Time zone WIB, currency IDR.

## Capabilities and Constraints

- Web today (PRD 14): login, leave balance dashboard, leave request form, manager leave approval portal with required rejection reason, 403 page. Attendance, employee master data, org chart, documents, audit trail, notifications are in the PRD but not yet in `apps/web`.
- Leave flow: request (type, dates, optional reason) is `PENDING`; approve moves pending days to used; reject requires a reason and restores the balance; requester cannot approve their own request.
- Terminology: tenant/organization, employee, leave type, entitled / used / pending / remaining days, `PENDING` / `APPROVED` / `REJECTED`.
- Deliberately out of scope this phase: payroll, tax (PPh21, BPJS, THR), recruitment, performance, WhatsApp/push, native mobile app. Never imply payroll works.
- Working hours, leave entitlement defaults and pricing are `tbd`; show real values from the API, never invented defaults.
- The interface is specified as bilingual `{id,en}` with a catalogue parity test (NFR-HRS-12). `apps/web` copy is currently English-only; localisation is an open item, not something design work should paper over.

## Brand Commitments

Product name "HRIS-Stery", part of the Stery Suite (with POS and SELA). Visual identity comes from the shared `@stery/kit` design system; apps brand by overriding kit variables, not by forking components. No logo or voice guide exists in the repo.

## Evidence on Hand

- `docs/PRD_HRIS.en.md`, `docs/BRD_HRIS.*.md`, `docs/FSD_HRIS.*.md` for requirements.
- No user research, testimonials, customers or benchmarks exist (PRD flags personas as hypotheses, Q-HRS-01/03). Do not fabricate any.

## Product Principles

1. **Correct data first.** Never show a state the data does not support; no claimed feature without code evidence (BR-HRS-22).
2. **Secure and auditable by default.** Sensitive actions (approve, reject, correct) are explicit and traceable; the UI never hints that data crosses tenants.
3. **No invented rules.** Undecided business values stay `tbd` and visibly empty rather than guessed.
4. **Work first, in the user's language.** Employees finish a request in few steps on a phone; managers clear a queue quickly. Indonesian is the default audience language.
5. **Standalone and suite are one product.** Nothing in the UI may assume suite chrome is present.

## Accessibility & Inclusion

NFR-HRS-17: usable with keyboard and screen reader in the main flows; WCAG level is `tbd`. Employees may be on a phone, so touch use is a primary case.
