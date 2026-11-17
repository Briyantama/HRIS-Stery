# Architecture Review Skill

**Trigger:** Before implementing HTTP handlers or repositories

**Purpose:** Enforce Clean Architecture boundaries — no business logic in I/O layers

## Architecture Boundaries (STRICT)

```
Domain Layer (pure, no I/O)
  ↓ depends on
Application Layer (use cases, business logic)
  ↓ depends on
Infrastructure Layer (I/O: DB, cache, event stream)
  ↓ depends on
http.go (HTTP adapter, thin)
```

**Rule:** Code can only depend on layers below it. Never upward.

## Forbidden Patterns

### Go Services

**❌ FORBIDDEN: Business logic in HTTP handlers**
```go
func (h *HTTPHandler) createEmployee(w http.ResponseWriter, r *http.Request) {
    // ❌ Validation logic here
    if body.Email == "" { httpx.WriteError(w, 422, "email required"); return }

    // ❌ Calculation here
    salary := body.BaseSalary * 1.13

    // ❌ DB queries here
    h.pool.Exec(r.Context(), "INSERT INTO employees...")
}
```

**✅ CORRECT: Handler is thin, delegates to application layer**
```go
func (h *HTTPHandler) createEmployee(w http.ResponseWriter, r *http.Request) {
    claims, _ := authn.FromContext(r.Context())
    var body createEmployeeBody
    if !httpx.DecodeJSON(w, r, &body, maxBody) {
        return
    }
    result, err := h.create.Handle(r.Context(), commands.CreateEmployeeCommand{
        TenantID: claims.TenantID, // from the token, never from the body
        Email:    body.Email,
    })
    if err != nil {
        h.writeErr(w, r, err)
        return
    }
    httpx.WriteSuccess(w, http.StatusCreated, result)
}
```

### Repository Pattern

**❌ FORBIDDEN: Query logic in repository**
```go
func (r *EmployeeRepository) GetActiveByDepartment(ctx context.Context, deptID string) ([]*Employee, error) {
    // ❌ Business logic filtering
    employees := []*Employee{}
    for _, emp := range all {
        if emp.status == "ACTIVE" && emp.dept == deptID {
            employees = append(employees, emp)
        }
    }
    return employees, nil
}
```

**✅ CORRECT: Repository handles data access, domain handles logic**
```go
func (r *EmployeeRepository) ListByTenant(ctx context.Context, tenantID TenantID, filters Filters) ([]*Employee, error) {
    // ✅ Just data access
    query := `SELECT ... FROM employees WHERE tenant_id = $1 AND status = $2`
    return executeQuery(ctx, query, tenantID, filters.Status)
}

// Business logic in application layer
func (h *ListEmployeesHandler) Handle(ctx context.Context, query ListEmployeesQuery) (*ListEmployeesResult, error) {
    employees, err := h.repo.ListByTenant(ctx, tenantID, filters)
    // ✅ Application layer applies additional filtering if needed
    return filtered, nil
}
```

## Checklist for Each Layer

### Domain Layer (services/<name>/domain/)
- [ ] No I/O (no DB, HTTP, files)
- [ ] No external dependencies except stdlib + value objects
- [ ] Pure business logic only
- [ ] Aggregates, value objects, domain events
- [ ] Invariants enforced

### Application Layer (services/<name>/application/)
- [ ] Orchestrates domain and infrastructure
- [ ] Commands (write side) and Queries (read side)
- [ ] Calls domain methods and repository interfaces
- [ ] Handles use case logic
- [ ] Publishes domain events
- [ ] No direct DB queries (via repositories only)

### Infrastructure Layer (services/<name>/infrastructure/)
- [ ] Implements repository interfaces
- [ ] Handles I/O: DB, cache, event stream
- [ ] No business logic, only data access and formatting
- [ ] RLS context injection (WithTenantTx)
- [ ] Error mapping for domain errors

### HTTP adapter (services/<name>/http.go)
- [ ] Routes registered on `*http.ServeMux` with `"VERB /api/v1/..."` patterns
- [ ] Request decoding and validation of shape only (`httpx.DecodeJSON`)
- [ ] Tenant and actor ids from `authn.FromContext`
- [ ] Delegates to application layer
- [ ] Error mapping to HTTP status, answers through `httpx.WriteSuccess` / `httpx.WriteError`

## Examples

**Trigger:** "Add validation to CreateEmployee"
→ Put validation in application layer (CreateEmployeeHandler), not the HTTP handler

**Trigger:** "Filter employees by department"
→ Put filter in query handler, not repository

**Trigger:** "Calculate employee tax"
→ Put calculation in domain (Employee aggregate), not the HTTP handler
