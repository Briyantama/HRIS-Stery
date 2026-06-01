# Go Style Review Skill

**Trigger:** Before reviewing any Go code changes

**Purpose:** Enforce HRIS-Stery Go conventions and prevent common errors

## Type Safety (No Escape Hatches)

**❌ FORBIDDEN in domain types:**
```go
type Employee struct {
    TenantID interface{}  // ❌ use TenantID value object
    ID       interface{}  // ❌ use EmployeeID value object
    Name     string       `json:"name,omitempty"` // ❌ omitempty hides required fields
}

var db *sql.DB  // ❌ global state
```

**✅ CORRECT:**
```go
type Employee struct {
    tenantID TenantID     // private, accessed via method
    id       EmployeeID   // private, accessed via method
    email    string       // required, no omitempty
}

func (e *Employee) TenantID() TenantID { return e.tenantID }
func (e *Employee) ID() EmployeeID { return e.id }
```

## Domain Value Objects (Mandatory)

For any ID field, use a typed value object:

```go
type EmployeeID string

func NewEmployeeID(id string) (EmployeeID, error) {
    if id == "" {
        return "", errors.New("invalid employee id")
    }
    return EmployeeID(id), nil
}

func (e EmployeeID) String() string { return string(e) }
func (e EmployeeID) IsZero() bool { return e == "" }
func (e EmployeeID) Equals(other EmployeeID) bool { return e == other }
```

## RLS Context (WithTenantTx)

**❌ FORBIDDEN: Raw connection for tenant-scoped queries**
```go
func (r *EmployeeRepository) GetByID(ctx context.Context, tenantID TenantID, id EmployeeID) (*Employee, error) {
    conn, _ := r.pool.Acquire(ctx)  // ❌ no RLS context set
    query := `SELECT ... FROM employees WHERE id = $1`
    return scanEmployee(conn.QueryRow(ctx, query, id.String()))
}
```

**✅ CORRECT: Use WithTenantTx for RLS context**
```go
func (r *EmployeeRepository) GetByID(ctx context.Context, tenantID TenantID, id EmployeeID) (*Employee, error) {
    var emp *Employee
    err := shared.WithTenantTx(ctx, r.pool, tenantID, func(ctx context.Context, tx pgx.Tx) error {
        // RLS context set automatically before this point
        query := `SELECT ... FROM employees WHERE id = $1`
        return scanEmployee(tx.QueryRow(ctx, query, id.String()), &emp)
    })
    return emp, err
}
```

## Error Handling

**Checklist:**
- [ ] Errors wrapped with context: `fmt.Errorf("describe operation: %w", err)`
- [ ] gRPC handlers return status codes: `status.Errorf(codes.NotFound, "...")`
- [ ] Domain errors defined in domain package
- [ ] No stack traces in logs
- [ ] Sensitive data masked in error messages

**✅ CORRECT:**
```go
emp, err := r.repo.GetByID(ctx, tenantID, id)
if err != nil {
    if err == ErrNotFound {
        return nil, status.Error(codes.NotFound, "employee not found")
    }
    return nil, status.Error(codes.Internal, "database error")
}
```

## Dependency Injection

**❌ FORBIDDEN: Package-level variables**
```go
var db *pgxpool.Pool  // ❌ global state
var logger *zap.Logger  // ❌ global state

func init() {  // ❌ implicit initialization
    var err error
    db, err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
}
```

**✅ CORRECT: Constructor injection**
```go
type EmployeeRepository struct {
    pool *pgxpool.Pool
    logger *zap.Logger
}

func NewEmployeeRepository(pool *pgxpool.Pool, logger *zap.Logger) *EmployeeRepository {
    return &EmployeeRepository{pool: pool, logger: logger}
}
```

## Testing

**Checklist:**
- [ ] Unit tests for domain layer (no I/O)
- [ ] Mock repositories in application tests
- [ ] Table-driven tests for validation logic
- [ ] Integration tests with `//go:build integration` tag
- [ ] No direct DB access in unit tests

## Code Review Checklist

- [ ] No `interface{}` in domain types
- [ ] No package-level global variables
- [ ] All I/O dependencies injected via constructor
- [ ] All tenant-scoped queries use `WithTenantTx()`
- [ ] Error wrapping with context
- [ ] gRPC errors use correct status codes
- [ ] Tests for domain logic (domain package tests)
- [ ] Comments explain WHY, not WHAT

## Examples

**Trigger:** "Review EmployeeRepository.GetByID"
→ Check that it uses WithTenantTx, not raw connection

**Trigger:** "Review CreateEmployeeHandler"
→ Check that it calls domain and infrastructure, not DB directly

**Trigger:** "Review validation logic"
→ Check that it's in domain (Employee.NewEmployee), not controller
