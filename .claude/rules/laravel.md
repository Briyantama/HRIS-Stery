---
description: Laravel API Gateway rules for HRIS-Stery.
globs: ["apps/api-gateway/**/*.php"]
---

# Laravel Rules

## Role: API Gateway + Auth Only

Laravel validates, authenticates, and proxies. It does not own domain logic.

## Mandatory Patterns

- Sanctum API token mode. JWT payload: sub, tid (tenant_id), roles[], exp, iat.
- TenantResolver middleware sets tenant_id on the authenticated user from JWT.
- Controllers: ≤ 20 lines of logic. Validation → Service → Response.
- Form Requests for every POST/PUT. No `$request->input()` in controllers directly.
- Service classes in `app/Services/` call Go services via Guzzle (HTTP/JSON grpc-gateway).
- Spatie Permission for RBAC: hr_admin | manager | employee.
- Return JSON errors: { code, message }. Never expose stack traces.
- Laravel Pint (PSR-12) before every commit.

## Forbidden

- gRPC-PHP extension — use HTTP/JSON grpc-gateway endpoints instead
- Business or domain logic in controllers or middleware
- Raw `DB::` calls for domain data — use gRPC service classes
- Fat controllers (> 20 lines)
- Phase 2 routes that return anything other than 501
