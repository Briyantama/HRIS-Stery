<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Illuminate\Http\Response;
use Illuminate\Support\Facades\Log;

/**
 * CheckRole middleware enforces role-based access control at the gateway edge.
 */
class CheckRole
{
    public function handle(Request $request, Closure $next, ...$roles)
    {
        $user = $request->attributes->get('user');

        if (! $user) {
            return response()->json([
                'code' => 'UNAUTHENTICATED',
                'message' => 'User authentication context not found',
            ], Response::HTTP_UNAUTHORIZED);
        }

        $userRoles = $user->roles ?? [];

        if (! $this->userHasRole($userRoles, $roles)) {
            Log::warning('Unauthorized access attempt', [
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

    private function userHasRole(array $userRoles, array $requiredRoles): bool
    {
        $userRolesLower = array_map('strtolower', $userRoles);

        foreach ($requiredRoles as $requiredRole) {
            if (in_array(strtolower($requiredRole), $userRolesLower)) {
                return true;
            }
        }

        return false;
    }
}
