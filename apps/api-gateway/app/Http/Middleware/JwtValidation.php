<?php

namespace App\Http\Middleware;

use Closure;
use Firebase\JWT\JWT;
use Firebase\JWT\Key;
use Illuminate\Http\Request;
use Illuminate\Http\Response;

class JwtValidation
{
    /**
     * Handle an incoming request to validate JWT token.
     *
     * @param  Request  $request
     * @param  Closure  $next
     * @return mixed
     */
    public function handle(Request $request, Closure $next)
    {
        // Skip validation for public routes if needed
        if ($this->isPublicRoute($request)) {
            return $next($request);
        }

        // Extract JWT from Authorization header
        $token = $this->extractToken($request);
        if (!$token) {
            return response()->json([
                'code' => 'UNAUTHENTICATED',
                'message' => 'Missing or malformed Authorization header',
            ], Response::HTTP_UNAUTHORIZED);
        }

        try {
            // Get public key from environment
            $publicKey = env('AUTH_PUBLIC_KEY');
            if (!$publicKey) {
                throw new \Exception('AUTH_PUBLIC_KEY not configured');
            }

            // Validate and decode JWT
            $decoded = JWT::decode($token, new Key($publicKey, 'RS256'));

            // Attach claims to request
            $request->attributes->set('user', (object) [
                'id' => $decoded->sub ?? $decoded->user_id ?? null,
                'user_id' => $decoded->user_id ?? $decoded->sub ?? null,
                'tenant_id' => $decoded->tid ?? $decoded->tenant_id ?? null,
                'email' => $decoded->email ?? null,
                'roles' => $decoded->roles ?? [],
            ]);

            // Store full claims for later access
            $request->attributes->set('jwt_claims', $decoded);

            return $next($request);
        } catch (\Firebase\JWT\ExpiredException $e) {
            return response()->json([
                'code' => 'UNAUTHENTICATED',
                'message' => 'Token has expired',
                'details' => $e->getMessage(),
            ], Response::HTTP_UNAUTHORIZED);
        } catch (\Firebase\JWT\SignatureInvalidException $e) {
            return response()->json([
                'code' => 'UNAUTHENTICATED',
                'message' => 'Invalid token signature',
                'details' => $e->getMessage(),
            ], Response::HTTP_UNAUTHORIZED);
        } catch (\Exception $e) {
            return response()->json([
                'code' => 'UNAUTHENTICATED',
                'message' => 'Token validation failed',
                'details' => $e->getMessage(),
            ], Response::HTTP_UNAUTHORIZED);
        }
    }

    /**
     * Extract JWT token from Authorization header.
     */
    private function extractToken(Request $request): ?string
    {
        $header = $request->header('Authorization');
        if (!$header) {
            return null;
        }

        // Expected format: "Bearer <token>"
        if (!str_starts_with($header, 'Bearer ')) {
            return null;
        }

        return substr($header, 7);
    }

    /**
     * Check if route is public (no JWT validation required).
     */
    private function isPublicRoute(Request $request): bool
    {
        $publicPaths = [
            '/api/auth/login',
            '/api/auth/register',
            '/api/auth/refresh',
            '/health',
        ];

        foreach ($publicPaths as $path) {
            if ($request->is($path)) {
                return true;
            }
        }

        return false;
    }
}
