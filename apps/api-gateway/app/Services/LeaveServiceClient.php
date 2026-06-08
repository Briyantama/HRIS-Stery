<?php

namespace App\Services;

use Illuminate\Http\Request;

/**
 * LeaveServiceClient provides access to the Go leave-service via gRPC-gateway.
 *
 * Maps leave operations to HTTP endpoints:
 * - POST   /v1/leave-requests         → ApplyLeave
 * - GET    /v1/leave-requests/{id}    → GetLeaveRequest
 * - GET    /v1/leave-requests         → ListLeaveRequests
 * - POST   /v1/leave-requests/{id}/approve → ApproveLeave
 * - POST   /v1/leave-requests/{id}/reject  → RejectLeave
 * - POST   /v1/leave-requests/{id}/cancel  → CancelLeave
 * - GET    /v1/leave-balance/{employee_id} → GetLeaveBalance
 * - GET    /v1/leave-types            → ListLeaveTypes
 */
class LeaveServiceClient extends BaseGrpcGatewayClient
{
    public function __construct(Request $request)
    {
        parent::__construct($request);
        // Configure service endpoint from environment
        $this->serviceUrl = $this->getServiceUrl('LEAVE_SERVICE');
    }

    /**
     * Apply for a new leave request.
     *
     * @param array $leaveData {
     *     "employee_id": "uuid",
     *     "leave_type_id": "uuid",
     *     "start_date": "YYYY-MM-DD",
     *     "end_date": "YYYY-MM-DD",
     *     "reason": "string",
     *     "document_key": "string|optional"
     * }
     *
     * @return array Leave request object with id, status, created_at, etc.
     */
    public function applyLeave(array $leaveData): array
    {
        $payload = [
            'tenant_id' => $this->tenantId,
            'employee_id' => $leaveData['employee_id'] ?? null,
            'leave_type_id' => $leaveData['leave_type_id'] ?? null,
            'start_date' => $leaveData['start_date'] ?? null,
            'end_date' => $leaveData['end_date'] ?? null,
            'reason' => $leaveData['reason'] ?? null,
            'document_key' => $leaveData['document_key'] ?? null,
        ];

        $response = $this->callGrpcService('POST', '/v1/leave-requests', $payload);

        return $response['leave_request'] ?? $response;
    }

    /**
     * Get a single leave request by ID.
     *
     * @param string $leaveRequestId UUID of the leave request
     * @return array Leave request object
     */
    public function getLeaveRequest(string $leaveRequestId): array
    {
        $response = $this->callGrpcService('GET', "/v1/leave-requests/{$leaveRequestId}", [], [
            'tenant_id' => $this->tenantId,
        ]);

        return $response['leave_request'] ?? $response;
    }

    /**
     * List leave requests with optional filters.
     *
     * @param array $filters {
     *     "employee_id": "uuid|optional",
     *     "approver_id": "uuid|optional",
     *     "status": "enum|optional",
     *     "year": "int|optional",
     *     "page_size": "int|optional",
     *     "page_token": "string|optional"
     * }
     *
     * @return array { leave_requests: [...], total_count: int, next_page_token: string }
     */
    public function listLeaveRequests(array $filters = []): array
    {
        $queryParams = [
            'tenant_id' => $this->tenantId,
        ];

        // Add optional filters
        if (isset($filters['employee_id'])) {
            $queryParams['employee_id'] = $filters['employee_id'];
        }
        if (isset($filters['approver_id'])) {
            $queryParams['approver_id'] = $filters['approver_id'];
        }
        if (isset($filters['status'])) {
            $queryParams['status'] = $filters['status'];
        }
        if (isset($filters['year'])) {
            $queryParams['year'] = $filters['year'];
        }
        if (isset($filters['page_size'])) {
            $queryParams['page_size'] = $filters['page_size'];
        }
        if (isset($filters['page_token'])) {
            $queryParams['page_token'] = $filters['page_token'];
        }

        return $this->callGrpcService('GET', '/v1/leave-requests', [], $queryParams);
    }

    /**
     * Approve a pending leave request.
     *
     * @param string $leaveRequestId UUID of the leave request to approve
     * @param string $approverId UUID of the approver (manager/hr_admin)
     * @return array Updated leave request object
     */
    public function approveLeave(string $leaveRequestId, string $approverId): array
    {
        $payload = [
            'id' => $leaveRequestId,
            'tenant_id' => $this->tenantId,
            'approver_id' => $approverId,
        ];

        $response = $this->callGrpcService('POST', "/v1/leave-requests/{$leaveRequestId}/approve", $payload);

        return $response['leave_request'] ?? $response;
    }

    /**
     * Reject a pending leave request.
     *
     * @param string $leaveRequestId UUID of the leave request to reject
     * @param string $approverId UUID of the approver (manager/hr_admin)
     * @param string $rejectionReason Reason for rejection
     * @return array Updated leave request object
     */
    public function rejectLeave(string $leaveRequestId, string $approverId, string $rejectionReason): array
    {
        $payload = [
            'id' => $leaveRequestId,
            'tenant_id' => $this->tenantId,
            'approver_id' => $approverId,
            'rejection_reason' => $rejectionReason,
        ];

        $response = $this->callGrpcService('POST', "/v1/leave-requests/{$leaveRequestId}/reject", $payload);

        return $response['leave_request'] ?? $response;
    }

    /**
     * Cancel a leave request (pending or approved).
     *
     * @param string $leaveRequestId UUID of the leave request to cancel
     * @param string $employeeId UUID of the employee (must be the requester)
     * @return array Updated leave request object
     */
    public function cancelLeave(string $leaveRequestId, string $employeeId): array
    {
        $payload = [
            'id' => $leaveRequestId,
            'tenant_id' => $this->tenantId,
            'employee_id' => $employeeId,
        ];

        $response = $this->callGrpcService('POST', "/v1/leave-requests/{$leaveRequestId}/cancel", $payload);

        return $response['leave_request'] ?? $response;
    }

    /**
     * Get leave balance for an employee.
     *
     * @param string $employeeId UUID of the employee
     * @param int $year Calendar year (defaults to current year)
     * @return array Leave balance object with per-type breakdowns
     */
    public function getLeaveBalance(string $employeeId, int $year = 0): array
    {
        $queryParams = [
            'tenant_id' => $this->tenantId,
            'employee_id' => $employeeId,
        ];

        if ($year > 0) {
            $queryParams['year'] = $year;
        }

        $response = $this->callGrpcService('GET', "/v1/leave-balance/{$employeeId}", [], $queryParams);

        return $response['balance'] ?? $response;
    }

    /**
     * List all leave types configured for the tenant.
     *
     * @return array Array of leave type objects
     */
    public function listLeaveTypes(): array
    {
        $queryParams = [
            'tenant_id' => $this->tenantId,
        ];

        $response = $this->callGrpcService('GET', '/v1/leave-types', [], $queryParams);

        return $response['leave_types'] ?? $response;
    }
}
