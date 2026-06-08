# PHASE 5: Deep Code-Level Architecture Review

**Date:** 2026-06-08  
**Status:** Code-level audit complete (zero modifications made)  
**Overall Assessment:** Architecture is sound but has critical contract and security gaps that will cause integration issues

---

## PART 1: SERVICE CONTRACT AUDIT

### 1.1 Proto Contract Inventory

**8 services with proto definitions:**

| Service | Implemented RPCs | Status | Ready |
|---------|-----------------|--------|-------|
| **auth-service** | 6/6 | ✅ Fully implemented | ✅ YES |
| **employee-service** | 8/8 | ✅ Fully implemented | ✅ YES |
| **attendance-service** | 5/5 | ✅ Fully implemented | ✅ YES |
| **leave-service** | 8/8 | ❌ ALL return Unimplemented | ❌ NO |
| **notification-service** | 5/5 | ✅ Fully implemented | ✅ YES |
| **audit-service** | 3/3 | ✅ Fully implemented | ✅ YES |
| **document-service** | 7/7 | ✅ Fully implemented | ✅ YES |
| **ai-service** | 3/3 | ⚠️ Return 501 (Phase 2) | ⚠️ INTENTIONAL |

---

### 1.2 Critical Contract Issues Found

#### 🔴 ISSUE #1: Request Objects Include `tenant_id` — SECURITY ANTI-PATTERN

**Location:** All services

**Problem:** Proto requests explicitly include `tenant_id` as a parameter:

```protobuf
// LEAVE SERVICE EXAMPLE
message ApplyLeaveRequest {
  string tenant_id = 1;      // ← PROBLEM: Frontend can send ANY tenant_id
  string employee_id = 2;
  string leave_type_id = 3;
  // ...
}

message ApproveLeaveRequest {
  string id = 1;
  string tenant_id = 2;      // ← PROBLEM: Frontend can override tenant_id
  string approver_id = 3;
}
```

**Why This Is Wrong:**
- Frontend could theoretically send a different `tenant_id` than in their JWT
- Handlers MUST validate that request tenant_id matches JWT tenant_id
- Best practice: Tenants should NEVER be in request bodies if coming from authenticated users

**Current Defense Level:**
- ✅ Database RLS enforces isolation (will reject invalid tenant_id queries)
- ❌ Application layer has NO validation that request tenant_id matches JWT tenant_id
- ❌ Leave service handlers unimplemented — unclear if they'll validate

**Impact on Gateway Integration:**
- Gateway must extract tenant_id from JWT and FORCE it into all requests (override request body)
- Gateway must validate request tenant_id matches JWT tenant_id and reject mismatches
- This is NON-NEGOTIABLE security

**Examples of vulnerable flow:**
```
Frontend: POST /api/leave/apply
  {tenant_id: "attacker-tenant", employee_id: "victim", ...}
  
Without validation, RLS prevents DB access, but encourages bad pattern.
With validation, gateway catches it.
```

---

#### 🟡 ISSUE #2: Proto Denormalization Not Implemented

**Location:** Employee proto + handler mismatch

**Problem:** Proto defines denormalized read fields that handlers don't populate:

```protobuf
message Employee {
  // ...
  string department_id   = 11;
  string position_id     = 12;
  
  // These are defined but never populated by handler
  string department_name = 18;  // ← Not returned by handler
  string position_name   = 19;  // ← Not returned by handler
  string manager_name    = 20;  // ← Not returned by handler
}
```

**Code Evidence:**
```go
// services/employee-service/internal/interfaces/grpc/employee_service.go:141-148
employees[i] = &employeev1.Employee{
  Id:           emp.ID,
  TenantId:     req.TenantId,
  Email:        emp.Email,
  FullName:     emp.FullName,
  DepartmentId: emp.DepartmentID,
  PositionId:   emp.PositionID,
  // ← department_name, position_name, manager_name NOT SET
}
```

**Impact:**
- Frontend receives IDs but no friendly names (poor UX)
- Frontend must make N additional requests to get names (N+1 problem)
- Alternative: Handler should lookup and populate these fields

**For Gateway Integration:**
- Gateway must decide: enrich in handler, or accept sparse responses?
- If sparse, frontend will make separate requests

---

#### 🟡 ISSUE #3: Pagination Token Not Implemented

**Location:** Employee service ListEmployees

