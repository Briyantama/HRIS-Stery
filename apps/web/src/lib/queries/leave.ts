import { createQuery, createMutation } from '@tanstack/svelte-query';
import { derived, type Readable } from 'svelte/store';
import { apiGet, apiPost } from '$lib/api/client';
import type {
	LeaveBalance,
	LeaveType,
	LeaveRequest,
	ApplyLeaveForm,
	ListLeaveRequestsResponse
} from '$lib/schemas/leave';

/** TanStack Query hooks for leave management. Each call returns the API envelope's data. */

// Hooks must be called during component init (createQuery reads the QueryClient context); reactive
// inputs are passed as stores so the one observer updates its options instead of being recreated.
export function useLeaveBalance(employeeId: Readable<string | null>, year?: number) {
	return createQuery(
		derived(employeeId, ($id) => ({
			queryKey: ['leave-balance', $id, year],
			queryFn: () =>
				apiGet<LeaveBalance>(`/leave-balance/${encodeURIComponent($id!)}`, { query: { year } }),
			enabled: !!$id
		}))
	);
}

export function useLeaveTypes() {
	return createQuery({
		queryKey: ['leave-types'],
		queryFn: () => apiGet<LeaveType[]>('/leave-types')
	});
}

export function useLeaveRequests(
	filters: Readable<{
		employee_id?: string;
		status?: string;
		page?: number;
		per_page?: number;
	}>
) {
	return createQuery(
		derived(filters, ($f) => ({
			queryKey: ['leave-requests', $f],
			queryFn: () =>
				apiGet<ListLeaveRequestsResponse>('/leaves', {
					query: { ...$f, per_page: $f.per_page ?? 50 }
				})
		}))
	);
}

export function useLeaveRequest(leaveId: string | null) {
	return createQuery({
		queryKey: ['leave-request', leaveId],
		queryFn: () => apiGet<LeaveRequest>(`/leaves/${encodeURIComponent(leaveId!)}`),
		enabled: !!leaveId
	});
}

export function useApplyLeave() {
	return createMutation({
		// The employee is the caller (from the token); the API rejects an employee_id in the body.
		mutationFn: ({ leave_type_id, start_date, end_date, reason }: ApplyLeaveForm) =>
			apiPost<LeaveRequest>('/leaves', { leave_type_id, start_date, end_date, reason })
	});
}

export function useApproveLeave() {
	return createMutation({
		mutationFn: (leaveId: string) =>
			apiPost<LeaveRequest>(`/leaves/${encodeURIComponent(leaveId)}/approve`)
	});
}

export function useRejectLeave() {
	return createMutation({
		mutationFn: ({ leaveId, reason }: { leaveId: string; reason: string }) =>
			apiPost<LeaveRequest>(`/leaves/${encodeURIComponent(leaveId)}/reject`, { reason })
	});
}

export function useCancelLeave() {
	return createMutation({
		mutationFn: ({ leaveId }: { leaveId: string }) =>
			apiPost<LeaveRequest>(`/leaves/${encodeURIComponent(leaveId)}/cancel`)
	});
}
