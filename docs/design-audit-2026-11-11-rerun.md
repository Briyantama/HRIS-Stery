# HRIS-Stery web: design audit and critique, re-run (2026-11-11)

⚠️ DEGRADED (partial): Assessment A ran as an isolated sub-agent (`AuditHris.CritiqueA`, scout, saw no detector output and no baseline). Assessment B (detector) ran inline in the parent, before A returned, because it is a single deterministic command; nothing from it was passed to A. No browser render was run (read-only assignment, no dev server), so contrast and tap-target notes come from tokens and class heights, not pixels. Browser overlay: not run.

Scope: `apps/web/src` (SvelteKit 2.63, Svelte 5.56, Tailwind 4, `@stery/kit`). Mode: Operate. Context: existing `PRODUCT.md` + `DESIGN.md` (loaded via `impeccable context`). Baseline: `docs/design-audit-2026-11-11.md` (untouched).
Detector: `impeccable detect --json apps/web/src` returned `[]` (0 findings).

Note on the delta: the source files are the post-polish files the baseline describes (the same 4 files are modified in the working tree; no further HRIS edits since). So "fixed" items below were verified in code, and every previously skipped item is still open.

## Audit Health Score

| # | Dimension | Base | Now | Key finding |
|---|-----------|------|-----|-------------|
| 1 | Accessibility | 2 | 3 | Contrast, `aria-pressed`, live regions, `h1`, autocomplete fixed. Left: no landmarks/nav/skip link, heading skip h1 to h3, field errors not tied via `aria-describedby`, native `confirm()`/`alert()` |
| 2 | Performance | 3 | 3 | Lean; `/` still redirects client-side after paint (`+page.svelte:6-8`) |
| 3 | Theming | 4 | 4 | Token-only, no raw colors, dark mode by construction |
| 4 | Responsive | 3 | 3 | Mobile tap targets raised to 44px on filter/Approve/Reject/Submit/Cancel. Kit `Input`/`Select` stay 40px; 403 and login fill the viewport fine |
| 5 | Implementation Integrity | 3 | 3 | Detector clean; hand-rolled toast/tabs/empty state/page header still duplicate kit components; legacy (`$:`, `on:`, `<slot>`, `$app/stores`) mixed with runes |
| **Total** | | **15/20** | **16/20** | **Good** |

**Implementation Integrity verdict: pass with drift.** One coherent token system, consistent loading/error/empty/data states. Drift: kit `Toast`, `Tabs`, `ConfirmDialog`, `PageHeader`, `EmptyState`, `AppShell`, `SidebarNav` exist but pages hand-roll or omit them. Legacy Svelte syntax in 5 of 6 route files while login uses runes.

## Design Health Score (Nielsen, from isolated Assessment A)

| # | Heuristic | Base | Now | Key issue |
|---|-----------|------|-----|-----------|
| 1 | Visibility of status | 2 | 2 | Approve spinner shared by every row (`admin/leaves:323`); toast gone at 3 s |
| 2 | Match real world | 2 | 1 | UUIDs shown as names (`admin/leaves:273,294,305`); `en-US` dates; English-only; "Multi-tenant HR Management System" (`login:16`) |
| 3 | User control | 2 | 2 | Generic native confirm; no detail of what is being approved |
| 4 | Consistency | 2 | 2 | Native dialogs vs DESIGN "Don't"; literal `href="/dashboard"` (`admin/leaves:187`); "Go Home" and "Go to Dashboard" same target (`403:29,32`) |
| 5 | Error prevention | 2 | 1 | Submit disabled until valid so Zod messages never show; "Total Leave Days" shows 0 or negative when end < start (`apply:243-249`); Reject disabled with no explanation (`admin/leaves:404`) |
| 6 | Recognition over recall | 2 | 1 | No nav, no link to `/leaves/apply` or `/admin/leaves`; dashboard has no "Request leave" action |
| 7 | Flexibility/efficiency | 1 | 1 | Fixed `per_page` 50 (`lib/queries/leave.ts:47`), no pager; no search/bulk |
| 8 | Aesthetic/minimalist | 2 | 3 | Restrained; reason shown twice (`admin/leaves:267,344`) and boilerplate help block (`apply:293-302`) |
| 9 | Error recovery | 2 | 2 | Raw `data.error` / backend text (`login:25`), no retry |
| 10 | Help/documentation | 2 | 2 | Legend wrong (`dashboard:143`); hard-coded "Annual, Sick" help (`apply:297`) |
| **Total** | | **19/40** | **17/40** | **Acceptable first build** |

The 19 to 17 move is grader strictness, not a regression: the code did not change, and this score came from an independent reviewer that did not see the baseline. Items 2, 5, 6 were scored lower on the same known gaps; item 8 higher.

