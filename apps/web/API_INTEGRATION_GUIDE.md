# SvelteKit Frontend - API Integration Guide

## Architecture Overview

```
SvelteKit Frontend
  ├─ Auth Store (Svelte Writable)
  │  └─ JWT token + user claims (persistent via localStorage)
  │
  ├─ API Client Wrapper
  │  ├─ Auto-inject Authorization header
  │  ├─ Handle 401 → redirect to /login
  │  └─ Request ID generation for tracing
  │
  ├─ TanStack Query Hooks
  │  ├─ useLeaveBalance() → fetch employee balance
  │  ├─ useLeaveTypes() → fetch all leave types
  │  ├─ useLeaveRequests() → list with filters
  │  └─ useMutation hooks → apply/approve/reject
  │
  └─ SvelteKit Pages
     ├─ (app)/dashboard → Balance display
     ├─ (app)/leaves/apply → Request form
     └─ (app)/admin/leaves → Approval portal
```

---

## 1. Authentication Flow

### Login (Frontend → Backend)
```typescript
// user clicks login, form submits email/password
POST /auth/login
Body: { email, password, tenant_slug }
Response: { access_token, refresh_token, claims: { user_id, tenant_id, roles, email } }
```

### Store Session (Auth Store)
```typescript
import { auth } from '$lib/stores/auth';

// After successful login:
auth.setSession(token, {
  user_id: 'uuid',
  tenant_id: 'uuid',
  email: 'user@example.com',
  roles: ['manager']
});

// Token + user data persisted to localStorage
// Available in subsequent page loads
```

### Automatic Token Injection (API Client)
```typescript
// Every API call automatically includes:
// Authorization: Bearer <token>
// X-Request-ID: <generated-uuid>

const response = await apiGet('/leaves');
// Sends: Authorization: Bearer eyJhbGc...
```

### 401 Handling (Session Expiry)
```typescript
// If server returns 401 Unauthorized:
// 1. auth.clearSession() removes token + user data
// 2. localStorage cleared
// 3. User redirected to /login
// 4. Page refreshes with auth store empty
```

---

## 2. API Client Usage Examples

### GET Requests
```typescript
import { apiGet } from '$lib/api/client';

// Fetch leave balance
const response = await apiGet('/leave-balance/user-id', {
  query: { year: 2026 }
});
console.log(response.data); // LeaveBalance object
```

### POST Requests
```typescript
import { apiPost } from '$lib/api/client';

// Apply for leave
const response = await apiPost('/leaves', {
  employee_id: 'uuid',
  leave_type_id: 'uuid',
  start_date: '2026-06-15',
  end_date: '2026-06-20',
  reason: 'Vacation'
});
console.log(response.data); // LeaveRequest object
```

### Error Handling
```typescript
import { apiGet, ApiCallError, getErrorMessage } from '$lib/api/client';

try {
  const response = await apiGet('/leaves/invalid-id');
} catch (error) {
  if (error instanceof ApiCallError) {
    console.error(error.code); // 'NOT_FOUND'
    console.error(error.message); // 'Leave request not found'
    console.error(error.status); // 404
  }
}
```

---

## 3. TanStack Query Hook Usage

### Simple Query (Fetch with Caching)
```typescript
<script lang="ts">
  import { useLeaveBalance } from '$lib/queries/leave';
  import { user } from '$lib/stores/auth';

  let userId: string | null = null;
  user.subscribe((u) => {
    userId = u?.user_id ?? null;
  });

  $: leaveBalance = useLeaveBalance(userId);
</script>

<!-- Auto-loading, caching, refetch on focus -->
{#if $leaveBalance.isLoading}
  <p>Loading...</p>
{:else if $leaveBalance.isError}
  <p>Error: {$leaveBalance.error.message}</p>
{:else}
  <p>Balance: {$leaveBalance.data.balances[0].remaining_days} days</p>
{/if}
```

### Mutation (Create/Update)
```typescript
<script lang="ts">
  import { useApplyLeave } from '$lib/queries/leave';

  const applyLeave = useApplyLeave();

  async function handleSubmit(form: ApplyLeaveForm) {
    try {
      const result = await $applyLeave.mutateAsync(form);
      console.log('Leave applied:', result);
    } catch (error) {
      console.error('Failed:', error.message);
    }
  }
</script>

<button on:click={() => handleSubmit(form)}>
  {$applyLeave.isPending ? 'Applying...' : 'Apply Leave'}
</button>
```

