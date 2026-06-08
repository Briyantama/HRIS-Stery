<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Illuminate\Http\Response;

/**
 * CheckRole middleware enforces role-based access control at the gateway edge.
 *
 * This middleware intercepts requests and validates that the authenticated user
 * possesses one of the required roles. If not, a 403 Forbidden response is
 * returned immediately, preventing the request from reaching backend microservices.
 *
 * Usage:
 *   Route::post('/leaves/{id}/approve', ...)
 *       ->middleware('role:manager,hr_admin')
 *
 * Or in route group:
 *   Route::middleware('role:manager')->group(...)
 */
class CheckRole
{
    /**
     * Handle an incoming request to verify user role authorization.
     *
     * @param  Request  $request
     * @param  Closure  $next
     * @param  string  ...$roles  Comma-separated role names required for access
     * @return mixed
     */
    public function handle(Request $request, Closure $next, ...$roles)
    {
        // Retrieve user from request (set by JwtValidation middleware)
        $user = $request->attributes->get('user');

        // Verify user is authenticated
        if (!$user) {
            return response()->json([
                'code' => 'UNAUTHENTICATED',
                'message' => 'User authentication context not found',
            ], Response::HTTP_UNAUTHORIZED);
        }

        // Extract user roles
        $userRoles = $user->roles ?? [];

        // Check if user has at least one of the required roles
        if (!$this->userHasRole($userRoles, $roles)) {
            // Log unauthorized access attempt
            \Log::warning('Unauthorized access attempt', [
                'user_id' => $user->user_id ?? 'unknown',
                'tenant_id' => $user->tenant_id ?? 'unknown',
                'endpoint' => $request->getPathInfo(),
                'method' => $request->getMethod(),
                'required_roles' => $roles,
                'user_roles' => $userRoles,
                'ip_address' => $request->ip(),
            ]);

            return response()->json([
                'code' => 'PERMISSION_DENIED',
                'message' => 'Insufficient permissions for this operation',
                'details' => [
                    'required_roles' => $roles,
                    'user_roles' => $userRoles,
                ],
            ], Response::HTTP_FORBIDDEN);
        }

        return $next($request);
    }

    /**
     * Check if user has at least one of the required roles.
     *
     * @param  array  $userRoles  Roles assigned to the user
     * @param  array  $requiredRoles  Roles required for access
     * @return bool  True if user has at least one required role
     */
    private function userHasRole(array $userRoles, array $requiredRoles): bool
    {
        // Convert user roles to lowercase for case-insensitive comparison
        $userRolesLower = array_map('strtolower', $userRoles);

        // Check if any required role matches user roles
        foreach ($requiredRoles as $requiredRole) {
            if (in_array(strtolower($requiredRole), $userRolesLower)) {
                return true;
            }
        }

        return false;
    }
}
