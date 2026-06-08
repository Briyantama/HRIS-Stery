# HRIS-Stery Gateway API Specification

**Version:** 1.0  
**Base URL:** `http://localhost:8000/api/v1`  
**Authentication:** JWT (RSA-256)  
**Content-Type:** `application/json`

---

## Authentication & Tenant Context

All protected endpoints require:
1. **Authorization Header:** `Authorization: Bearer <JWT_TOKEN>`
2. **JWT Claims:** Must contain `sub`, `tid` (tenant_id), `roles`
3. **Tenant Isolation:** Gateway enforces `tenant_id` from JWT; request body values ignored

### JWT Token Format (from Auth Service)
```json
{
  "sub": "user-uuid",
  "user_id": "user-uuid",
  "tid": "tenant-uuid",
  "tenant_id": "tenant-uuid",
  "email": "user@example.com",
  "roles": ["hr_admin", "manager"],
  "exp": 1717000000,
  "iat": 1716996400
}
```

### Error Response Format (Standard)
```json
{
  "code": "ERROR_CODE",
  "message": "Human-readable error message",
  "details": ["Additional context if applicable"]
}
```

---

## Leave Management API

### 1. Apply for Leave
**POST** `/leaves`

Create a new leave request.

**Request Body:**
```json
{
  "employee_id": "550e8400-e29b-41d4-a716-446655440000",
  "leave_type_id": "660e8400-e29b-41d4-a716-446655440000",
  "start_date": "2026-06-15",
  "end_date": "2026-06-20",
  "reason": "Annual vacation",
  "document_key": "uploads/leave-docs/abc123.pdf"
}
```

**Response (201 Created):**
```json
{
  "code": "SUCCESS",
  "message": "Leave request created successfully",
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "employee_id": "550e8400-e29b-41d4-a716-446655440000",
    "leave_type_id": "660e8400-e29b-41d4-a716-446655440000",
    "status": "PENDING",
    "start_date": "2026-06-15",
    "end_date": "2026-06-20",
    "days_count": 6,
    "reason": "Annual vacation",
    "created_at": "2026-06-08T10:30:00Z",
    "updated_at": "2026-06-08T10:30:00Z"
  }
}
```

**Validation Errors (400 Bad Request):**
```json
{
  "code": "INVALID_ARGUMENT",
  "message": "Validation failed",
  "details": [
    "The start_date field is required.",
    "The end_date must be a date after or equal to start_date."
  ]
}
```

---

### 2. List Leave Requests
**GET** `/leaves?employee_id=uuid&status=APPROVED&year=2026&page_size=20`

Retrieve leave requests with optional filters. Returns paginated results.

**Query Parameters:**
- `employee_id` (uuid, optional): Filter by employee
- `approver_id` (uuid, optional): Filter by approver (manager view)
- `status` (enum, optional): PENDING, APPROVED, REJECTED, CANCELLED
- `year` (int, optional): Filter by year
- `page_size` (int, optional, default: 50, max: 100): Results per page
- `page_token` (string, optional): Pagination token

**Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "data": {
    "leave_requests": [
      {
        "id": "770e8400-e29b-41d4-a716-446655440000",
        "employee_id": "550e8400-e29b-41d4-a716-446655440000",
        "status": "APPROVED",
        "start_date": "2026-06-15",
        "end_date": "2026-06-20",
        "days_count": 6,
        "reason": "Annual vacation",
        "approved_by_id": "880e8400-e29b-41d4-a716-446655440000",
        "approved_at": "2026-06-08T11:00:00Z",
        "created_at": "2026-06-08T10:30:00Z"
      }
    ],
    "total_count": 1,
    "next_page_token": null
  }
}
```

---

### 3. Get Leave Request Details
**GET** `/leaves/{id}`

Fetch a specific leave request by ID.

**Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "employee_id": "550e8400-e29b-41d4-a716-446655440000",
    "leave_type_id": "660e8400-e29b-41d4-a716-446655440000",
    "status": "PENDING",
    "start_date": "2026-06-15",
    "end_date": "2026-06-20",
    "days_count": 6,
    "reason": "Annual vacation",
    "created_at": "2026-06-08T10:30:00Z",
    "updated_at": "2026-06-08T10:30:00Z"
  }
}
```

**Not Found (404):**
```json
{
  "code": "NOT_FOUND",
  "message": "Leave request not found"
}
```

---

### 4. Approve Leave Request
**POST** `/leaves/{id}/approve`

**RBAC:** Requires `manager` or `hr_admin` role

**Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "message": "Leave request approved successfully",
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "status": "APPROVED",
    "approved_by_id": "880e8400-e29b-41d4-a716-446655440000",
    "approved_at": "2026-06-08T11:30:00Z"
  }
}
```

**Permission Denied (403):**
```json
{
  "code": "PERMISSION_DENIED",
  "message": "Only managers and HR admins can approve leave requests"
}
```

---

### 5. Reject Leave Request
**POST** `/leaves/{id}/reject`

**RBAC:** Requires `manager` or `hr_admin` role

**Request Body:**
```json
{
  "reason": "Insufficient leave balance for this period"
}
```

**Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "message": "Leave request rejected successfully",
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "status": "REJECTED",
    "rejection_reason": "Insufficient leave balance for this period"
  }
}
```

---

### 6. Cancel Leave Request
**POST** `/leaves/{id}/cancel`

Employee can cancel their own requests. HR admin can cancel any request.

**Request Body:**
```json
{
  "employee_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "message": "Leave request cancelled successfully",
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "status": "CANCELLED"
  }
}
```

**Permission Denied (403) - Not Your Request:**
```json
{
  "code": "PERMISSION_DENIED",
  "message": "You can only cancel your own leave requests"
}
```

---

### 7. Get Leave Balance
**GET** `/leave-balance/{employeeId}?year=2026`

Retrieve leave balance breakdown by leave type for an employee.

**Query Parameters:**
- `year` (int, optional, default: current year)

**Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "data": {
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
    "employee_id": "550e8400-e29b-41d4-a716-446655440000",
    "year": 2026,
    "balances": [
      {
        "leave_type_id": "660e8400-e29b-41d4-a716-446655440000",
        "leave_type_code": "ANNUAL",
        "leave_type_name": "Annual Leave",
        "entitled_days": 20,
        "used_days": 6,
        "pending_days": 6,
        "remaining_days": 8
      },
      {
        "leave_type_id": "770e8400-e29b-41d4-a716-446655440000",
        "leave_type_code": "SICK",
        "leave_type_name": "Sick Leave",
        "entitled_days": 10,
        "used_days": 2,
        "pending_days": 0,
        "remaining_days": 8
      }
    ]
  }
}
```

---

### 8. List Leave Types
**GET** `/leave-types`

Retrieve all leave types configured for the tenant.

**Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "code": "ANNUAL",
      "name": "Annual Leave",
      "max_days_per_year": 20,
      "requires_document": false,
      "is_paid": true
    },
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "code": "SICK",
      "name": "Sick Leave",
      "max_days_per_year": 10,
      "requires_document": true,
      "is_paid": true
    }
  ]
}
```

---

## HTTP Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| 200 | OK | Successfully retrieved/updated resource |
| 201 | Created | New resource created (POST /leaves) |
| 400 | Bad Request | Validation error, malformed request |
| 401 | Unauthorized | Missing/invalid JWT token |
| 403 | Forbidden | Insufficient permissions (RBAC) |
| 404 | Not Found | Resource not found |
| 500 | Internal Server Error | Unhandled server error |
| 503 | Service Unavailable | Backend service unreachable |

---

## Error Code Reference

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `SUCCESS` | 200/201 | Operation completed successfully |
| `INVALID_ARGUMENT` | 400 | Validation error or bad request |
| `UNAUTHENTICATED` | 401 | Missing or invalid authentication |
| `PERMISSION_DENIED` | 403 | User lacks required role/permission |
| `NOT_FOUND` | 404 | Requested resource not found |
| `ALREADY_EXISTS` | 409 | Resource already exists |
| `INTERNAL` | 500 | Server-side error |
| `UNAVAILABLE` | 503 | Service temporarily unavailable |

---

## Example cURL Requests

### Apply for Leave
```bash
curl -X POST http://localhost:8000/api/v1/leaves \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_id": "550e8400-e29b-41d4-a716-446655440000",
    "leave_type_id": "660e8400-e29b-41d4-a716-446655440000",
    "start_date": "2026-06-15",
    "end_date": "2026-06-20",
    "reason": "Annual vacation"
  }'
```

### Approve Leave (Manager Only)
```bash
curl -X POST http://localhost:8000/api/v1/leaves/770e8400-e29b-41d4-a716-446655440000/approve \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json"
```

### Get Leave Balance
```bash
curl -X GET "http://localhost:8000/api/v1/leave-balance/550e8400-e29b-41d4-a716-446655440000?year=2026" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

---

## Implementation Notes

1. **Tenant Isolation:** Gateway enforces `tenant_id` from JWT; frontend values are ignored
2. **RBAC:** Approval/rejection require `manager` or `hr_admin` role
3. **Idempotency:** Callers should implement request idempotency via `x-request-id` header for retry safety
4. **Rate Limiting:** Gateway implements per-tenant rate limiting (100 req/min default)
5. **Tracing:** All requests include `x-request-id` header for distributed tracing

---

**API Status:** ✅ Ready for Phase 5C (Frontend Implementation)
