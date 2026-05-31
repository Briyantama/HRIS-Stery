#!/usr/bin/env bash
# End-to-end auth smoke test via grpcurl.
# Prerequisites: make dev, make migrate-up, auth-service running on :50051
set -euo pipefail

HOST="${GRPC_HOST:-localhost:50051}"
SLUG="e2e-$(date +%s)"
EMAIL="admin@${SLUG}.test"
PASSWORD="ValidPass123!"

echo "→ RegisterTenant"
grpcurl -plaintext -d "{
  \"company_name\": \"E2E Corp\",
  \"tenant_slug\": \"${SLUG}\",
  \"admin_email\": \"${EMAIL}\",
  \"admin_password\": \"${PASSWORD}\",
  \"admin_name\": \"E2E Admin\"
}" "${HOST}" hris.auth.v1.AuthService/RegisterTenant

echo "→ Login"
LOGIN=$(grpcurl -plaintext -d "{
  \"email\": \"${EMAIL}\",
  \"password\": \"${PASSWORD}\",
  \"tenant_slug\": \"${SLUG}\"
}" "${HOST}" hris.auth.v1.AuthService/Login)

ACCESS=$(echo "${LOGIN}" | jq -r '.accessToken')
REFRESH=$(echo "${LOGIN}" | jq -r '.refreshToken')

echo "→ ValidateToken"
grpcurl -plaintext -d "{\"access_token\": \"${ACCESS}\"}" \
  "${HOST}" hris.auth.v1.AuthService/ValidateToken

echo "→ RefreshToken"
REFRESHED=$(grpcurl -plaintext -d "{\"refresh_token\": \"${REFRESH}\"}" \
  "${HOST}" hris.auth.v1.AuthService/RefreshToken)
NEW_REFRESH=$(echo "${REFRESHED}" | jq -r '.refreshToken')

echo "→ RevokeToken"
grpcurl -plaintext -d "{\"refresh_token\": \"${NEW_REFRESH}\"}" \
  "${HOST}" hris.auth.v1.AuthService/RevokeToken

echo "✓ E2E auth flow complete"
