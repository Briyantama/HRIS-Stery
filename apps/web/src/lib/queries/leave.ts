import { createQuery, createMutation } from '@tanstack/svelte-query';
import { apiGet, apiPost } from '$lib/api/client';
import type { LeaveBalance, LeaveType, LeaveRequest, ApplyLeaveForm } from '$lib/schemas/leave';

/**
 * TanStack Query hooks for leave management.
 * Used for server state management and caching.
 */

/**
 * Hook to fetch employee leave balance
 */
export function useLeaveBalance(employeeId: string | null, year?: number) {
	return createQuery({
		queryKey: ['leave-balance', employeeId, year],
		queryFn: async () => {
			if (!employeeId) {
				throw new Error('Employee ID required');
			}

			const response = await apiGet<LeaveBalance>(`/leave-balance/${employeeId}`, {
				query: year ? { year: year.toString() } : {}
			});

			if (!response.data) {
				throw new Error('No balance data returned');
			}

			return response.data;
		},
		enabled: !!employeeId
	});
}

/**
 * Hook to fetch all leave types for tenant
 */
export function useLeaveTypes() {
	return createQuery({
		queryKey: ['leave-types'],
		queryFn: async () => {
			const response = await apiGet<LeaveType[]>('/leave-types');

			if (!response.data) {
				throw new Error('No leave types returned');
			}

			return response.data;
		}
	});
}

/**
 * Hook to fetch leave requests list (with optional filters)
 */
export function useLeaveRequests(filters?: {
	employee_id?: string;
	approver_id?: string;
	status?: string;
	year?: number;
	page_size?: number;
}) {
	return createQuery({
		queryKey: ['leave-requests', filters],
		queryFn: async () => {
			const response = await apiGet<{
				leave_requests: LeaveRequest[];
				total_count: number;
				next_page_token: string | null;
			}>('/leaves', {
				query: filters
					? {
							...filters,
							page_size: filters.page_size || '50'
						}
					: {}
			});

			if (!response.data) {
				throw new Error('No leave requests returned');
			}

			return response.data;
		}
	});
}

/**
 * Hook to fetch single leave request by ID
 */
export function useLeaveRequest(leaveId: string | null) {
	return createQuery({
		queryKey: ['leave-request', leaveId],
		queryFn: async () => {
			if (!leaveId) {
				throw new Error('Leave ID required');
			}

			const response = await apiGet<LeaveRequest>(`/leaves/${leaveId}`);

			if (!response.data) {
				throw new Error('No leave request returned');
			}

			return response.data;
		},
		enabled: !!leaveId
	});
}

/**
 * Hook to create (apply for) a new leave request
 */
export function useApplyLeave() {
	return createMutation({
		mutationFn: async (form: ApplyLeaveForm) => {
			const response = await apiPost<LeaveRequest>('/leaves', form);

			if (!response.data) {
				throw new Error('No leave request returned');
			}

			return response.data;
		}
	});
}

/**
 * Hook to approve a leave request
 */
export function useApproveLeave() {
	return createMutation({
		mutationFn: async (leaveId: string) => {
			const response = await apiPost<LeaveRequest>(`/leaves/${leaveId}/approve`);

			if (!response.data) {
				throw new Error('No leave request returned');
			}

			return response.data;
		}
	});
}

/**
 * Hook to reject a leave request
 */
export function useRejectLeave() {
	return createMutation({
		mutationFn: async ({
			leaveId,
			reason
		}: {
			leaveId: string;
			reason: string;
		}) => {
			const response = await apiPost<LeaveRequest>(`/leaves/${leaveId}/reject`, {
				reason
			});

			if (!response.data) {
				throw new Error('No leave request returned');
			}

			return response.data;
		}
	});
}

/**
 * Hook to cancel a leave request
 */
export function useCancelLeave() {
	return createMutation({
		mutationFn: async ({
			leaveId,
			employeeId
		}: {
			leaveId: string;
			employeeId: string;
		}) => {
			const response = await apiPost<LeaveRequest>(`/leaves/${leaveId}/cancel`, {
				employee_id: employeeId
			});

			if (!response.data) {
				throw new Error('No leave request returned');
			}

			return response.data;
		}
	});
}
