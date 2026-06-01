# Architecture Review Skill

**Trigger:** Before implementing handlers, controllers, or repositories

**Purpose:** Enforce Clean Architecture boundaries — no business logic in I/O layers

## Architecture Boundaries (STRICT)

```
Domain Layer (pure, no I/O)
  ↓ depends on
Application Layer (use cases, business logic)
  ↓ depends on
Infrastructure Layer (I/O: DB, cache, NATS)
  ↓ depends on
Interface Layer (HTTP/gRPC/CLI, thin)
```

**Rule:** Code can only depend on layers below it. Never upward.

## Forbidden Patterns

### Go Services

**❌ FORBIDDEN: Business logic in gRPC handlers**
```go
func (h *EmployeeServiceServer) CreateEmployee(ctx context.Context, req *pb.CreateEmployeeRequest) (*pb.EmployeeResponse, error) {
    // ❌ Validation logic here
    if req.Email == "" { return nil, status.Error(...) }
    
    // ❌ Calculation here
    salary := req.BaseSalary * 1.13
    
    // ❌ DB queries here
    db.Exec("INSERT INTO employees...")
}
```

**✅ CORRECT: Handler is thin, delegates to application layer**
```go
func (h *EmployeeServiceServer) CreateEmployee(ctx context.Context, req *pb.CreateEmployeeRequest) (*pb.EmployeeResponse, error) {
    result, err := h.createEmployeeHandler.Handle(ctx, commands.CreateEmployeeCommand{
        TenantID: req.TenantId,
        Email: req.Email,
        // ...
    })
    if err != nil {
        return nil, grpcError(err)
    }
    return &pb.EmployeeResponse{...}, nil
}
```

### Laravel Gateway

**❌ FORBIDDEN: Business logic in controller**
```php
public function store(Request $request) {
    $salary = $request->base * 1.13; // ❌ domain logic
    Employee::create($request->all()); // ❌ raw DB access
    return response()->json(['ok' => true]);
}
```

**✅ CORRECT: Controller delegates to service**
```php
public function store(CreateEmployeeRequest $request) {
    $employee = $this->employeeService->create($request->validated());
    return response()->json(new EmployeeResource($employee));
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

### Domain Layer (internal/domain/)
- [ ] No I/O (no DB, HTTP, gRPC, files)
- [ ] No external dependencies except stdlib + value objects
- [ ] Pure business logic only
- [ ] Aggregates, value objects, domain events
- [ ] Invariants enforced

### Application Layer (internal/application/)
- [ ] Orchestrates domain and infrastructure
- [ ] Commands (write side) and Queries (read side)
- [ ] Calls domain methods and repository interfaces
- [ ] Handles use case logic
- [ ] Publishes domain events
- [ ] No direct DB queries (via repositories only)

### Infrastructure Layer (internal/infrastructure/)
- [ ] Implements repository interfaces
- [ ] Handles I/O: DB, cache, NATS, HTTP calls
- [ ] No business logic, only data access and formatting
- [ ] RLS context injection (WithTenantTx)
- [ ] Error mapping for domain errors

### Interface Layer (internal/interfaces/)
- [ ] gRPC handlers (services/*/internal/interfaces/grpc/)
- [ ] Laravel controllers (gateway/app/Http/Controllers/)
- [ ] Request validation only
- [ ] Delegates to application layer
- [ ] Error mapping to protocol format (gRPC status codes, HTTP status)
- [ ] Maximum 20 lines of logic per endpoint (Laravel rule)

## Examples

**Trigger:** "Add validation to CreateEmployee"
→ Put validation in application layer (CreateEmployeeHandler), not gRPC handler

**Trigger:** "Filter employees by department"
→ Put filter in query handler, not repository

**Trigger:** "Calculate employee tax"
→ Put calculation in domain (Employee aggregate), not controller
