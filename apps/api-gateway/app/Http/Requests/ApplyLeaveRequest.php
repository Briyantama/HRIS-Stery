<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

/**
 * ApplyLeaveRequest validates incoming leave application requests.
 *
 * This form request class can be used as an alternative to inline
 * validation in the controller for more sophisticated validation logic.
 *
 * Usage in controller:
 *     public function applyLeave(ApplyLeaveRequest $request)
 *     {
 *         $validated = $request->validated();
 *         // $validated already contains sanitized data
 *     }
 */
class ApplyLeaveRequest extends FormRequest
{
    /**
     * Determine if the user is authorized to make this request.
     *
     * All authenticated users (with valid JWT) can apply for leave.
     * The leave service will enforce employee-specific constraints.
     */
    public function authorize(): bool
    {
        // JWT validation middleware ensures user is authenticated
        return true;
    }

    /**
     * Get the validation rules that apply to the request.
     */
    public function rules(): array
    {
        return [
            'employee_id' => 'required|uuid|string',
            'leave_type_id' => 'required|uuid|string',
            'start_date' => 'required|date_format:Y-m-d',
            'end_date' => 'required|date_format:Y-m-d|after_or_equal:start_date',
            'reason' => 'nullable|string|max:1000',
            'document_key' => 'nullable|string|max:500',
        ];
    }

    /**
     * Get custom messages for validator errors.
     */
    public function messages(): array
    {
        return [
            'employee_id.required' => 'Employee ID is required',
            'employee_id.uuid' => 'Employee ID must be a valid UUID',
            'leave_type_id.required' => 'Leave type ID is required',
            'leave_type_id.uuid' => 'Leave type ID must be a valid UUID',
            'start_date.required' => 'Start date is required',
            'start_date.date_format' => 'Start date must be in YYYY-MM-DD format',
            'end_date.required' => 'End date is required',
            'end_date.date_format' => 'End date must be in YYYY-MM-DD format',
            'end_date.after_or_equal' => 'End date must be on or after start date',
            'reason.max' => 'Reason cannot exceed 1000 characters',
            'document_key.max' => 'Document key cannot exceed 500 characters',
        ];
    }
}
