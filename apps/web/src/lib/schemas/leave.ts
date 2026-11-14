import { z } from 'zod';

/**
 * Leave API response schemas using Zod for runtime type validation.
 * These schemas ensure type safety for backend responses and frontend validation.
 */

// Enums for leave statuses
// The strings the Go leave handler returns.
export const LeaveStatus = z.enum(['PENDING', 'APPROVED', 'REJECTED', 'CANCELLED']);

export type LeaveStatus = z.infer<typeof LeaveStatus>;

// Backend response shapes (plain types, the server is trusted).
export interface LeaveType {
	id: string;
	code: string;
	name: string;
	max_days_per_year: number;
	requires_document: boolean;
	is_paid: boolean;
	is_active: boolean;
}

export interface LeaveTypeBalance {
	leave_type_id: string;
	entitled_days: number;
	used_days: number;
	pending_days: number;
	remaining_days: number;
}

export interface LeaveBalance {
	tenant_id: string;
	employee_id: string;
	year: number;
	by_type: LeaveTypeBalance[];
}

export interface LeaveRequest {
	id: string;
	tenant_id: string;
	employee_id: string;
	leave_type_id: string;
	status: LeaveStatus;
	start_date: string;
	end_date: string;
	days_count: number;
	reason: string;
	rejection_reason?: string | null;
	approved_by_id?: string | null;
	approved_at?: string | null;
	document_key?: string | null;
	created_at: string;
	updated_at: string;
}

export interface ListLeaveRequestsResponse {
	items: LeaveRequest[];
	page: number;
	per_page: number;
	total: number;
}

// Apply leave form validation
export const ApplyLeaveFormSchema = z.object({
	leave_type_id: z.string().uuid('Please select a leave type'),
	start_date: z.string().refine((date) => /^\d{4}-\d{2}-\d{2}$/.test(date), 'Invalid date format'),
	end_date: z.string().refine((date) => /^\d{4}-\d{2}-\d{2}$/.test(date), 'Invalid date format'),
	reason: z.string().max(1000, 'Reason cannot exceed 1000 characters').optional()
});

export type ApplyLeaveForm = z.infer<typeof ApplyLeaveFormSchema>;

// Refined validation: end_date must be >= start_date
export const ApplyLeaveRequestSchema = ApplyLeaveFormSchema.refine(
	(data) => new Date(data.end_date) >= new Date(data.start_date),
	{
		message: 'End date must be on or after start date',
		path: ['end_date']
	}
);

// Reject leave form
export const RejectLeaveFormSchema = z.object({
	reason: z.string().min(1, 'Reason is required').max(1000, 'Reason cannot exceed 1000 characters')
});

export type RejectLeaveForm = z.infer<typeof RejectLeaveFormSchema>;