**Design specificity verdict:** generic. Kit composition with correct token use, nothing expresses the DESIGN.md "Tidy Ledger" idea, and it could be any admin. For Operate that is acceptable; the gap is task flow and content (names, nav, language), not ornament.

Strengths: status always stated in words; every async region has all four states in one repeated pattern; progress bar `aria-hidden` with numeric ratio beside it, `tabular-nums` on values.

## Findings (P0 0 · P1 5 · P2 5 · P3 5)

### P1
1. **No navigation / shell.** `(app)` has no `+layout.svelte`; nothing links dashboard, request, approval; no landmarks or skip link. Impact: core flows undiscoverable. WCAG 2.4.1, 1.3.1. Fix: `(app)/+layout.svelte` with kit `AppShell` + `SidebarNav`, approval item gated by `data.isManager` (from `(app)/+layout.server.ts:8`), and a primary "Request leave" button in the dashboard header (`dashboard/+page.svelte:33`). `/impeccable shape` then `layout`. *Open (baseline P2, raised).*
2. **Raw IDs as data.** `admin/leaves/+page.svelte:273` (`employee_id`), `:294` (`leave_type_id`), `:305` (`approved_by_id`). Fix: map type via `useLeaveTypes` (already used on dashboard:10); employee names need an API field. `/impeccable clarify`. *Open (baseline P1).*
3. **Native `confirm()`/`alert()` and shared spinner.** `admin/leaves:54,70`; `loading={$approveLeave.isPending}` at `:323` and `$rejectLeave.isPending` at `:332` spin every card. Fix: kit `ConfirmDialog` that names employee, type and dates; track pending id per row; surface errors with `Alert`; kit `Toaster` replaces the toast at `:200-208`. `/impeccable harden`. *Open (baseline P1) + spinner part new.*
4. **Dashboard legend contradicts PRD.** `dashboard:143` "Approved requests awaiting start date"; PRD 5.2 pending = awaiting approval. Remaining-color tiers (`dashboard:14-18`, 50% / 25%) are not explained. `/impeccable clarify`. *Open (baseline P2, raised).*
5. **Approval list silently capped.** `lib/queries/leave.ts:47` sends `per_page: 50`; `admin/leaves:354-357` prints "Showing 50 of N" with no next page, so request 51+ is unreachable for a manager. Fix: page controls using the response `page`/`per_page`/`total`. `/impeccable harden`. *New.*

### P2
6. **Apply form validation is silent.** `apply:273` `disabled={!isFormValid()}` hides Zod errors and is not keyboard-explicable; `handleFieldChange` (`:71-83`) re-checks order only on `end_date`, not when start changes (`:215`); no `aria-describedby`/`aria-invalid` linkage of messages (`:201,220,236`); mixed `*` and "(Optional)" marking (`:182,256`). Fix: enable submit, validate on submit and focus first invalid field. *Open (baseline P3, raised) + start-date gap new.*
7. **Negative or zero day count.** `apply:243-249` renders "Total Leave Days" when end < start. Fix: render only when end >= start (also real working days come from backend per PRD). *New.*
8. **Heading level skip.** `admin/leaves:267` card titles are `h3` under the page `h1` with no `h2`; DESIGN "One H1 Rule" says steps must not skip. Fix: `h2`. WCAG 1.3.1. *New.*
9. **Filter row should be kit `Tabs`.** `admin/leaves:211-230`, `aria-pressed` is the interim fix; no counts, no arrow keys. *Open (baseline).*
10. **Localisation absent.** English copy, `en-US` dates (`admin/leaves:137`), vs PRD NFR-HRS-12. *Open (baseline P3).*

### P3
11. **Legacy Svelte syntax mixed with runes:** `$app/stores` (`403:2`; use `$app/state`), `<slot />` (`+layout.svelte:16`), `$:` and `on:click`/`on:submit` (`admin/leaves:43,204,220`, `dashboard:11`, `apply:165`) vs `$props`/`$state` in `login:6`. *New.*
12. **`/` client redirect flash** (`+page.svelte:6-11`); use `redirect(307)` in `+page.server.ts`. *Open (baseline).*
13. **Copy/behaviour nits:** hard-coded "Annual, Sick" in help (`apply:297`, PRODUCT.md forbids invented defaults); header year from client clock (`dashboard:34`) while `balance.year` is shown at `:69`; duplicated "Go Home" target and `/403` vs inline denial (`403:29-33`, `admin/leaves:173-190`); raw `PENDING` enum shown (`admin/leaves:269`); reason repeated (`admin/leaves:267,344`). *Mix of open and new.*
14. **Login:** jargon subtitle (`login:16`), no forgot-password (FR-HRS-01.8 not built), tenant field not labelled as a slug, raw backend errors not tied to fields (`login:25,31`). *Open (baseline).*
15. **Manager may be offered Approve on own request** (`admin/leaves:317-337`); [INFERENCE] backend blocks it, UI does not hide it. *New, unverified.*

