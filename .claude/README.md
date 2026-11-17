# Claude Code Project Configuration: HRIS-Stery

Updated 2026-11-08 for the single-module Go API (ADR-0005) and Dragonfly Streams (ADR-0006). `CLAUDE.md` at the repo root holds the rules; this folder holds review checklists and reviewer prompts.

## Skills (checklists, read before you implement)

| Skill | Use it before |
|---|---|
| `adr-review` | changing a service package or a cross-cutting decision |
| `architecture-review` | writing an `http.go` handler or a repository |
| `migration-review` | adding a goose file under `apps/api/internal/db/migrations` |
| `event-review` | publishing to or consuming from `hris:events` |
| `go-style-review` | reviewing any Go change |

## Subagents (reviewers, use after you implement)

| Agent | Reviews |
|---|---|
| `go-backend-reviewer` | Go layering, tenant context, error mapping, tests |
| `migration-reviewer` | goose Up/Down reversibility, RLS, grants |
| `event-stream-reviewer` | envelope, publish-after-commit, consumer idempotency |

No hooks are configured (`settings.json` has `"hooks": {}`); the rules are enforced by review.

## Rules

- `rules/system.md`: architecture, ADR index, event envelope
- `rules/golang.md`: `apps/api` layout and Go patterns
- `rules/svelte.md`: `apps/web` patterns

## Typical flows

| Task | Skills | Reviewer |
|---|---|---|
| New endpoint | adr-review, architecture-review | go-backend-reviewer |
| New table or column | migration-review | migration-reviewer |
| New event or consumer | event-review | event-stream-reviewer |

See `REVIEW_DECISION_TREE.md` for the order of steps and `SETUP_GUIDE.md` for how to call a skill or reviewer.
