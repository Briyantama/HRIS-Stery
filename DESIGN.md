---
name: HRIS-Stery
description: Tidy, trustworthy HR workspace on the shared Stery kit: cool neutral surfaces, one indigo-blue action color, status color only for meaning.
colors:
  primary: "oklch(0.5 0.19 262)"
  primary-foreground: "oklch(0.99 0 0)"
  background: "oklch(0.985 0.002 250)"
  foreground: "oklch(0.21 0.02 260)"
  surface: "oklch(1 0 0)"
  muted: "oklch(0.955 0.005 250)"
  muted-foreground: "oklch(0.5 0.02 260)"
  border: "oklch(0.9 0.008 250)"
  input: "oklch(0.86 0.01 250)"
  ring: "oklch(0.55 0.18 260)"
  secondary: "oklch(0.94 0.01 250)"
  accent: "oklch(0.75 0.15 70)"
  accent-foreground: "oklch(0.22 0.03 60)"
  destructive: "oklch(0.55 0.21 27)"
  success: "oklch(0.55 0.14 150)"
  warning: "oklch(0.72 0.16 75)"
  warning-foreground: "oklch(0.22 0.04 60)"
  info: "oklch(0.58 0.14 235)"
typography:
  display:
    fontFamily: "ui-sans-serif, system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif"
    fontSize: "1.875rem"
    fontWeight: 700
    lineHeight: "2.25rem"
    letterSpacing: "-0.025em"
  title:
    fontFamily: "ui-sans-serif, system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif"
    fontSize: "1rem"
    fontWeight: 600
    lineHeight: "1.5rem"
  body:
    fontFamily: "ui-sans-serif, system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: "1.25rem"
  label:
    fontFamily: "ui-sans-serif, system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 500
    lineHeight: "1rem"
    letterSpacing: "0.05em"
rounded:
  sm: "calc(0.5rem - 4px)"
  md: "calc(0.5rem - 2px)"
  lg: "0.5rem"
  xl: "calc(0.5rem + 4px)"
spacing:
  tight: "8px"
  base: "16px"
  section: "24px"
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-foreground}"
    rounded: "{rounded.md}"
    height: "40px"
    padding: "0 16px"
  button-destructive:
    backgroundColor: "{colors.destructive}"
    textColor: "{colors.primary-foreground}"
    rounded: "{rounded.md}"
    height: "40px"
    padding: "0 16px"
  button-outline:
    backgroundColor: "{colors.background}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.md}"
    height: "40px"
    padding: "0 16px"
  input:
    backgroundColor: "{colors.background}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.md}"
    height: "40px"
    padding: "0 12px"
  card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.lg}"
    padding: "24px"
---

# Design System: HRIS-Stery

Source of truth: `stery-kit/web/src/styles/{tokens,theme,base}.css`, consumed through `apps/web/src/app.css` (`@import 'tailwindcss'; @import '@stery/kit/styles/theme.css';`). HRIS adds no tokens of its own. Values above are the kit's light theme; dark values live under `[data-theme='dark']`. Never rename or fork kit variables here; override them after the kit import if HRIS ever needs to diverge.

## Overview

**Creative North Star: "The Tidy Ledger"** (inferred from PRD: "tidy, secure, auditable"; not user-confirmed).

An HR workspace where the page behaves like a well-kept register: cool, quiet neutrals, one confident blue for the action that matters, and color reserved for status (leave approved, pending, rejected, remaining balance). Density is moderate and scannable; people come here to clear a queue or file a request, then leave.

This is a refinement surface: the incumbent look is the kit's, and HRIS stays thin on top of it. Brand lives in precision (consistent spacing, honest status color, clear states), not in ornament.

**Key Characteristics:**
- Cool off-white page (`background`), white raised cards (`surface`), hairline borders.
- Single action color (`primary`); `accent` amber is secondary emphasis, used by the kit for hover fills.
- Status colors carry meaning only: `success`, `warning`, `destructive`, `info`.
- System font stack, no webfont. Sentence-case copy, uppercase only for micro-labels.
- Flat by default; one soft shadow on cards and toasts.
- Light and dark themes via `data-theme`; every color must be a token so both work.

## Colors

Cool blue-gray neutrals (hue 250-260) with a single indigo-blue primary; warm amber accent and three status hues.

### Primary
- **Ledger Indigo** (`oklch(0.5 0.19 262)`, dark `oklch(0.66 0.16 262)`): primary buttons, active tab underline, progress fill, links.

### Secondary / Tertiary
- **Amber Accent** (`oklch(0.75 0.15 70)`): the kit uses it as the hover fill for outline and ghost buttons; HRIS does not use it as a standalone brand color.
- **Success Green** (`oklch(0.55 0.14 150)`), **Warning Amber** (`oklch(0.72 0.16 75)`), **Destructive Red** (`oklch(0.55 0.21 27)`), **Info Blue** (`oklch(0.58 0.14 235)`): leave status, alerts, validation.

### Neutral
- **Page Paper** (`background`, `oklch(0.985 0.002 250)`): page and input fill.
- **Card White** (`surface`, `oklch(1 0 0)`): cards, dialogs, toasts.
- **Mist** (`muted`, `oklch(0.955 0.005 250)`): skeletons, help panels, code chips, progress track.
- **Ink** (`foreground`, `oklch(0.21 0.02 260)`) and **Slate** (`muted-foreground`, `oklch(0.5 0.02 260)`): text. Slate on Mist is the weakest pairing in use; keep it for secondary text only.
- **Hairline** (`border`, `oklch(0.9 0.008 250)`), **Field Edge** (`input`, `oklch(0.86 0.01 250)`), **Focus Ring** (`ring`).