## Delta vs baseline

**Fixed (all 8 baseline "Polished" items verified in code):** status text contrast uses `text-foreground` on tint (`dashboard:16-18`, `admin/leaves:299-310`); two-segment usage bar (`dashboard:115-118`); filter buttons `type`/`aria-pressed`/`min-h-11` (`admin/leaves:217-221`); `max-md:min-h-11` on actions (`admin/leaves:324,333`, `apply:274,284`); toast `role="status"` + dismiss `type="button"` (`:201,204`); login `h1` + `autocomplete` (`login:15,56,70,85`); leave-types `aria-busy` and day-count `role="status"` (`apply:185,247`); `tabular-nums` (`dashboard:86,93,97,109,119`). Detector 0 before and after.

**Still open (baseline skipped):** names instead of IDs; native confirm/alert + hand-rolled toast; kit `Tabs`; no shell/nav; `/403` vs inline denial; dashboard hierarchy; legend wording; apply-form behavior (disabled CTA, calendar days, 1.5 s redirect, help block, mixed required marking); `/` redirect; enum casing/hard-coded hrefs; login copy/recovery; i18n; kit-side options (44px Button, soft Badge tone, two-segment Progress).

**New:** no pagination past 50 (P1); shared approve/reject spinner (P1, part of #3); negative day count; heading skip h3; start-date order not rechecked; Svelte-5 legacy syntax drift; hard-coded "Annual, Sick"; client-clock year header; no tier legend for remaining colors; Reject disabled without explanation (`admin/leaves:404`); [INFERENCE] own-request Approve.

**Score movement:** audit 15 to 16 (Accessibility +1 credited for the verified polish); Nielsen 19 to 17 (independent stricter grader, same code; see note above).

## Positive findings
Token-only styling and dark mode by construction; complete async states; labels/`invalid` wired; date parsing avoids UTC off-by-one (`admin/leaves:136`); progress bar not duplicated for screen readers; server-side role guard in hooks plus in-page guard.

## Recommended actions
1. **[P1] `/impeccable shape` then `/impeccable layout`:** `(app)` shell with nav, "Request leave" CTA, dashboard hierarchy.
2. **[P1] `/impeccable harden`:** kit `ConfirmDialog`/`Toaster`, per-row pending state, pagination, apply-form submit validation, i18n.
3. **[P1] `/impeccable clarify`:** names for IDs (needs API field), Pending legend, enum casing, tier legend.
4. **[P2] `/impeccable adapt`/`typeset`:** `h2` headings, kit `Tabs`.
5. **[P3] `/impeccable distill`:** drop duplicate reason and help block; modernise legacy Svelte syntax.
6. `/impeccable polish` as the final step.

## Applied (2026-11-11)

| File | Change |
|------|--------|
| `apps/web/src/routes/(app)/admin/leaves/+page.svelte` | Native `confirm()` replaced by kit `ConfirmDialog` (description copy unchanged: "Are you sure you want to approve this leave request?"); native `alert()` and hand-rolled toast replaced by kit `toast`/`Toaster` (same success messages; errors via `toast.error`); Approve spinner now per row (`approvingId === leave.id`), shared `isPending` spinner removed from Approve and Reject (the reject dialog keeps its own loading); kit `Pagination` driven by response `total`/`per_page` with `page` sent in the query (status change resets to page 1, last row removed on a later page steps back); card title `h3` to `h2`. (#3, #5, #8) |
| `apps/web/src/routes/(app)/dashboard/+page.svelte` | Legend: "Pending: Requests awaiting approval" per PRD 5.2. (#4, legend part) |
| `apps/web/src/routes/(app)/leaves/apply/+page.svelte` | "Total Leave Days" rendered only when end >= start. (#7) |
| `apps/web/src/routes/(app)/+layout.svelte` (new) | Kit `AppShell` + `SidebarNav`: skip link, `main` landmark, nav to `/dashboard`, `/leaves/apply`, and `/admin/leaves` for managers/hr_admin (via `$isManager`, covers `hr_admin` too). (#1) |

## Skipped

- #2 Raw IDs: employee name / `approved_by_id` need a new API field; leave type name could use `useLeaveTypes` but was left with the rest of the clarify pass.
- #1 "Request leave" CTA on the dashboard: covered by the nav item; no further hierarchy change.
- #3 `Alert` for approve errors: errors shown via toast instead.
- #6 Apply-form submit validation / `aria-describedby`: behavior change, needs a product decision.
- #9 Kit `Tabs` for the status filter, #10 i18n, #11-#15 (legacy syntax, `/` redirect, copy nits, login copy, own-request Approve): not in the P1 scope of this pass.
- Note: `/403` keeps its own `min-h-screen` and now renders inside the shell.
