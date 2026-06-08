<?php

namespace App\Services;

use Illuminate\Http\Request;

/**
 * LeaveServiceClient provides access to the Go leave-service via gRPC-gateway.
 */
class LeaveServiceClient extends BaseGrpcGatewayClient
{
    public function __construct(Request $request)
    {
        parent::__construct($request);
        $this->serviceUrl = $this->getServiceUrl('LEAVE_SERVICE');
    }

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

    public function getLeaveRequest(string $leaveRequestId): array
    {
        $response = $this->callGrpcService('GET', "/v1/leave-requests/{$leaveRequestId}", [], [
            'tenant_id' => $this->tenantId,
        ]);

        return $response['leave_request'] ?? $response;
    }

    public function listLeaveRequests(array $filters = []): array
    {
        $queryParams = [
            'tenant_id' => $this->tenantId,
        ];

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

    public function listLeaveTypes(): array
    {
        $queryParams = [
            'tenant_id' => $this->tenantId,
        ];

        $response = $this->callGrpcService('GET', '/v1/leave-types', [], $queryParams);

        return $response['leave_types'] ?? $response;
    }
}