### Named Rules
**The Meaning-Only Color Rule.** `success`, `warning`, `destructive`, `info` appear only to state status. Never use them as decoration or to differentiate unrelated things.
**The Token-Only Rule.** No raw hex, rgb or oklch in routes. Use `bg-primary`, `text-muted-foreground`, `border-border` and their `/opacity` forms.

## Typography

**Display / Body / Label Font:** system UI stack (`ui-sans-serif, system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif`); mono `ui-monospace, SFMono-Regular, Menlo, Consolas`.

**Character:** neutral and native. Hierarchy comes from weight and size, not from font personality.

### Hierarchy
- **Display** (700, `text-3xl` 1.875rem, tight tracking): one `<h1>` per page.
- **Title** (600, `text-base`, card titles; `font-semibold` list-item headings): card and section headings.
- **Body** (400, `text-sm`): default control and content size; page intro lines use the inherited 1rem in `text-muted-foreground`.
- **Label** (500, `text-xs`, `uppercase`): micro-labels over values (Entitled, Dates, Leave Type). Add `tracking-wide`; keep to short single words.

### Named Rules
**The One H1 Rule.** Every page has exactly one `h1` and headings step down without skipping.
**The Tabular Numbers Rule.** Day counts, dates and balances that sit in columns or compare across cards use `tabular-nums`.

## Layout

Page content sits in a `space-y-6` vertical stack. Forms are a centered `max-w-2xl` column; the approval list is full width with cards stacked at `space-y-4`; the balance dashboard is a responsive grid (`md:grid-cols-2 lg:grid-cols-4`, `gap-4`). Spacing rhythm: 8 / 16 / 24 px (`gap-2`, `gap-4`, `space-y-6`). Mobile-first: columns collapse to one at default width, and action rows stack. The `(app)` group currently has no shared shell (no `+layout.svelte`); navigation between dashboard, request and approval pages is not yet part of the design system.

## Elevation & Depth

Tonal layering with hairline borders. Surfaces are flat at rest; cards get `shadow-sm` (`--elevation-1`: `0 1px 2px hsl(220 3% 15% / 0.08)`). Toasts use `shadow-lg` (`--elevation-3`). Dialogs and overlays use `--overlay` scrim and `--z-modal` (70). In dark mode the shadow color deepens and surfaces step lighter.

### Shadow Vocabulary
- **Card rest** (`shadow-sm`): default card.
- **Raised** (`shadow-md`, `--elevation-2`): menus and popovers.
- **Floating** (`shadow-lg`, `--elevation-3`): toasts, dialogs.

### Named Rules
**The Flat-By-Default Rule.** Depth is a response (toast, dialog, menu), not decoration on static content.

## Shapes

One radius family from `--radius: 0.5rem`: controls `rounded-md` (6px), cards `rounded-lg` (8px). The leave status badge is the one pill (`rounded-full px-3 py-1` override on the kit Badge); keep it as is, do not spread the pill to other components. Borders are 1px `border-border`; inputs use `border-input`. Invalid fields switch the border and ring to `destructive`.

## Components

All components come from `@stery/kit/ui`; routes compose, never restyle the kit.

### Buttons
- **Shape:** `rounded-md`, heights `sm` 32px / `md` 40px / `lg` 48px, `icon` 40px square.
- **Primary:** `bg-primary text-primary-foreground`, hover `primary/90`. One per view region.
- **Destructive:** solid red for Reject. **Outline / Ghost:** hover fills with `accent`. **Link:** `text-primary` underline on hover.
- **Focus:** 2px ring + 2px offset in `ring`. Disabled: 50% opacity, no pointer events. `loading` prop for async actions.

### Cards / Containers
- `rounded-lg border border-border bg-surface shadow-sm`; variants `outline`, `muted`. Internal padding `p-6` via `CardHeader`/`CardContent`. Help and legend panels use `rounded-lg border border-border bg-muted p-4`.

### Inputs / Fields
- `h-10 rounded-md border-input bg-background px-3 text-sm`; label above via `Label`; `Select`, `Textarea`, native `type="date"` share the look. Invalid: destructive border and ring, message in `text-sm text-destructive` below. Required fields marked with `*` in the label.

### Badges / Status
- Kit `Badge` with `tone`: `warning` Pending, `success` Approved, `destructive` Rejected. Solid fills; text is the status word (never color alone). Status text on tinted surfaces uses `text-foreground`, never `text-warning` or `text-success` (fails 4.5:1 on light tints); color goes on the surface and border.

### Alerts and Toasts
- Kit `Alert` with `tone` and optional `title`/`description`; success toast is a fixed `shadow-lg` card with an icon and dismiss button.

### Tabs (status filter)
- Text buttons on a `border-b border-border` rail; the active tab carries `border-b-2 border-primary text-primary`.

### Dialog
- Kit `Dialog` with header (title + description), body, right-aligned footer (Cancel outline, destructive confirm).

## Do's and Don'ts

### Do:
- **Do** use Tailwind utilities backed by kit tokens (`bg-surface`, `text-muted-foreground`, `rounded-lg`) so light and dark both work.
- **Do** give every async region loading (kit `Skeleton`), empty, error (`Alert`) and success states.
- **Do** keep touch targets at or above the kit `--tap-target` (2.75rem) for primary mobile actions.
- **Do** state status in words as well as color.
- **Do** keep copy plain and task-first; reflect real API values, not placeholders.

### Don't:
- **Don't** add raw colors, custom fonts, gradients, or glassmorphism.
- **Don't** nest cards inside cards or use side-stripe accent borders.
- **Don't** rename or redefine kit CSS variables in HRIS; override after the kit import if ever needed.
- **Don't** use `confirm()` / `alert()` for flows the kit `ConfirmDialog` / `Alert` can express.
- **Don't** claim payroll, tax or other deferred modules in any UI.
