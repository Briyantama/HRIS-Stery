<?php

namespace App\Services;

use GuzzleHttp\Client;
use Illuminate\Http\Response;

/**
 * AuthServiceClient communicates with the auth-service via gRPC/HTTP gateway.
 *
 * This client handles all authentication flows:
 * - User login with email/password
 * - Token refresh
 * - Token revocation
 */
class AuthServiceClient extends BaseGrpcGatewayClient
{
    /**
     * Service endpoint configuration.
     */
    protected function __construct()
    {
        parent::__construct(request());
        $this->serviceUrl = $this->getServiceUrl('AUTH_SERVICE');
    }

    /**
     * Authenticate user and return JWT tokens.
     *
     * Calls auth-service Login RPC via gRPC gateway.
     *
     * @param string $email User email
     * @param string $password User password
     * @param string $tenantSlug Tenant slug
     * @return array Containing: access_token, refresh_token, expires_in, user_id, email, tenant_id, roles
     *
     * @throws \Exception
     */
    public function login(string $email, string $password, string $tenantSlug): array
    {
        $response = $this->callGrpcService('POST', '/hris.auth.v1.AuthService/Login', [
            'email' => $email,
            'password' => $password,
            'tenant_slug' => $tenantSlug,
        ]);

        if (empty($response)) {
            throw new \Exception('Empty response from auth service');
        }

        // Map gRPC response to API response format
        return [
            'access_token' => $response['access_token'] ?? null,
            'refresh_token' => $response['refresh_token'] ?? null,
            'expires_in' => $response['expires_in'] ?? 3600,
            'user_id' => $response['user_id'] ?? null,
            'email' => $response['email'] ?? null,
            'tenant_id' => $response['tenant_id'] ?? null,
            'roles' => $response['roles'] ?? [],
        ];
    }

    /**
     * Refresh access token.
     *
     * Calls auth-service RefreshToken RPC via gRPC gateway.
     *
     * @param string $refreshToken The refresh token
     * @return array Containing: access_token, expires_in
     *
     * @throws \Exception
     */
    public function refreshToken(string $refreshToken): array
    {
        $response = $this->callGrpcService('POST', '/hris.auth.v1.AuthService/RefreshToken', [
            'refresh_token' => $refreshToken,
        ]);

        if (empty($response)) {
            throw new \Exception('Empty response from auth service');
        }

        return [
            'access_token' => $response['access_token'] ?? null,
            'expires_in' => $response['expires_in'] ?? 3600,
        ];
    }

    /**
     * Revoke token.
     *
     * Calls auth-service RevokeToken RPC via gRPC gateway.
     *
     * @param string $token The token to revoke
     * @return void
     *
     * @throws \Exception
     */
    public function revokeToken(string $token): void
    {
        $this->callGrpcService('POST', '/hris.auth.v1.AuthService/RevokeToken', [
            'token' => $token,
        ]);
    }
}
