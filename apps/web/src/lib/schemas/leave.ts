import { z } from 'zod';

/**
 * Leave API response schemas using Zod for runtime type validation.
 * These schemas ensure type safety for backend responses and frontend validation.
 */

// Enums for leave statuses
export const LeaveStatus = z.enum([
	'LEAVE_STATUS_UNSPECIFIED',
	'LEAVE_STATUS_PENDING',
	'LEAVE_STATUS_APPROVED',
	'LEAVE_STATUS_REJECTED',
	'LEAVE_STATUS_CANCELLED'
]);

export type LeaveStatus = z.infer<typeof LeaveStatus>;

// Leave type codes
export const LeaveTypeCode = z.enum([
	'LEAVE_TYPE_CODE_UNSPECIFIED',
	'LEAVE_TYPE_CODE_ANNUAL',
	'LEAVE_TYPE_CODE_SICK',
	'LEAVE_TYPE_CODE_MATERNITY',
	'LEAVE_TYPE_CODE_PATERNITY',
	'LEAVE_TYPE_CODE_BEREAVEMENT',
	'LEAVE_TYPE_CODE_UNPAID',
	'LEAVE_TYPE_CODE_OTHER'
]);

export type LeaveTypeCode = z.infer<typeof LeaveTypeCode>;

// Leave type definition
export const LeaveTypeSchema = z.object({
	id: z.string().uuid(),
	tenant_id: z.string().uuid(),
	code: z.string(),
	name: z.string(),
	max_days_per_year: z.number(),
	requires_document: z.boolean(),
	is_paid: z.boolean()
});

export type LeaveType = z.infer<typeof LeaveTypeSchema>;

// Single leave type balance breakdown
export const LeaveTypeBalanceSchema = z.object({
	leave_type_id: z.string().uuid(),
	leave_type_code: z.string(),
	leave_type_name: z.string(),
	entitled_days: z.number(),
	used_days: z.number(),
	pending_days: z.number(),
	remaining_days: z.number()
});

export type LeaveTypeBalance = z.infer<typeof LeaveTypeBalanceSchema>;

// Overall leave balance for an employee
export const LeaveBalanceSchema = z.object({
	tenant_id: z.string().uuid(),
	employee_id: z.string().uuid(),
	year: z.number(),
	balances: z.array(LeaveTypeBalanceSchema)
});

export type LeaveBalance = z.infer<typeof LeaveBalanceSchema>;

// Single leave request
export const LeaveRequestSchema = z.object({
	id: z.string().uuid(),
	tenant_id: z.string().uuid(),
	employee_id: z.string().uuid(),
	leave_type_id: z.string().uuid(),
	status: z.string(),
	start_date: z.string(),
	end_date: z.string(),
	days_count: z.number(),
	reason: z.string(),
	rejection_reason: z.string().optional().nullable(),
	approved_by_id: z.string().optional().nullable(),
	approved_at: z.string().optional().nullable(),
	document_key: z.string().optional().nullable(),
	created_at: z.string(),
	updated_at: z.string()
});

export type LeaveRequest = z.infer<typeof LeaveRequestSchema>;

// List leave requests response
export const ListLeaveRequestsResponseSchema = z.object({
	leave_requests: z.array(LeaveRequestSchema),
	total_count: z.number(),
	next_page_token: z.string().nullable()
});

export type ListLeaveRequestsResponse = z.infer<typeof ListLeaveRequestsResponseSchema>;

// Apply leave form validation
export const ApplyLeaveFormSchema = z.object({
	employee_id: z.string().uuid('Invalid employee ID'),
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

// API error response
export const ApiErrorSchema = z.object({
	code: z.string(),
	message: z.string(),
	details: z.array(z.string()).optional()
});

export type ApiError = z.infer<typeof ApiErrorSchema>;

// Generic API response wrapper
export const ApiResponseSchema = z.object({
	code: z.string(),
	message: z.string().optional(),
	data: z.unknown().optional()
});

export type ApiResponse = z.infer<typeof ApiResponseSchema>;
