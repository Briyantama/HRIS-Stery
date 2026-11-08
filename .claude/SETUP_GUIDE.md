# Claude Code Setup Guide: HRIS-Stery

## What is here

- Skills in `.claude/skills/<name>/SKILL.md`: `adr-review`, `architecture-review`, `migration-review`, `event-review`, `go-style-review`.
- Reviewer prompts in `.claude/agents/`: `go-backend-reviewer`, `migration-reviewer`, `event-stream-reviewer`.
- Path-scoped rules in `.claude/rules/`: `system.md` (all files), `golang.md` (`apps/api/**`), `svelte.md` (`apps/web/**`).
- `settings.json` registers the skills and reviewers. No hooks are configured.

## How to call them

Name the skill or reviewer in the request:

```text
Use the migration-review skill for the new leave_policies table.
Review services/leave/http.go with go-backend-reviewer.
```

## Commands the reviewers expect to run

Run Go commands from `apps/api`:

```bash
go vet ./...
go test ./services/leave/... -cover
go test ./internal/db -run . -v          # needs TEST_DATABASE_URL, otherwise skipped
DATABASE_URL=... go run ./cmd/migrate up
```

The full `scripts/check.sh` runs in GitHub CI, not on the dev machine. `go test -race` needs cgo and runs only in CI.

## Where decisions live

- `CLAUDE.md`: rules and invariants
- `docs/adr/`: ADR-0002 (RLS), ADR-0004 (Phase 2 deferral), ADR-0005 (single-module API), ADR-0006 (Dragonfly Streams)
- `docs/FSD_HRIS.en.md`: functional spec
