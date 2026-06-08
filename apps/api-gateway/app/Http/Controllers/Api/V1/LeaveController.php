<?php

namespace App\Http\Controllers\Api\V1;

use App\Services\LeaveServiceClient;
use Illuminate\Http\Request;
use Illuminate\Http\Response;
use Illuminate\Validation\ValidationException;

/**
 * LeaveController handles all leave request workflows.
 *
 * Routes:
 * - POST   /api/v1/leaves           → applyLeave
 * - GET    /api/v1/leaves           → listLeaveRequests
 * - GET    /api/v1/leaves/{id}      → getLeaveRequest
 * - POST   /api/v1/leaves/{id}/approve  → approveLeave
 * - POST   /api/v1/leaves/{id}/reject   → rejectLeave
 * - POST   /api/v1/leaves/{id}/cancel   → cancelLeave
 * - GET    /api/v1/leave-balance/{employeeId} → getLeaveBalance
 * - GET    /api/v1/leave-types     → listLeaveTypes
 */
class LeaveController
{
    private LeaveServiceClient $leaveClient;

    public function __construct(LeaveServiceClient $leaveClient)
    {
        $this->leaveClient = $leaveClient;
    }

    /**
     * Apply for a new leave request.
     *
     * POST /api/v1/leaves
     *
     * @param Request $request
     * @return \Illuminate\Http\JsonResponse
     */
    public function applyLeave(Request $request)
    {
        try {
            // Validate incoming request
            $validated = $request->validate([
                'employee_id' => 'required|uuid',
                'leave_type_id' => 'required|uuid',
                'start_date' => 'required|date_format:Y-m-d',
                'end_date' => 'required|date_format:Y-m-d|after_or_equal:start_date',
                'reason' => 'nullable|string|max:1000',
                'document_key' => 'nullable|string|max:500',
            ]);

            // Call gRPC service
            $leaveRequest = $this->leaveClient->applyLeave($validated);

            return response()->json([
                'code' => 'SUCCESS',
                'message' => 'Leave request created successfully',
                'data' => $leaveRequest,
            ], Response::HTTP_CREATED);
        } catch (ValidationException $e) {
            return $this->validationErrorResponse($e);
        } catch (\Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    /**
     * Get a specific leave request by ID.
     *
     * GET /api/v1/leaves/{id}
     *
     * @param Request $request
     * @param string $id Leave request ID
     * @return \Illuminate\Http\JsonResponse
     */
    public function getLeaveRequest(Request $request, string $id)
    {
        try {
            // Validate ID is a UUID
            if (!$this->isValidUuid($id)) {
                return response()->json([
                    'code' => 'INVALID_ARGUMENT',
                    'message' => 'Invalid leave request ID format',
                ], Response::HTTP_BAD_REQUEST);
            }

            // Call gRPC service
            $leaveRequest = $this->leaveClient->getLeaveRequest($id);

            return response()->json([
                'code' => 'SUCCESS',
                'data' => $leaveRequest,
            ], Response::HTTP_OK);
        } catch (\Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    /**
     * List leave requests with optional filters.
     *
     * GET /api/v1/leaves?employee_id=uuid&status=APPROVED&year=2026
     *
     * @param Request $request
     * @return \Illuminate\Http\JsonResponse
     */
    public function listLeaveRequests(Request $request)
    {
        try {
            // Validate query parameters
            $filters = $request->validate([
                'employee_id' => 'nullable|uuid',
                'approver_id' => 'nullable|uuid',
                'status' => 'nullable|in:PENDING,APPROVED,REJECTED,CANCELLED',
                'year' => 'nullable|integer|min:2000|max:2100',
                'page_size' => 'nullable|integer|min:1|max:100',
                'page_token' => 'nullable|string',
            ]);

            // Remove null values
            $filters = array_filter($filters, fn($value) => $value !== null);

            // Call gRPC service
            $result = $this->leaveClient->listLeaveRequests($filters);

            return response()->json([
                'code' => 'SUCCESS',
                'data' => $result,
            ], Response::HTTP_OK);
        } catch (ValidationException $e) {
            return $this->validationErrorResponse($e);
        } catch (\Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    /**
     * Approve a pending leave request.
     *
     * POST /api/v1/leaves/{id}/approve
     *
     * Requires: manager or hr_admin role
     *
     * @param Request $request
     * @param string $id Leave request ID
     * @return \Illuminate\Http\JsonResponse
     */
    public function approveLeave(Request $request, string $id)
    {
        try {
            // RBAC: Check if user has manager or hr_admin role
            if (!$this->hasApprovalRole($request)) {
                return response()->json([
                    'code' => 'PERMISSION_DENIED',
                    'message' => 'Only managers and HR admins can approve leave requests',
                ], Response::HTTP_FORBIDDEN);
            }

            // Validate ID is a UUID
            if (!$this->isValidUuid($id)) {
                return response()->json([
                    'code' => 'INVALID_ARGUMENT',
                    'message' => 'Invalid leave request ID format',
                ], Response::HTTP_BAD_REQUEST);
            }

            // Get authenticated user ID
            $approverId = $request->attributes->get('user_id');
            if (!$approverId) {
                return response()->json([
                    'code' => 'UNAUTHENTICATED',
                    'message' => 'User ID not found in authentication context',
                ], Response::HTTP_UNAUTHORIZED);
            }

            // Call gRPC service
            $leaveRequest = $this->leaveClient->approveLeave($id, $approverId);

            return response()->json([
                'code' => 'SUCCESS',
                'message' => 'Leave request approved successfully',
                'data' => $leaveRequest,
            ], Response::HTTP_OK);
        } catch (\Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    /**
     * Reject a pending leave request.
     *
     * POST /api/v1/leaves/{id}/reject
     *
     * Requires: manager or hr_admin role
     * Body: { "reason": "Reason for rejection" }
     *
     * @param Request $request
     * @param string $id Leave request ID
     * @return \Illuminate\Http\JsonResponse
     */
    public function rejectLeave(Request $request, string $id)
    {
        try {
            // RBAC: Check if user has manager or hr_admin role
            if (!$this->hasApprovalRole($request)) {
                return response()->json([
                    'code' => 'PERMISSION_DENIED',
                    'message' => 'Only managers and HR admins can reject leave requests',
                ], Response::HTTP_FORBIDDEN);
            }

            // Validate ID is a UUID
            if (!$this->isValidUuid($id)) {
                return response()->json([
                    'code' => 'INVALID_ARGUMENT',
                    'message' => 'Invalid leave request ID format',
                ], Response::HTTP_BAD_REQUEST);
            }

            // Validate rejection reason
            $validated = $request->validate([
                'reason' => 'required|string|max:1000',
            ]);

            // Get authenticated user ID
            $approverId = $request->attributes->get('user_id');
            if (!$approverId) {
                return response()->json([
                    'code' => 'UNAUTHENTICATED',
                    'message' => 'User ID not found in authentication context',
                ], Response::HTTP_UNAUTHORIZED);
            }

            // Call gRPC service
            $leaveRequest = $this->leaveClient->rejectLeave(
                $id,
                $approverId,
                $validated['reason']
            );

            return response()->json([
                'code' => 'SUCCESS',
                'message' => 'Leave request rejected successfully',
                'data' => $leaveRequest,
            ], Response::HTTP_OK);
        } catch (ValidationException $e) {
            return $this->validationErrorResponse($e);
        } catch (\Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    /**
     * Cancel a leave request.
     *
     * POST /api/v1/leaves/{id}/cancel
     *
     * Employees can cancel their own leave requests.
     * HR admins can cancel any leave request.
     *
     * @param Request $request
     * @param string $id Leave request ID
     * @return \Illuminate\Http\JsonResponse
     */
    public function cancelLeave(Request $request, string $id)
    {
        try {
            // Validate ID is a UUID
            if (!$this->isValidUuid($id)) {
                return response()->json([
                    'code' => 'INVALID_ARGUMENT',
                    'message' => 'Invalid leave request ID format',
                ], Response::HTTP_BAD_REQUEST);
            }

            // Validate employee_id is provided
            $validated = $request->validate([
                'employee_id' => 'required|uuid',
            ]);

            // RBAC: Check if user is canceling their own leave or is HR admin
            $userId = $request->attributes->get('user_id');
            $roles = $request->attributes->get('roles', []);
            $isAdmin = in_array('hr_admin', $roles);

            if ($validated['employee_id'] !== $userId && !$isAdmin) {
                return response()->json([
                    'code' => 'PERMISSION_DENIED',
                    'message' => 'You can only cancel your own leave requests',
                ], Response::HTTP_FORBIDDEN);
            }

            // Call gRPC service
            $leaveRequest = $this->leaveClient->cancelLeave(
                $id,
                $validated['employee_id']
            );

            return response()->json([
                'code' => 'SUCCESS',
                'message' => 'Leave request cancelled successfully',
                'data' => $leaveRequest,
            ], Response::HTTP_OK);
        } catch (ValidationException $e) {
            return $this->validationErrorResponse($e);
        } catch (\Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    /**
     * Get leave balance for an employee.
     *
     * GET /api/v1/leave-balance/{employeeId}?year=2026
     *
     * @param Request $request
     * @param string $employeeId Employee ID
     * @return \Illuminate\Http\JsonResponse
     */
    public function getLeaveBalance(Request $request, string $employeeId)
    {
        try {
            // Validate employee ID is a UUID
            if (!$this->isValidUuid($employeeId)) {
                return response()->json([
                    'code' => 'INVALID_ARGUMENT',
                    'message' => 'Invalid employee ID format',
                ], Response::HTTP_BAD_REQUEST);
            }

            // Validate optional year parameter
            $validated = $request->validate([
                'year' => 'nullable|integer|min:2000|max:2100',
            ]);

            $year = $validated['year'] ?? date('Y');

            // Call gRPC service
            $balance = $this->leaveClient->getLeaveBalance($employeeId, (int) $year);

            return response()->json([
                'code' => 'SUCCESS',
                'data' => $balance,
            ], Response::HTTP_OK);
        } catch (ValidationException $e) {
            return $this->validationErrorResponse($e);
        } catch (\Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    /**
     * List all leave types configured for the tenant.
     *
     * GET /api/v1/leave-types
     *
     * @param Request $request
     * @return \Illuminate\Http\JsonResponse
     */
    public function listLeaveTypes(Request $request)
    {
        try {
            // Call gRPC service
            $leaveTypes = $this->leaveClient->listLeaveTypes();

            return response()->json([
                'code' => 'SUCCESS',
                'data' => $leaveTypes,
            ], Response::HTTP_OK);
        } catch (\Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    /**
     * Check if user has role required to approve/reject leave.
     *
     * @param Request $request
     * @return bool
     */
    private function hasApprovalRole(Request $request): bool
    {
        $roles = $request->attributes->get('roles', []);
        return in_array('manager', $roles) || in_array('hr_admin', $roles);
    }

    /**
     * Validate UUID format.
     *
     * @param string $uuid
     * @return bool
     */
    private function isValidUuid(string $uuid): bool
    {
        return preg_match(
            '/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i',
            $uuid
        ) === 1;
    }

    /**
     * Format validation error response.
     *
     * @param ValidationException $e
     * @return \Illuminate\Http\JsonResponse
     */
    private function validationErrorResponse(ValidationException $e): \Illuminate\Http\JsonResponse
    {
        $errors = [];
        foreach ($e->validator->errors()->all() as $error) {
            $errors[] = $error;
        }

        return response()->json([
            'code' => 'INVALID_ARGUMENT',
            'message' => 'Validation failed',
            'details' => $errors,
        ], Response::HTTP_BAD_REQUEST);
    }

    /**
     * Format gRPC error response.
     *
     * Translates gRPC error messages to HTTP responses.
     *
     * @param \Exception $e
     * @return \Illuminate\Http\JsonResponse
     */
    private function grpcErrorResponse(\Exception $e): \Illuminate\Http\JsonResponse
    {
        $statusCode = $e->getCode();

        // Default to 500 if code not a valid HTTP status
        if ($statusCode < 100 || $statusCode >= 600) {
            $statusCode = Response::HTTP_INTERNAL_SERVER_ERROR;
        }

        return response()->json([
            'code' => $this->httpStatusToGrpcCode($statusCode),
            'message' => $e->getMessage(),
        ], $statusCode);
    }

    /**
     * Map HTTP status code to gRPC status code.
     *
     * @param int $statusCode
     * @return string
     */
    private function httpStatusToGrpcCode(int $statusCode): string
    {
        return match ($statusCode) {
            Response::HTTP_BAD_REQUEST => 'INVALID_ARGUMENT',
            Response::HTTP_UNAUTHORIZED => 'UNAUTHENTICATED',
            Response::HTTP_FORBIDDEN => 'PERMISSION_DENIED',
            Response::HTTP_NOT_FOUND => 'NOT_FOUND',
            Response::HTTP_CONFLICT => 'ALREADY_EXISTS',
            Response::HTTP_INTERNAL_SERVER_ERROR => 'INTERNAL',
            Response::HTTP_SERVICE_UNAVAILABLE => 'UNAVAILABLE',
            default => 'INTERNAL',
        };
    }
}
