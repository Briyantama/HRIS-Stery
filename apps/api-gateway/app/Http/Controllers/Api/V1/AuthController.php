<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Requests\LoginRequest;
use App\Services\AuthServiceClient;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

/**
 * AuthController handles authentication endpoints.
 *
 * This controller proxies auth requests to the auth-service via gRPC.
 * It manages JWT token issuance, refresh, and revocation.
 */
class AuthController
{
    protected AuthServiceClient $authService;

    /**
     * Inject auth service client.
     */
    public function __construct(AuthServiceClient $authService)
    {
        $this->authService = $authService;
    }

    /**
     * Login endpoint: authenticates user and returns JWT tokens.
     *
     * POST /api/v1/auth/login
     *
     * @param LoginRequest $request
     * @return JsonResponse
     */
    public function login(LoginRequest $request): JsonResponse
    {
        try {
            // Call auth-service to validate credentials and get tokens
            $response = $this->authService->login(
                email: $request->input('email'),
                password: $request->input('password'),
                tenantSlug: $request->input('tenant_slug')
            );

            // Return tokens and user info
            return response()->json([
                'success' => true,
                'access_token' => $response['access_token'],
                'refresh_token' => $response['refresh_token'],
                'expires_in' => $response['expires_in'],
                'user' => [
                    'user_id' => $response['user_id'],
                    'email' => $response['email'],
                    'tenant_id' => $response['tenant_id'],
                    'roles' => $response['roles'] ?? [],
                ]
            ], 200);
        } catch (\Exception $e) {
            // Log auth failure for audit trail
            \Log::warning('Login failed', [
                'email' => $request->input('email'),
                'error' => $e->getMessage(),
                'ip' => $request->ip(),
            ]);

            return response()->json([
                'success' => false,
                'code' => 'AUTHENTICATION_FAILED',
                'message' => 'Invalid email or password',
            ], 401);
        }
    }

    /**
     * Refresh endpoint: exchanges refresh token for new access token.
     *
     * POST /api/v1/auth/refresh
     *
     * @param Request $request
     * @return JsonResponse
     */
    public function refresh(Request $request): JsonResponse
    {
        try {
            $refreshToken = $request->bearerToken();
            if (!$refreshToken) {
                return response()->json([
                    'code' => 'INVALID_REQUEST',
                    'message' => 'Refresh token required',
                ], 400);
            }

            $response = $this->authService->refreshToken($refreshToken);

            return response()->json([
                'success' => true,
                'access_token' => $response['access_token'],
                'expires_in' => $response['expires_in'],
            ], 200);
        } catch (\Exception $e) {
            return response()->json([
                'success' => false,
                'code' => 'TOKEN_INVALID',
                'message' => 'Invalid or expired refresh token',
            ], 401);
        }
    }

    /**
     * Logout endpoint: revokes all tokens for the user.
     *
     * POST /api/v1/auth/logout
     *
     * @param Request $request
     * @return JsonResponse
     */
    public function logout(Request $request): JsonResponse
    {
        try {
            $token = $request->bearerToken();
            if (!$token) {
                return response()->json([
                    'code' => 'INVALID_REQUEST',
                    'message' => 'Token required',
                ], 400);
            }

            // Revoke the token
            $this->authService->revokeToken($token);

            return response()->json([
                'success' => true,
                'message' => 'Logged out successfully',
            ], 200);
        } catch (\Exception $e) {
            return response()->json([
                'success' => false,
                'code' => 'LOGOUT_FAILED',
                'message' => 'Failed to logout',
            ], 500);
        }
    }
}
