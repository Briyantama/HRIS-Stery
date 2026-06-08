<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Illuminate\Http\Response;

class TenantResolver
{
    /**
     * Handle an incoming request to enforce tenant isolation.
     *
     * Extract tenant_id exclusively from JWT claims and inject it into the request.
     * Prevents frontend from specifying or tampering with tenant_id.
     *
     * @param  Request  $request
     * @param  Closure  $next
     * @return mixed
     */
    public function handle(Request $request, Closure $next)
    {
        // Skip for public routes
        if ($this->isPublicRoute($request)) {
            return $next($request);
        }

        // Get user claims from JWT validation middleware
        $user = $request->attributes->get('user');
        if (!$user || !$user->tenant_id) {
            return response()->json([
                'code' => 'PERMISSION_DENIED',
                'message' => 'Tenant context not found in JWT claims',
            ], Response::HTTP_FORBIDDEN);
        }

        $tenantId = $user->tenant_id;

        // Check if request contains a tenant_id parameter (either in body or query)
        $requestTenantId = $request->input('tenant_id') ?? $request->query('tenant_id');

        // If frontend supplied a tenant_id, validate it matches JWT
        if ($requestTenantId && $requestTenantId !== $tenantId) {
            return response()->json([
                'code' => 'PERMISSION_DENIED',
                'message' => 'Tenant ID mismatch: request tenant does not match authenticated tenant',
            ], Response::HTTP_FORBIDDEN);
        }

        // Override/inject tenant_id from JWT into request
        // This ensures all backend calls use the correct, validated tenant_id
        $request->merge(['tenant_id' => $tenantId]);

        // Store in attributes for service classes to access
        $request->attributes->set('tenant_id', $tenantId);
        $request->attributes->set('user_id', $user->user_id ?? $user->id);

        return $next($request);
    }

    /**
     * Check if route is public (no tenant enforcement required).
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
