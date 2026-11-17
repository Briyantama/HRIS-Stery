# HRIS-Stery web: design audit and critique (2026-11-11)

Scope: `apps/web/src` (SvelteKit 2, Svelte 5, Tailwind 4, `@stery/kit`). Mode: Operate. Context: `PRODUCT.md`, `DESIGN.md` (both written in this run; PRODUCT.md is inferred from `docs/PRD_HRIS.en.md`, no human interview was possible).

Method: dual-agent critique (A: design review with no detector output; B: detector run) plus parent verification of every cited line. The impeccable detector (`detect --json apps/web/src`) returned `[]` before and after polish. No browser render was run (no build/dev server per assignment), so contrast notes are computed from tokens, not measured on screen; touch-target notes come from kit class heights.

## Audit Health Score

| # | Dimension | Score | Key finding |
|---|-----------|-------|-------------|
| 1 | Accessibility | 2 | `text-warning`/`text-success` text on tints; filter "tabs" had no state semantics; toast had no live region; native `confirm()`/`alert()`; no landmarks/nav (no `(app)` shell) |
| 2 | Performance | 3 | Lean; client `onMount` redirect on `/` causes a flash |
| 3 | Theming | 4 | Kit tokens used everywhere, no raw colors; dark mode works by construction |
| 4 | Responsive | 3 | Grids collapse correctly; action/filter controls were 32-40px tall |
| 5 | Implementation Integrity | 3 | Detector clean; hand-rolled toast/tabs/empty-state/page header duplicate kit components |
| **Total** | | **15/20** | **Good** |

**Implementation Integrity verdict: pass with drift.** The pages express one coherent system (kit tokens, kit components, consistent states). Drift: several kit components (`Toast`/`Toaster`, `Tabs`, `ConfirmDialog`, `PageHeader`, `StatCard`, `EmptyState`, `Progress`, `AppShell`, `SidebarNav`) exist but pages hand-roll the equivalents. Detector: 0 findings.

## Design Health Score (critique, Nielsen)

| # | Heuristic | Score | Key issue |
|---|-----------|-------|-----------|
| 1 | Visibility of status | 2 | Toast lacked live region (fixed); bare "Redirecting..." page |
| 2 | Match with real world | 2 | Manager sees raw `employee_id` and `leave_type_id`; dashboard legend text for Pending contradicts admin meaning |
| 3 | User control | 2 | Native `confirm()`/`alert()`; form resets and redirects after 1.5 s |
| 4 | Consistency | 2 | Hand-rolled kit equivalents; `/403` page vs inline denial in admin |
| 5 | Error prevention | 2 | Day count is calendar days, not working days, and not checked against balance |
| 6 | Recognition over recall | 2 | Admin cards title by reason, not by person; apply page does not show balance |
| 7 | Flexibility/efficiency | 1 | No search, filter by person/date, bulk approve or pagination |
| 8 | Aesthetic/minimalist | 2 | Redundant help block and legend; five data points per balance card |
| 9 | Error recovery | 2 | Raw `error.message`, no retry action |
| 10 | Help/documentation | 2 | Generic inline help |
| **Total** | | **19/40** | Acceptable first build |

**Design specificity verdict:** plain and category-generic, which is acceptable for Operate; the gap is functional specifics (names instead of IDs, a navigation shell, Indonesian-language UI per PRD NFR-HRS-12), not ornament. Strengths: consistent semantic tokens and global focus ring; sound form wiring (label `for`/`id`, `invalid` -> `aria-invalid`, button `loading` -> `aria-busy`, Alert role picks `alert`/`status`); all four async states (loading, error, empty, data) present; timezone-safe date parsing in admin.

Persona red flags: **Manager** approves a list of UUIDs on a phone with a native confirm as the only guard. **Employee on mobile** cannot see remaining balance on the request form and sees the confirmation for 1.5 s. **HR Admin** has no cross-employee view, search, export or paging past the first page.

## Findings

Severity P0-P3. "Polished" = fixed in this run; "Skipped" = reason given.

### Polished

- **[P1] Status text on tinted surfaces fails contrast.** `dashboard/+page.svelte` Remaining badge used `text-success|warning|destructive` on a 10-15% tint of itself; the Pending value and `admin/leaves` "Awaiting approval" used `text-warning` (oklch 0.72 on white is roughly 2:1). Fix: text is `text-foreground` (muted for secondary), meaning carried by tint + 1px toned border; no copy change. Category: Accessibility. WCAG 1.4.3.
- **[P2] Usage bar contradicts its legend.** Bar painted used+pending as one primary fill while the legend says pending is amber. Fix: two segments (used primary, pending warning), clamped to 100%; bar marked `aria-hidden` since the `x / y` text beside it states the same value. Category: Implementation Integrity.
- **[P2] Status filter buttons had no state semantics and were ~40px tall.** `admin/leaves`: added `type="button"`, `aria-pressed`, `min-h-11`. Category: Accessibility / Responsive. WCAG 4.1.2, 2.5.8.
- **[P2] Approve / Reject (32px) and apply Submit/Cancel (40px) below the 44px tap target on phones.** Added `max-md:min-h-11`; desktop size unchanged. Category: Responsive.
- **[P2] Success toast not announced.** Added `role="status"`; dismiss button got `type="button"` and a padded hit area. Category: Accessibility.
- **[P3] Login page used `h2` as its only heading.** Now `h1` (same classes, same look). Email and password inputs got `autocomplete="username"` / `current-password` (WCAG 1.3.5). Category: Accessibility.
- **[P3] Leave-types loading placeholder and live day count.** Placeholder has `aria-busy`; "Total Leave Days" block has `role="status"` so the changing count is announced. Category: Accessibility.
- **[P3] Numbers in columns.** `tabular-nums` on entitled/used/pending/remaining and the usage fraction. Category: Typography.

