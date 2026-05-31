# Auth-Service Implementation Status

**Sprint 1 — Weeks 3-4 — M1: Auth End-to-End**

## ✅ Completed

### Domain Layer (100%)

- [x] Value objects: `TenantID`, `UserID`, `TenantSlug`, `RoleID`, `Permission`
- [x] Aggregates: `Tenant`, `User`, `Role` with encapsulation
- [x] Domain events: `UserRegisteredEvent`, `UserLoggedInEvent`, `UserLoginFailedEvent`, `PasswordChangedEvent`, `TenantCreatedEvent`
- [x] Repository port interfaces: `TenantRepository`, `UserRepository`, `RoleRepository`, `PermissionRepository`
- [x] Password hashing (bcrypt) + validation in User aggregate
- [x] Email validation

### Application Layer (100%)

- [x] Commands:
  - `LoginCommand` + `LoginHandler` — validates tenant, user, password; generates tokens; publishes events
  - `RefreshTokenCommand` + `RefreshTokenHandler` — exchanges refresh for new access token
  - `RevokeTokenCommand` + `RevokeTokenHandler` — logout (revoke token)
  - `RegisterTenantCommand` + `RegisterTenantHandler` — seed tenant, roles, permissions, create admin user
- [x] Queries:
  - `ValidateTokenQuery` + `ValidateTokenHandler` — parse and verify access token
  - `GetPermissionsQuery` + `GetPermissionsHandler` — fetch all permissions for a user
- [x] Port interfaces: `TokenService`, `EventPublisher`

### Infrastructure Layer (60%)

- [x] JWT Signer: RSA signing/verification, CustomClaims, RefreshClaims
- [x] TokenService Implementation: access/refresh token pair generation, validation, revocation via Redis
- [x] Redis-backed token revocation list
- [ ] Postgres Repositories (TenantRepository, UserRepository, RoleRepository, PermissionRepository)
  - Stub structure created, implementations pending
  - Must use `services/_shared/postgres/rls.go` for RLS context
  - Must use sqlc for type-safe queries
- [ ] NATS Event Publisher implementation
- [ ] gRPC handlers (interfaces layer)

### Migrations (100%)

- [x] `auth-service/migrations/001_create_auth_schema.up.sql` — tenants, users, roles, permissions, refresh_tokens, login_attempts tables with RLS
- [x] `auth-service/migrations/001_create_auth_schema.down.sql` — rollback

### Module Setup (100%)

- [x] `go.mod` with all required dependencies

## ⏳ Remaining (40% of sprint)

### Infrastructure Layer

1. **Postgres Repositories** (~6-8 hours)
   - TenantRepository.GetBySlug, GetByID, Create, Update, Delete
   - UserRepository.GetByTenantAndEmail, GetByID, Create, Update, SetRoles, GetRoles
   - RoleRepository.GetByTenantAndName, ListByTenant, Create
   - PermissionRepository.GetForRole, AssignToRole, RemoveFromRole
   - All queries use sqlc + pgx with RLS context from `services/_shared/postgres/rls.go`

2. **NATS Event Publisher** (~2 hours)
   - Implement EventPublisher interface
   - Serialize domain events to event envelope format (ADR-0003)
   - Publish to appropriate subjects (hris.identity.user.* etc.)
   - Async and sync variants

### Interfaces Layer (Handlers)

3. **gRPC Service Handlers** (~4 hours)
   - `services/auth-service/internal/interfaces/grpc/auth_service.go`
   - Implement gRPC server for AuthService
   - Map proto messages ↔ domain/application types
   - Error handling → gRPC status codes
   - Tenant context injection from JWT or header

4. **gRPC-Gateway HTTP Handlers** (Auto-generated)
   - `buf generate` produces HTTP/JSON wrappers
   - Ensure grpc-gateway annotations are in proto (already done)

### Testing

5. **Unit Tests** (~4 hours)
   - Domain layer: User password hashing, Tenant slug validation, Role/Permission logic
   - Application layer: Login command success/failure paths, token refresh, permissions aggregation

6. **Integration Tests** (~4 hours)
   - Docker compose + real Postgres + Redis + NATS
   - Full login flow: RegisterTenant → CreateUser → Login → RefreshToken → Revoke
   - RLS enforcement: user from tenant A cannot see tenant B's data
   - Event publishing: verify events land on NATS with correct payload

## Next Steps

### Immediate (Next 2-4 hours)

1. Create sqlc-generated query files and repository implementations
2. Implement NATS EventPublisher
3. Write gRPC handlers to wire up commands/queries

### Then (Following 4-6 hours)

4. Write unit tests for domain and application layers
5. Write integration tests using Docker Compose
6. Generate proto stubs with `buf generate` and verify gateway works

### Final (Last 2 hours)

7. End-to-end manual test: SvelteKit login → Laravel gateway → auth-service gRPC
8. Verify tokens in Redis, events in NATS, audit log written
9. M1 milestone: auth working end-to-end (should be Week 4, Day 3-4)

## Caveats & Decisions

- **RSA Key Generation**: Assumes keys are generated externally and provided via environment
  - In production, use AWS KMS or similar
  - In dev, generate with: `openssl genrsa -out private.pem 2048 && openssl rsa -in private.pem -pubout -out public.pem`
- **Token Revocation**: Refresh tokens stored in Redis with TTL; no database persistence (acceptable for MVP)
  - Option: upgrade to persistent list in Postgres later if audit is required
- **Password Reset**: Not implemented in Sprint 1 (add in Phase 2)
- **Email Verification**: User marked verified on registration by admin; no email confirmation flow (Phase 2)
- **Rate Limiting**: Not implemented yet (move to gateway level in hardening sprint)
