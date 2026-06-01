# Makefile reference

Run all commands from the **repository root**. Requires GNU Make and Bash (Git Bash on Windows).

## Quick validation

```bash
./scripts/validate-makefile.sh
```

```powershell
.\scripts\validate-makefile.ps1
```

Logs are written to `.make-validate/`.

## Auto-discovered services

| Variable           | Source                           |
| ------------------ | -------------------------------- |
| `GO_SERVICES`      | `services/*-service/go.mod`      |
| `MIGRATE_SERVICES` | `services/*-service/migrations/` |

Phase 2 services (`audit-service`, `payroll-service`, `ai-service`) are **not** in the tree and are no longer hard-coded.

## Prerequisites

| Tool                                                        | Used by                                         |
| ----------------------------------------------------------- | ----------------------------------------------- |
| Go 1.24+                                                    | `tidy`, `vet`, `test-go`, `build-go`            |
| [buf](https://buf.build/docs/installation)                  | `proto-*`                                       |
| [golang-migrate](https://github.com/golang-migrate/migrate) | `migrate-*`                                     |
| Docker Compose                                              | `dev`, `setup`                                  |
| golangci-lint                                               | `lint-go` (optional; skips if missing)          |
| Node / npm                                                  | `apps/web` targets (skipped if missing)         |
| PHP / Composer                                              | `apps/api-gateway` targets (skipped if missing) |

## Environment

| Variable | Default                                                                          |
| -------- | -------------------------------------------------------------------------------- |
| `DB_URL` | `postgres://hris_admin:hris_admin_secret@localhost:5432/hris_db?sslmode=disable` |

Migrations use **direct Postgres (5432)** with the `hris_admin` role, not PgBouncer.

## Manual / infra targets

These require Docker and are not run by the validation script:

- `make dev` / `dev-down` / `dev-reset` / `dev-logs`
- `make migrate-up` / `migrate-down` (after `make dev`)
- `make setup` (= `dev` + `migrate-up`)
- `make test-integration` (Postgres + Redis + `INTEGRATION=true`)
