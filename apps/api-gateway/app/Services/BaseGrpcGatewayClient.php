<?php

namespace App\Services;

use Exception;
use GuzzleHttp\Client;
use GuzzleHttp\Exception\GuzzleException;
use Illuminate\Http\Request;
use Illuminate\Http\Response;

/**
 * BaseGrpcGatewayClient provides HTTP/JSON wrapper around gRPC services.
 */
abstract class BaseGrpcGatewayClient
{
    protected Client $httpClient;

    protected string $serviceUrl;

    protected string $tenantId;

    protected string $userId;

    protected array $userRoles;

    protected string $requestId;

    public function __construct(Request $request)
    {
        $this->httpClient = new Client([
            'timeout' => 30,
            'connect_timeout' => 10,
        ]);

        $this->tenantId = $request->attributes->get('tenant_id', '');
        $this->userId = $request->attributes->get('user_id', '');

        $user = $request->attributes->get('user');
        $this->userRoles = $user?->roles ?? [];

        $this->requestId = $request->header('x-request-id') ?? $this->generateRequestId();
    }

    /**
     * @throws Exception
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

            if (! empty($queryParams)) {
                $options['query'] = $queryParams;
            }

            if (! empty($data) && in_array($method, ['POST', 'PUT', 'PATCH'])) {
                $options['json'] = $data;
            }

            $response = $this->httpClient->request($method, $this->serviceUrl.$path, $options);

            return json_decode((string) $response->getBody(), true) ?? [];
        } catch (GuzzleException $e) {
            return $this->handleGrpcError($e);
        }
    }

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

    private function getBearerToken(): string
    {
        return '';
    }

    /**
     * @throws Exception
     */
    private function handleGrpcError(GuzzleException $e): array
    {
        if ($e->hasResponse()) {
            $statusCode = $e->getResponse()->getStatusCode();
            $body = json_decode((string) $e->getResponse()->getBody(), true);

            $grpcCode = $body['code'] ?? null;
            $grpcMessage = $body['message'] ?? 'Unknown error';

            $httpStatus = $this->mapGrpcToHttpStatus($grpcCode, $statusCode);

            throw new Exception($grpcMessage, $httpStatus);
        }

        throw new Exception('Service unavailable: '.$e->getMessage(), Response::HTTP_SERVICE_UNAVAILABLE);
    }

    private function mapGrpcToHttpStatus(?string $grpcCode, int $httpStatus): int
    {
        if ($httpStatus >= 400) {
            return $httpStatus;
        }

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

    private function generateRequestId(): string
    {
        return bin2hex(random_bytes(16));
    }

    protected function getServiceUrl(string $serviceEnvVar): string
    {
        $host = env($serviceEnvVar.'_HOST', 'localhost');
        $port = env($serviceEnvVar.'_PORT', '50054');

        return "http://{$host}:{$port}";
    }
}
