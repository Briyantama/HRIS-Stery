<?php

namespace App\Services;

use GuzzleHttp\Client;
use GuzzleHttp\Exception\GuzzleException;
use Illuminate\Http\Request;
use Illuminate\Http\Response;

/**
 * BaseGrpcGatewayClient provides HTTP/JSON wrapper around gRPC services.
 *
 * This client communicates with Go backend services via grpc-gateway,
 * which translates HTTP/JSON to gRPC and back.
 *
 * All calls include metadata headers for tracing and multi-tenancy:
 * - x-tenant-id: Extracted from validated JWT
 * - x-user-id: Extracted from validated JWT
 * - x-request-id: Generated for distributed tracing
 */
abstract class BaseGrpcGatewayClient
{
    protected Client $httpClient;
    protected string $serviceUrl;
    protected string $tenantId;
    protected string $userId;
    protected array $userRoles;
    protected string $requestId;

    /**
     * Constructor initializes HTTP client and metadata from request context.
     */
    public function __construct(Request $request)
    {
        // Initialize Guzzle HTTP client with timeout and retries
        $this->httpClient = new Client([
            'timeout' => 30,
            'connect_timeout' => 10,
        ]);

        // Extract tenant_id and user_id from request (set by middleware)
        $this->tenantId = $request->attributes->get('tenant_id', '');
        $this->userId = $request->attributes->get('user_id', '');

        // Extract user roles from JWT claims (set by JwtValidation middleware)
        $user = $request->attributes->get('user');
        $this->userRoles = $user?->roles ?? [];

        // Generate or retrieve request_id for distributed tracing
        $this->requestId = $request->header('x-request-id') ?? $this->generateRequestId();
    }

    /**
     * Make a gRPC call via HTTP/JSON gateway.
     *
     * Maps HTTP methods and paths to backend endpoints.
     * Handles error responses and translates gRPC status codes to HTTP.
     *
     * @param string $method HTTP method (GET, POST, PUT, DELETE)
     * @param string $path API path (e.g., /v1/leave-requests)
     * @param array $data Request body data (for POST/PUT)
     * @param array $queryParams Query parameters (for GET)
     * @return array Decoded JSON response from backend
     *
     * @throws \Exception gRPC errors translated to HTTP errors
     */
    protected function callGrpcService(
        string $method,
        string $path,
        array $data = [],
        array $queryParams = []
    ): array {
        try {
            $options = [
                'headers' => $this->buildHeaders(),
                'timeout' => 30,
            ];

            // Add query parameters for GET requests
            if (!empty($queryParams)) {
                $options['query'] = $queryParams;
            }

            // Add JSON body for POST/PUT requests
            if (!empty($data) && in_array($method, ['POST', 'PUT', 'PATCH'])) {
                $options['json'] = $data;
            }

            // Make HTTP request to grpc-gateway endpoint
            $response = $this->httpClient->request($method, $this->serviceUrl . $path, $options);

            // Return decoded JSON response
            return json_decode((string) $response->getBody(), true) ?? [];
        } catch (GuzzleException $e) {
            return $this->handleGrpcError($e);
        }
    }

    /**
     * Build HTTP headers for outbound gRPC calls.
     *
     * Includes metadata for tracing, multi-tenancy, and RBAC authorization.
     */
    private function buildHeaders(): array
    {
        return [
            'Content-Type' => 'application/json',
            'Accept' => 'application/json',
            'x-tenant-id' => $this->tenantId,
            'x-user-id' => $this->userId,
            'x-user-roles' => implode(',', $this->userRoles),
            'x-request-id' => $this->requestId,
            'Authorization' => $this->getBearerToken(),
        ];
    }

    /**
     * Get bearer token from request context.
     *
     * This token is passed to backend services for further validation/context.
     */
    private function getBearerToken(): string
    {
        // In a full implementation, you would extract this from the original request
        // For now, return empty; services may not require it since we validate at gateway
        return '';
    }

    /**
     * Translate gRPC errors to HTTP exceptions.
     *
     * Maps gRPC status codes to HTTP status codes.
     *
     * @throws \Exception
     */
    private function handleGrpcError(GuzzleException $e): array
    {
        // Extract gRPC error response if available
        if ($e->hasResponse()) {
            $statusCode = $e->getResponse()->getStatusCode();
            $body = json_decode((string) $e->getResponse()->getBody(), true);

            $grpcCode = $body['code'] ?? null;
            $grpcMessage = $body['message'] ?? 'Unknown error';

            // Map gRPC status codes to HTTP status codes
            $httpStatus = $this->mapGrpcToHttpStatus($grpcCode, $statusCode);

            throw new \Exception($grpcMessage, $httpStatus);
        }

        // Network error or timeout
        throw new \Exception('Service unavailable: ' . $e->getMessage(), Response::HTTP_SERVICE_UNAVAILABLE);
    }

    /**
     * Map gRPC status code to HTTP status code.
     */
    private function mapGrpcToHttpStatus(?string $grpcCode, int $httpStatus): int
    {
        // If we got an HTTP status already, respect it
        if ($httpStatus >= 400) {
            return $httpStatus;
        }

        // Map gRPC codes
        return match ($grpcCode) {
            'INVALID_ARGUMENT', 'FAILED_PRECONDITION' => Response::HTTP_BAD_REQUEST,
            'UNAUTHENTICATED' => Response::HTTP_UNAUTHORIZED,
            'PERMISSION_DENIED' => Response::HTTP_FORBIDDEN,
            'NOT_FOUND' => Response::HTTP_NOT_FOUND,
            'ALREADY_EXISTS' => Response::HTTP_CONFLICT,
            'RESOURCE_EXHAUSTED' => Response::HTTP_TOO_MANY_REQUESTS,
            'INTERNAL' => Response::HTTP_INTERNAL_SERVER_ERROR,
            'UNAVAILABLE' => Response::HTTP_SERVICE_UNAVAILABLE,
            default => Response::HTTP_INTERNAL_SERVER_ERROR,
        };
    }

    /**
     * Generate a unique request ID for distributed tracing.
     */
    private function generateRequestId(): string
    {
        return bin2hex(random_bytes(16));
    }

    /**
     * Get the service URL from environment.
     */
    protected function getServiceUrl(string $serviceEnvVar): string
    {
        $host = env($serviceEnvVar . '_HOST', 'localhost');
        $port = env($serviceEnvVar . '_PORT', '50054');
        return "http://{$host}:{$port}";
    }
}
