---
description: SvelteKit 5 frontend rules for HRIS-Stery.
globs: ["apps/web/**/*.svelte", "apps/web/**/*.ts"]
---

# SvelteKit Rules

## Stack

SvelteKit 2.x + Svelte 5 (runes mode) + TypeScript + TailwindCSS + shadcn-svelte + TanStack Query

## Mandatory Patterns

- TanStack Query (`createQuery`, `createMutation`) for all server state. No raw fetch in `+page.svelte`.
- Zod schemas in `src/lib/schemas/` for all form validation and API response shapes.
- Auth guard in `hooks.server.ts`. Route groups: `(auth)` public, `(app)` protected.
- shadcn-svelte for all UI components. No raw `<input>` elements without a wrapper.
- Svelte stores for UI-only state (sidebar, modals). Never store server data in stores.
- OpenTelemetry in `instrumentation.server.js`.
- E2E: Playwright in `tests/e2e/`. Unit: Vitest in `tests/unit/`.

## Forbidden

- Raw `fetch()` in `+page.svelte` `<script>` — use `createQuery`
- Business calculations in Svelte components
- Importing server-only code in client files (use `$app/environment` browser check)
- Phase 2 route files