---

## 4. Store Usage

### Check Authentication
```typescript
import { isAuthenticated, user, roles, hasRole, isManager } from '$lib/stores/auth';

<!-- Conditional rendering -->
{#if $isAuthenticated}
  <p>Welcome, {$user.email}</p>
{:else}
  <p>Please log in</p>
{/if}

<!-- Role checks -->
{#if $isManager}
  <a href="/admin/leaves">Approval Portal</a>
{/if}

<!-- Dynamic role check -->
{@const canApprove = $hasRole('manager')}
{#if canApprove}
  <button on:click={approveLeave}>Approve</button>
{/if}
```

### Extract User Data
```typescript
import { user, tenantId } from '$lib/stores/auth';

$: userId = $user?.user_id;
$: tenant = $tenantId;
```

---

## 5. Route Guards (hooks.server.ts)

### Protect Routes from Unauthenticated Access
```typescript
// hooks.server.ts
import { redirect } from '@sveltejs/kit';
import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
  // Check if user is authenticated
  const token = event.cookies.get('auth_token');

  // Protect (app) routes
  if (event.url.pathname.startsWith('/dashboard') || event.url.pathname.startsWith('/leaves')) {
    if (!token) {
      throw redirect(303, '/login');
    }
  }

  return await resolve(event);
};
```

---

## 6. Environment Variables

### .env.local (Frontend)
```bash
# API Gateway URL
VITE_API_URL=http://localhost:8000/api/v1

# Enable debug logging
VITE_DEBUG=true
```

### Usage
```typescript
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8000/api/v1';
```

---

## 7. Validation Schemas

### Zod Schemas for Type Safety
```typescript
import { ApplyLeaveRequestSchema } from '$lib/schemas/leave';

// Validate form data before submission
const form = {
  employee_id: 'uuid',
  leave_type_id: 'uuid',
  start_date: '2026-06-15',
  end_date: '2026-06-10' // ERROR: end < start
};

try {
  const validated = ApplyLeaveRequestSchema.parse(form);
} catch (error) {
  console.error(error.errors[0].message); // 'End date must be on or after start date'
}
```

---

## 8. Error Handling Pattern

### Graceful Error Display in Components
```typescript
<script lang="ts">
  import { getErrorMessage, getValidationErrors } from '$lib/api/client';

  let error: string = '';
  let fieldErrors: Record<string, string[]> = {};

  async function submitForm(form: ApplyLeaveForm) {
    try {
      const result = await $applyLeave.mutateAsync(form);
      // Success
    } catch (err) {
      error = getErrorMessage(err);
      fieldErrors = getValidationErrors(err);
    }
  }
</script>

<!-- Display errors -->
{#if error}
  <div class="alert alert-error">{error}</div>
{/if}

{#each Object.entries(fieldErrors) as [field, messages]}
  <div class="field-error">
    <p class="label">{field}</p>
    {#each messages as msg}
      <p class="error">{msg}</p>
    {/each}
  </div>
{/each}
```

---

## 9. Request Lifecycle with Metadata

### Headers Added to Every Request
```
GET /api/v1/leaves HTTP/1.1
Authorization: Bearer eyJhbGc...
X-Request-ID: 1717000000-a1b2c3d4e
Content-Type: application/json

[Response]
HTTP 200 OK
{
  "code": "SUCCESS",
  "data": {
    "leave_requests": [...],
    "total_count": 25,
    "next_page_token": null
  }
}
```

---

## 10. Session Persistence

### Logout & Re-login

```typescript
import { auth } from '$lib/stores/auth';

// Logout
auth.clearSession();
// Clears: localStorage, store, redirects to /login

// After page reload
// Auth store checks localStorage and restores session automatically
```

---

## Status: ✅ Data Layer Complete

The foundational data connectivity layer is ready:
- ✅ Auth store with persistent session
- ✅ API client with automatic token injection & 401 handling
- ✅ TanStack Query hooks for data fetching
- ✅ Zod schemas for type safety
- ✅ Dashboard component fetching and displaying leave balance

**Next Phase: Leave Request Form & Manager Approval Portal**
