<?php

namespace App\Http\Controllers\Api\V1;

use App\Services\LeaveServiceClient;
use Exception;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Http\Response;
use Illuminate\Validation\ValidationException;

/**
 * LeaveController handles all leave request workflows.
 */
class LeaveController
{
    private LeaveServiceClient $leaveClient;

    public function __construct(LeaveServiceClient $leaveClient)
    {
        $this->leaveClient = $leaveClient;
    }

    public function applyLeave(Request $request): JsonResponse
    {
        try {
            $validated = $request->validate([
                'employee_id' => 'required|uuid',
                'leave_type_id' => 'required|uuid',
                'start_date' => 'required|date_format:Y-m-d',
                'end_date' => 'required|date_format:Y-m-d|after_or_equal:start_date',
                'reason' => 'nullable|string|max:1000',
                'document_key' => 'nullable|string|max:500',
            ]);

            $leaveRequest = $this->leaveClient->applyLeave($validated);

            return response()->json([
                'code' => 'SUCCESS',
                'message' => 'Leave request created successfully',
                'data' => $leaveRequest,
            ], Response::HTTP_CREATED);
        } catch (ValidationException $e) {
            return $this->validationErrorResponse($e);
        } catch (Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    public function getLeaveRequest(Request $request, string $id): JsonResponse
    {
        try {
            if (! $this->isValidUuid($id)) {
                return response()->json([
                    'code' => 'INVALID_ARGUMENT',
                    'message' => 'Invalid leave request ID format',
                ], Response::HTTP_BAD_REQUEST);
            }

            $leaveRequest = $this->leaveClient->getLeaveRequest($id);

            return response()->json([
                'code' => 'SUCCESS',
                'data' => $leaveRequest,
            ], Response::HTTP_OK);
        } catch (Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    public function listLeaveRequests(Request $request): JsonResponse
    {
        try {
            $filters = $request->validate([
                'employee_id' => 'nullable|uuid',
                'approver_id' => 'nullable|uuid',
                'status' => 'nullable|in:PENDING,APPROVED,REJECTED,CANCELLED',
                'year' => 'nullable|integer|min:2000|max:2100',
                'page_size' => 'nullable|integer|min:1|max:100',
                'page_token' => 'nullable|string',
            ]);

            $filters = array_filter($filters, fn ($value) => $value !== null);

            $result = $this->leaveClient->listLeaveRequests($filters);

            return response()->json([
                'code' => 'SUCCESS',
                'data' => $result,
            ], Response::HTTP_OK);
        } catch (ValidationException $e) {
            return $this->validationErrorResponse($e);
        } catch (Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    public function approveLeave(Request $request, string $id): JsonResponse
    {
        try {
            if (! $this->hasApprovalRole($request)) {
                return response()->json([
                    'code' => 'PERMISSION_DENIED',
                    'message' => 'Only managers and HR admins can approve leave requests',
                ], Response::HTTP_FORBIDDEN);
            }

            if (! $this->isValidUuid($id)) {
                return response()->json([
                    'code' => 'INVALID_ARGUMENT',
                    'message' => 'Invalid leave request ID format',
                ], Response::HTTP_BAD_REQUEST);
            }

            $approverId = $request->attributes->get('user_id');
            if (! $approverId) {
                return response()->json([
                    'code' => 'UNAUTHENTICATED',
                    'message' => 'User ID not found in authentication context',
                ], Response::HTTP_UNAUTHORIZED);
            }

            $leaveRequest = $this->leaveClient->approveLeave($id, $approverId);

            return response()->json([
                'code' => 'SUCCESS',
                'message' => 'Leave request approved successfully',
                'data' => $leaveRequest,
            ], Response::HTTP_OK);
        } catch (Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    public function rejectLeave(Request $request, string $id): JsonResponse
    {
        try {
            if (! $this->hasApprovalRole($request)) {
                return response()->json([
                    'code' => 'PERMISSION_DENIED',
                    'message' => 'Only managers and HR admins can reject leave requests',
                ], Response::HTTP_FORBIDDEN);
            }

            if (! $this->isValidUuid($id)) {
                return response()->json([
                    'code' => 'INVALID_ARGUMENT',
                    'message' => 'Invalid leave request ID format',
                ], Response::HTTP_BAD_REQUEST);
            }

            $validated = $request->validate([
                'reason' => 'required|string|max:1000',
            ]);

            $approverId = $request->attributes->get('user_id');
            if (! $approverId) {
                return response()->json([
                    'code' => 'UNAUTHENTICATED',
                    'message' => 'User ID not found in authentication context',
                ], Response::HTTP_UNAUTHORIZED);
            }

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
        } catch (Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    public function cancelLeave(Request $request, string $id): JsonResponse
    {
        try {
            if (! $this->isValidUuid($id)) {
                return response()->json([
                    'code' => 'INVALID_ARGUMENT',
                    'message' => 'Invalid leave request ID format',
                ], Response::HTTP_BAD_REQUEST);
            }

            $validated = $request->validate([
                'employee_id' => 'required|uuid',
            ]);

            $userId = $request->attributes->get('user_id');
            $roles = $request->attributes->get('roles', []);
            $isAdmin = in_array('hr_admin', $roles);

            if ($validated['employee_id'] !== $userId && ! $isAdmin) {
                return response()->json([
                    'code' => 'PERMISSION_DENIED',
                    'message' => 'You can only cancel your own leave requests',
                ], Response::HTTP_FORBIDDEN);
            }

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
        } catch (Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    public function getLeaveBalance(Request $request, string $employeeId): JsonResponse
    {
        try {
            if (! $this->isValidUuid($employeeId)) {
                return response()->json([
                    'code' => 'INVALID_ARGUMENT',
                    'message' => 'Invalid employee ID format',
                ], Response::HTTP_BAD_REQUEST);
            }

            $validated = $request->validate([
                'year' => 'nullable|integer|min:2000|max:2100',
            ]);

            $year = $validated['year'] ?? date('Y');

            $balance = $this->leaveClient->getLeaveBalance($employeeId, (int) $year);

            return response()->json([
                'code' => 'SUCCESS',
                'data' => $balance,
            ], Response::HTTP_OK);
        } catch (ValidationException $e) {
            return $this->validationErrorResponse($e);
        } catch (Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    public function listLeaveTypes(Request $request): JsonResponse
    {
        try {
            $leaveTypes = $this->leaveClient->listLeaveTypes();

            return response()->json([
                'code' => 'SUCCESS',
                'data' => $leaveTypes,
            ], Response::HTTP_OK);
        } catch (Exception $e) {
            return $this->grpcErrorResponse($e);
        }
    }

    private function hasApprovalRole(Request $request): bool
    {
        $roles = $request->attributes->get('roles', []);

        return in_array('manager', $roles) || in_array('hr_admin', $roles);
    }

    private function isValidUuid(string $uuid): bool
    {
        return preg_match(
            '/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i',
            $uuid
        ) === 1;
    }

    private function validationErrorResponse(ValidationException $e): JsonResponse
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

    private function grpcErrorResponse(Exception $e): JsonResponse
    {
        $statusCode = $e->getCode();

        if ($statusCode < 100 || $statusCode >= 600) {
            $statusCode = Response::HTTP_INTERNAL_SERVER_ERROR;
        }

        return response()->json([
            'code' => $this->httpStatusToGrpcCode($statusCode),
            'message' => $e->getMessage(),
        ], $statusCode);
    }

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
