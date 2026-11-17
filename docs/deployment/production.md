# Production deployment

`docker-compose.production.yml` at the repo root describes the whole stack: Postgres 18, Dragonfly, RustFS
(plus a one-shot bucket init) and the API image. Every secret comes from `.env`; compose fails fast (`:?`)
when one is missing.

## Build

The API image needs the kit sources, so the build context is the parent directory that holds both
`HRIS-Stery/` and `stery-kit/`:

```bash
docker build -f HRIS-Stery/apps/api/Dockerfile -t hris-api .
```

The image carries two binaries: `api` (entrypoint) and `migrate`.

## Release

```bash
docker compose -f docker-compose.production.yml --env-file .env up -d postgres dragonfly rustfs rustfs-init
docker compose -f docker-compose.production.yml --env-file .env run --rm --entrypoint migrate api up
docker compose -f docker-compose.production.yml --env-file .env up -d api
curl -fsS http://127.0.0.1:8000/healthz
```

Run migrations once per release, before the new API starts. Roll back the schema with
`--entrypoint migrate api down` (one step) only when the release's migration has a tested down section.

## Required configuration

| Variable | Notes |
| --- | --- |
| `POSTGRES_PASSWORD`, `DATABASE_URL` | the API connects as the app role, not the superuser |
| `REDIS_PASSWORD` | Dragonfly `--requirepass` and the API client |
| `AUTH_PRIVATE_KEY_BASE64`, `AUTH_PUBLIC_KEY_BASE64` | RS256 key pair (or the PEM variants `AUTH_PRIVATE_KEY`/`AUTH_PUBLIC_KEY`) |
| `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_BUCKET`, `S3_PUBLIC_ENDPOINT` | RustFS credentials and the URL clients use for presigned links |
| `SMTP_*`, `EMAIL_TEMPLATES_DIR` | notification email |
| `TRUSTED_PROXY_CIDRS` | the reverse proxy range, so rate limits see the client IP |

## Operations

- Health: `GET /healthz` is a liveness probe (process up); it does not check Postgres, Dragonfly or RustFS.
- Logs: JSON lines from slog on stdout; level from `LOG_LEVEL`.
- Events: `XINFO GROUPS hris:events` in Dragonfly shows the four consumer groups (employee, leave, audit,
  notification) and their lag.
- Backups: Postgres is the only state that matters for recovery; take daily dumps plus WAL archiving.
  RustFS data (`rustfs_data`) holds documents and needs its own backup.
