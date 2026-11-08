import { createQuery, createMutation } from '@tanstack/svelte-query';
import { apiGet, apiPost } from '$lib/api/client';
import type {
	LeaveBalance,
	LeaveType,
	LeaveRequest,
	ApplyLeaveForm,
	ListLeaveRequestsResponse
} from '$lib/schemas/leave';

/** TanStack Query hooks for leave management. Each call returns the API envelope's data. */

export function useLeaveBalance(employeeId: string | null, year?: number) {
	return createQuery({
		queryKey: ['leave-balance', employeeId, year],
		queryFn: () =>
			apiGet<LeaveBalance>(`/leave-balance/${encodeURIComponent(employeeId!)}`, { query: { year } }),
		enabled: !!employeeId
	});
}

export function useLeaveTypes() {
	return createQuery({
		queryKey: ['leave-types'],
		queryFn: () => apiGet<LeaveType[]>('/leave-types')
	});
}

export function useLeaveRequests(filters?: {
	employee_id?: string;
	status?: string;
	page?: number;
	per_page?: number;
}) {
	return createQuery({
		queryKey: ['leave-requests', filters],
		queryFn: () =>
			apiGet<ListLeaveRequestsResponse>('/leaves', {
				query: { ...filters, per_page: filters?.per_page ?? 50 }
			})
	});
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
