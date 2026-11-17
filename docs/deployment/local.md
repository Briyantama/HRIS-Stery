# Local development

One Go process (`apps/api`) serves `/api/v1` and runs the event consumers. Postgres 18 holds the data,
Dragonfly holds the token revocation list, rate limits and the `hris:events` stream, and RustFS stores documents.

## Prerequisites

- Go (version in `apps/api/go.mod`) and Node >= 20.19
- Docker (Dragonfly and RustFS come from `deploy/docker-compose.yml`)
- PostgreSQL 18 on the host
- The kit repo checked out next to this one: `../stery-kit` (the Go module and `@stery/kit` resolve from it)

## Start

```bash
cp .env.example .env          # fill the secrets; nothing has a default
make up                       # Dragonfly + RustFS, migrations, API on :8000 (log in .run/api.log)
make status                   # pid and /healthz
cd apps/web && npm install && API_URL=http://localhost:8000 npm run dev
```

`make down` stops the API and the containers. `make build` writes the binary to `.run/api`.

## Migrations

```bash
cd apps/api
go run ./cmd/migrate up       # also: down, status
```

Migrations are goose files embedded from `apps/api/internal/db/migrations`.

## Tests

```bash
cd apps/api
TEST_DATABASE_URL=postgres://postgres:<password>@127.0.0.1:5432/postgres?sslmode=disable go test ./...
go test ./services/leave/...  # one package
cd ../web && npm run check && npm run test:unit
```

Tests that need Postgres create and drop their own databases through `TEST_DATABASE_URL`.
`make check` (`scripts/check.sh`) is what CI runs.

## Troubleshooting

- `go` complains about `go.work`: run with `GOWORK=off`.
- 401 on every call after login: the API's `AUTH_PUBLIC_KEY` does not match the key that signed the token.
- Web login returns "Cannot reach the HRIS API": `API_URL` is wrong or the API is down (`make status`).