**Code:**
```go
return &employeev1.ListEmployeesResponse{
  Employees:     employees,
  TotalCount:    int32(result.Total),
  NextPageToken: "", // TODO: Implement pagination tokens  ← NOT IMPLEMENTED
}, nil
```

**Impact:**
- ListEmployees returns empty `next_page_token`
- Clients can't properly paginate large result sets
- Only limit/offset works, which is inefficient

**For Gateway Integration:**
- Gateway must document that pagination is incomplete
- Frontend must use limit/offset, not tokens

---

### 1.3 RPC Contract Validation

✅ **All HTTP annotations present:**
- Every RPC has `google.api.http` annotation
- POST/GET methods correctly mapped
- grpc-gateway can auto-generate routes

**Example:**
```protobuf
rpc ApplyLeave(ApplyLeaveRequest) returns (ApplyLeaveResponse) {
  option (google.api.http) = {
    post: "/v1/leave-requests"
    body: "*"
  };
}
```

✅ **Request/response types well-formed:**
- All messages are proto3 compatible
- Proper use of google.protobuf.Timestamp
- No `interface{}` or raw strings

---

## PART 2: HANDLER IMPLEMENTATION AUDIT

### 2.1 Auth Service Handlers ✅ PRODUCTION-READY

**Status:** All 6 RPCs implemented

**Code Quality Check:**
- ✅ Validates all required fields (line 48-50)
- ✅ Maps errors to gRPC status codes (line 59: `codes.PermissionDenied`)
- ✅ Extracts claims properly (line 102-116)
- ✅ Returns typed responses

**Note:** Handlers are NOT checking if request tenant_id matches JWT tenant_id (architectural choice, relies on RLS)

---

### 2.2 Leave Service Handlers ❌ BLOCKER

**Status:** 8/8 RPCs return `codes.Unimplemented`

```go
func (s *LeaveServiceServer) ApplyLeave(...) (*pb.ApplyLeaveResponse, error) {
    return nil, status.Errorf(codes.Unimplemented, "ApplyLeave not yet implemented")
}
```

**Why This Blocks Integration:**
- Frontend cannot create leave requests
- Manager approval workflow is non-functional
- Entire leave module is inoperable

**What's Missing:**
- Dependency injection of command/query handlers
- Handler logic to call leave domain commands
- No error mapping to status codes
- No context extraction for request_id, actor_id

**Effort to Fix:**  
- ~1-2 days (foundation already complete per IMPLEMENTATION_STATUS.md)

---

### 2.3 Employee Service Handlers ✅ FUNCTIONAL

**Status:** All CRUD operations implemented

**Code Quality:**
- ✅ Proper input validation
- ✅ Error codes mapped correctly
- ✅ Creates correct response types

