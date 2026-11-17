#!/bin/sh
# Entrypoint init hook: feeds env values to init-postgres.sql as psql variables (no secret in git).
set -eu
exec psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  -v app_password="${HRIS_APP_PASSWORD:?HRIS_APP_PASSWORD is required}" \
  -v POSTGRES_USER="$POSTGRES_USER" -v POSTGRES_DB="$POSTGRES_DB" \
  -f /opt/init-postgres.sql