Files edited: `src/routes/(app)/dashboard/+page.svelte`, `src/routes/(app)/admin/leaves/+page.svelte`, `src/routes/(app)/leaves/apply/+page.svelte`, `src/routes/(auth)/login/+page.svelte`.

### Skipped (with reason)

- **[P1] Manager sees raw `employee_id` / `leave_type_id`** (`admin/leaves` ~271, ~292). Needs an employee-name source (API/query not present) and changes card IA; not a visual fix. Leave type could use the dashboard's `useLeaveTypes` map, but that changes displayed content, so deferred. Suggested: `/impeccable clarify` + a data-layer change.
- **[P1] Native `confirm()` / `alert()` for approve** (`admin/leaves` ~54, ~70) and hand-rolled toast. Replacing with kit `ConfirmDialog` and `Toaster` changes interaction behavior; brief says behavior unchanged. Suggested: `/impeccable harden`.
- **[P1/P2] Filter row should be kit `Tabs`** with panel association, arrow keys and counts. Behavior/markup change beyond polish; `aria-pressed` applied as the minimal correct fix.
- **[P2] `(app)` group has no shell, nav, landmark or skip link** (no `(app)/+layout.svelte`); no path from dashboard to "Request leave" or from there to approvals. New UI, not a refinement. Suggested: `/impeccable shape` then `layout` with kit `AppShell` + `SidebarNav`, approval item gated on `isManager`.
- **[P2] Denial duplicated:** `/403` page vs inline "Access Denied" card in `admin/leaves`. Product decision.
- **[P2] Dashboard hierarchy:** Remaining days (the answer the employee wants) sits third; colored badge carries low/ok by color alone. Reorder and a text label would add copy. Suggested: `/impeccable layout` + `clarify`.
- **[P2] Legend copy** says Pending = "Approved requests awaiting start date", contradicting `PENDING` meaning in PRD 5.2 (awaiting approval). Copy is frozen for this pass; needs a content owner. Suggested: `/impeccable clarify`.
- **[P3] Apply form:** CTA disabled until valid (no error feedback on click, unreachable by keyboard); calendar-day count vs working days and balance (PRD requires working days computed by backend); 1.5 s redirect hides confirmation; redundant "Need Help?" block; mixed `*` / "(Optional)" required marking. All behavior or copy changes. Suggested: `/impeccable harden`, `distill`.
- **[P3] `/` redirect** is client-side `onMount` + "Redirecting..." flash; a server `redirect(307)` in `+page.server.ts` would remove it. Behavior change in routing, deferred.
- **[P3] Status enum shown raw** ("PENDING"); sentence case is a copy change. Hard-coded `href="/dashboard"` in `403` and `admin/leaves` vs `resolve()` elsewhere. Leave status badge overrides kit shape with `rounded-full px-3 py-1`; documented in DESIGN.md as the one pill rather than removed.
- **[P3] Login:** "Multi-tenant HR Management System" is jargon, no forgot-password link, tenant field does not say it is a slug, errors not tied to fields via `aria-describedby`. Copy/behavior; FR-HRS-01.8 (password recovery) is not built yet.
- **[P3] Localisation:** all UI copy is English, dates `en-US`; PRD NFR-HRS-12 requires `{id,en}` with a parity test and the personas expect Indonesian. This repo has no i18n layer; out of design-pass scope. Suggested: `/impeccable harden`.
- **Kit-side options (not touched, outside repo):** a 44px Button `size`, a soft `Badge` tone for remaining-days status, and kit `Progress` supporting two segments.

## Patterns and systemic issues

- Hand-rolled equivalents of kit components appear on every page; one adoption pass would remove the drift and bring ARIA for free.
- Semantic status color used as text color on its own tint (three places).
- Missing app shell means page-level IA (landmarks, nav, role-gated links) is not yet designed.

## Positive findings

Token-only styling and working dark mode by construction; complete loading/error/empty/data states; labels and invalid state wired correctly; date parsing avoids the UTC-midnight off-by-one; role guard shown both server-side (hooks) and in the page.

## Recommended actions

1. **[P1] `/impeccable clarify`**: names instead of IDs, Pending legend wording, enum casing (needs data and copy owner).
2. **[P1] `/impeccable harden`**: kit `ConfirmDialog` + `Toaster`, kit `Tabs`, apply-form submit validation, i18n.
3. **[P2] `/impeccable shape` then `/impeccable layout`**: `(app)` shell and navigation, dashboard hierarchy.
4. **[P3] `/impeccable distill`**: remove redundant help and legend blocks.
5. `/impeccable polish` as the final step.