**Gap:** Denormalized fields not populated (see Issue #2 above)

---

## PART 3: GATEWAY READINESS AUDIT

### 3.1 Current State: 0% Implementation

**Directory structure:**
```
apps/api-gateway/app/
├─ Http/          # Empty
├─ Services/      # Empty
└─ Exceptions/    # Not checked
```

**No PHP files found in gateway**

### 3.2 What's Missing for Integration

**Critical infrastructure needed:**

#### 1. gRPC Client Layer
- Guzzle HTTP client wrapper for service calls
- Base class: `BaseGrpcGatewayClient` (needs implementation)
- Service classes: `AuthService`, `EmployeeService`, `LeaveService`, `AttendanceService`, etc.
- Error translation: gRPC JSON error codes → HTTP status codes

**Pattern Required:**
```php
// Example pattern (NOT in codebase)
class BaseGrpcGatewayClient {
    public function callService(string $service, string $endpoint, array $data): array {
        // Call HTTP grpc-gateway endpoint
        // Translate gRPC errors to HTTP
        // Return JSON response
    }
}

class EmployeeService {
    public function listEmployees(string $tenantId, array $filters) {
        return $this->client->callService(
            'employee-service',
            '/v1/employees',
            $filters
        );
    }
}
```

#### 2. Authentication Middleware
- **JWT Validation:** Validate RSA signature using public key from `AUTH_PUBLIC_KEY` env
- **Tenant Extraction:** Extract `tid` from JWT claims
- **Tenant Enforcement:** OVERRIDE request body `tenant_id` with JWT `tid`

**Critical Security Implementation:**
```php
// Pattern (NOT in codebase)
public function validateAndExtractTenant(Request $request): string {
    $jwt = $request->bearerToken();
    $claims = JWT::decode($jwt, $publicKey, ['RS256']);
    
    // CRITICAL: Override request tenant_id if present
    // Never trust frontend-supplied tenant_id
    $tenantId = $claims->tid;
    
    if (isset($request->input('tenant_id')) && 
        $request->input('tenant_id') !== $tenantId) {
        // Log security event
        abort(403, 'Tenant mismatch');
    }
    
    return $tenantId;
}
```

#### 3. Request/Response Controllers
- Empty controllers in `Http/Controllers/Api/`
- Must follow: Validate → ServiceCall → Response

#### 4. Error Translation
```php
// Pattern needed (NOT in codebase)
private function translateGrpcError(array $grpcError): Response {
    $code = $grpcError['code'] ?? 'UNKNOWN';
    return match($code) {
        'NOT_FOUND' => response()->json([...], 404),
        'INVALID_ARGUMENT' => response()->json([...], 400),
        'PERMISSION_DENIED' => response()->json([...], 403),
        'INTERNAL' => response()->json([...], 500),
        default => response()->json([...], 500),
    };
}
```

### 3.3 gRPC-Gateway Endpoint URLs

**Pattern:** Services expose gRPC-gateway on standard HTTP ports

**Current Services (from code audit):**
- auth-service: 50051 (port inference)
- employee-service: 50052
- attendance-service: 50053
- leave-service: 50054
- notification-service: 50055
- audit-service: 50056
- document-service: 50057

**Gateway Will Need to Call:**
```
http://auth-service:50051/v1/auth/login
http://employee-service:50052/v1/employees
http://leave-service:50054/v1/leave-requests
etc.
```

---

## PART 4: FRONTEND READINESS AUDIT

### 4.1 Current State: 0% Implementation

**Directory structure:**
```
apps/web/src/
├─ routes/        # Empty
├─ lib/
│  ├─ schemas/    # Empty
│  ├─ stores/     # Empty
│  └─ components/ # Empty
└─ tests/         # Empty
```

**No .svelte or .ts files found**

### 4.2 What's Missing for MVP

#### 1. Auth Flow
- Login page (email, password, tenant_slug → token)
- Token storage (httpOnly cookie or localStorage)
- Auto-redirect to dashboard
- Token refresh logic

#### 2. Layout & Guards
- `(app)` route group with auth guard
- `(auth)` public route group
- Navigation sidebar

#### 3. Pages
- Dashboard (summary)
- Employees list + detail
- Leave requests (list + apply form)
- Leave approvals (manager view)
- Attendance (check-in/out)

#### 4. Data Layer
- TanStack Query hooks
- Zod schemas for API types
- API client wrapper

---

## PART 5: SECURITY AUDIT

### 5.1 JWT Token Structure ✅ CORRECT

**Token payload (from auth service):**
```json
{
  "sub": "user-uuid",
  "user_id": "user-uuid",
  "tenant_id": "tenant-uuid",
  "tid": "tenant-uuid",
  "email": "user@example.com",
  "roles": ["hr_admin"],
  "exp": 1717000000,
  "iat": 1716996400
}
```

**Strengths:**
- ✅ RSA signing (public key available for validation)
- ✅ Tenant ID included (`tid`)
- ✅ Roles included for RBAC
- ✅ Proper expiry and issued-at times

**Potential Gaps:**
- Missing `full_name` field (frontend will need separate lookup)
- `token_hash` only in refresh tokens (good for revocation)

### 5.2 Tenant Isolation ✅ DATABASE-ENFORCED

**RLS Enforcement:**
```sql
CREATE POLICY tenant_isolation ON employees
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

**Pattern (from code):**
```go
shared.WithTenantTx(ctx, pool, tenantID, func(ctx, tx) error {
    tx.Exec(ctx, `SELECT set_config('app.tenant_id', $1, true)`, tenantID.String())
    // Query automatically filtered by RLS
})
```

**Assessment:**
- ✅ RLS prevents accidental cross-tenant access
- ✅ Works at the database level (most secure)
- ⚠️ Application layer should also validate (defense in depth)

### 5.3 Permission Enforcement ❌ UNCLEAR LOCATION

**Current state:**
- No code found enforcing role-based access in handlers
- Spatie Permission not yet integrated with gateway
- Leave approval (requires manager role) — unclear who enforces

**Questions:**
- Where is `ApproveLeave` permission checked?
- Who verifies user has `manager` role before approval?

**For Integration:**
- Gateway must check permissions via Spatie
- OR services must validate roles from JWT claims

---

## PART 6: END-TO-END FLOW AUDIT

### Critical Path: Login → Employee List → Leave Request → Approval

```
Step 1: Login
  Frontend POST /api/auth/login {email, password, tenant_slug}
    ↓
  Gateway calls Auth Service ✅ (implemented)
    ↓
  Returns: {access_token, refresh_token, claims}
  Status: ✅ FUNCTIONAL

Step 2: Employee List
  Frontend GET /api/employees -H "Authorization: Bearer {token}"
    ↓
  Gateway validates JWT ❌ (NOT IMPLEMENTED)
  Gateway extracts tenant_id ❌ (NOT IMPLEMENTED)
    ↓
  Gateway calls Employee Service ❌ (NO SERVICE CLASS)
    ↓
  Employee Service returns employees ✅ (implemented, but missing names)
  Status: ⚠️ PARTIALLY BLOCKED

Step 3: Leave Request
  Frontend POST /api/leave/apply {employee_id, leave_type_id, start_date, end_date, reason}
    ↓
  Gateway validates JWT ❌ (NOT IMPLEMENTED)
  Gateway enforces tenant_id ❌ (NOT IMPLEMENTED)
    ↓
  Gateway calls Leave Service ❌ (LEAVE SERVICE UNIMPLEMENTED)
    ↓
  Leave service creates request ❌ (handlers return Unimplemented)
  Status: ❌ BLOCKED

Step 4: Manager Approval
  Frontend POST /api/leave/{id}/approve
    ↓
  Gateway validates manager role ❌ (NOT IMPLEMENTED)
    ↓
  Leave service approves ❌ (handlers return Unimplemented)
    ↓
  Publishes leave.approved event ❌ (can't test without handlers)
  Notification service receives event ✅ (consumer implemented)
  Status: ❌ BLOCKED
```

**Missing Components in Flow:**
1. Gateway JWT validation and extraction
2. Gateway tenant_id enforcement
3. Gateway service client classes
4. Leave service handler implementations
5. Permission checks (at gateway and/or service)
6. Request/response transformation

---

## PART 7: RISK ASSESSMENT

### 🔴 CRITICAL BLOCKERS (Must fix before integration)

1. **Leave Service Unimplemented** (8/8 handlers)
   - Estimated effort: 2 days
   - Blocks entire leave workflow
   - Domain layer complete, just needs handler wiring

2. **Gateway Tenant Validation** (Security)
   - Estimated effort: 2 days
   - Required to prevent cross-tenant data access
   - Non-negotiable security control

3. **Gateway Service Client Layer** (Infrastructure)
   - Estimated effort: 2-3 days
   - Needed for all service calls
   - Foundation for all gateway routes

---

### 🟡 HIGH-RISK ASSUMPTIONS

1. **"Request bodies can include tenant_id"**
   - Proto contracts expose tenant_id in requests
   - RLS prevents violation, but bad pattern
   - Gateway must override requests with JWT tenant_id

2. **"Denormalized fields will be populated"**
   - Proto defines `employee_name`, `department_name`, etc.
   - Handlers don't populate them
   - Frontend will need to fetch separately or handlers need updates

3. **"Permission checks happen automatically"**
   - No evidence of role enforcement in handlers
   - Approval operations have no visible permission guards
   - Gateway or service must validate roles from JWT

4. **"Frontend can trust tenant_id from JWT"**
   - True IF gateway enforces it
   - False IF gateway accepts request body tenant_id
   - Critical decision for gateway implementation

---

## PART 8: READINESS SCORES

### Gateway Readiness Score: **5/100** 🔴 NOT READY

- **Infrastructure:** 0/30 (no client layer, no middleware)
- **Authentication:** 5/20 (JWT generation works, validation missing)
- **Error Handling:** 0/20 (no error translation)
- **Request/Response:** 0/20 (no controllers)
- **Service Integration:** 0/10 (no service classes)

### Frontend Readiness Score: **0/100** 🔴 NOT READY

- **Pages:** 0/30 (no routes)
- **Auth Guard:** 0/20 (no hooks.server.ts)
- **Data Layer:** 0/20 (no TanStack Query setup)
- **Components:** 0/20 (no UI components)
- **Testing:** 0/10 (no tests)

### Security Readiness Score: **60/100** ⚠️ PARTIAL

- **Database Isolation:** 20/20 (RLS enforced)
- **JWT Structure:** 15/20 (good, missing full_name)
- **Signature Validation:** 0/20 (not in gateway yet)
- **Permission Enforcement:** 15/20 (unclear who enforces)
- **Tenant Validation:** 10/20 (database prevents violation, app layer missing)

### Service Readiness Score: **75/100** ⚠️ MOSTLY READY

- **Auth Service:** 20/20 (fully implemented)
- **Employee Service:** 17/20 (implemented, missing denormalization)
- **Attendance Service:** 20/20 (fully implemented)
- **Leave Service:** 5/20 (unimplemented handlers)
- **Notification Service:** 10/10 (fully implemented)
- **Audit Service:** 10/10 (fully implemented)
- **Document Service:** 10/10 (fully implemented)

---

## PART 9: RECOMMENDED IMPLEMENTATION ORDER

### Phase 5A: Leave Service Wiring (CRITICAL PATH BLOCKER)
**Effort:** 2 days
**Why First:** Without this, ~50% of the E2E flow is broken

1. Wire `main.go` (DB, NATS, handlers)
2. Implement 8 gRPC handlers
3. Verify leave workflow end-to-end

### Phase 5B: Gateway Foundation (INFRASTRUCTURE BLOCKER)
**Effort:** 3 days
**Why Second:** Needed for all subsequent integration

1. Create gRPC client wrapper (Guzzle-based)
2. Create JWT validation middleware
3. Create tenant enforcement middleware
4. Create service client classes
5. Create error translation utility

### Phase 5C: Gateway Routes (INTEGRATION)
**Effort:** 2-3 days
**Why Third:** Controllers are simple once infrastructure exists

1. Auth routes (login, refresh, validate)
2. Employee routes (list, create, get, terminate)
3. Leave routes (apply, approve, reject, list, get, balance, types)
4. Attendance routes (check-in, check-out, list, summary)
5. Document and notification routes

### Phase 5D: Frontend (UI LAYER)
**Effort:** 5-7 days
**Why Last:** Requires working backend to test against

1. Auth flow (login, token storage, auto-refresh)
2. Layout (sidebar, route guards, dashboard)
3. Employee pages (list, detail, terminate)
4. Leave pages (request form, approval view)
5. Attendance page (check-in button, history)

---

## PART 10: ESTIMATED TIMELINE

| Component | Days | Status |
|-----------|------|--------|
| Leave Service | 2 | 🔴 BLOCKED on Phase 5A |
| Gateway | 3 | 🔴 BLOCKED on Phase 5A |
| Frontend | 5 | 🔴 BLOCKED on Phases 5A & 5B |
| Testing & Integration | 2 | 🔴 BLOCKED on all |
| **TOTAL MVP** | **~12-14 days** | 🔴 BLOCKED |

---

## PART 11: KEY DECISIONS FOR GATEWAY IMPLEMENTATION

**Decision 1: Tenant Validation Strategy**
- RECOMMENDATION: Override request body tenant_id with JWT claim
- Rationale: Prevents any possibility of cross-tenant access via tampering

**Decision 2: Denormalized Field Population**
- RECOMMENDATION: Update employee handler to lookup and populate names
- Rationale: Better UX and avoids N+1 queries from frontend

**Decision 3: Permission Enforcement Location**
- RECOMMENDATION: Validate at Gateway (coarse-grained) + Service (fine-grained)
- Rationale: Defense in depth, prevents escalation-of-privilege bugs

**Decision 4: Error Response Format**
- RECOMMENDATION: Standardize: `{code, message, details}` (current gateway pattern)
- Rationale: Consistent with gRPC error model, easy for frontend

---

## Summary Table: Critical Path Blockers

| Issue | Severity | Effort | Blocker For |
|-------|----------|--------|-------------|
| Leave service unimplemented | 🔴 CRITICAL | 2 days | All leave workflows |
| Gateway infrastructure missing | 🔴 CRITICAL | 3 days | All service calls |
| Gateway JWT validation | 🔴 CRITICAL | 1 day | All authenticated routes |
| Gateway tenant enforcement | 🔴 CRITICAL | 1 day | Security |
| Employee denormalization | 🟡 HIGH | 1 day | UX (N+1 avoidance) |
| Frontend empty | 🟡 HIGH | 5-7 days | User interface |
| Permission enforcement unclear | 🟡 HIGH | 2 days | Role-based access |
| Pagination incomplete | 🟡 MEDIUM | 1 day | Large result sets |

---

**Report Status:** Code-level audit complete. Safe to proceed to implementation with blockers identified. Recommend starting Phase 5A (Leave Service) before Gateway.

