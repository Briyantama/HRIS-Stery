---
description: SvelteKit 5 frontend rules for HRIS-Stery.
globs: ["apps/web/**/*.svelte", "apps/web/**/*.ts"]
---

# SvelteKit Rules

## Stack

SvelteKit 2 + Svelte 5 (runes) + Vite 8 + TypeScript 6 + Tailwind 4 + TanStack Query + Zod + lucide. Node >= 20.19, npm.

## Mandatory patterns

- TanStack Query (`createQuery`, `createMutation`) for all server state. No raw `fetch` in `+page.svelte`.
- `src/lib/api/client.ts` is the only API caller. It unwraps the `{success, data, error}` envelope and throws on `success: false`.
- Zod schemas in `src/lib/schemas/` for forms and API responses. Response fields are snake_case, as the Go API sends them.
- Auth guard in `hooks.server.ts`. Route groups: `(auth)` public, `(app)` protected.
- Svelte stores hold UI-only state (sidebar, modals). Server data stays in the TanStack Query cache.
- Unit tests: Vitest. E2E: Playwright.

## Forbidden

- Raw `fetch()` in a `+page.svelte` `<script>`; use `createQuery`
- Business calculations in Svelte components
- Importing server-only code in client files
- Phase 2 route files
