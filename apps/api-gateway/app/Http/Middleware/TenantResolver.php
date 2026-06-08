<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Illuminate\Http\Response;

class TenantResolver
{
    public function handle(Request $request, Closure $next)
    {
        if ($this->isPublicRoute($request)) {
            return $next($request);
        }

        $user = $request->attributes->get('user');
        if (! $user || ! $user->tenant_id) {
            return response()->json([
                'code' => 'PERMISSION_DENIED',
                'message' => 'Tenant context not found in JWT claims',
            ], Response::HTTP_FORBIDDEN);
        }

        $tenantId = $user->tenant_id;
        $requestTenantId = $request->input('tenant_id') ?? $request->query('tenant_id');

        if ($requestTenantId && $requestTenantId !== $tenantId) {
            return response()->json([
                'code' => 'PERMISSION_DENIED',
                'message' => 'Tenant ID mismatch: request tenant does not match authenticated tenant',
            ], Response::HTTP_FORBIDDEN);
        }

        $request->merge(['tenant_id' => $tenantId]);
        $request->attributes->set('tenant_id', $tenantId);
        $request->attributes->set('user_id', $user->user_id ?? $user->id);

        return $next($request);
    }

    private function isPublicRoute(Request $request): bool
    {
        $publicPaths = [
            '/api/auth/login',
            '/api/auth/register',
            '/api/auth/refresh',
            '/api/health',
            '/up',
        ];

        foreach ($publicPaths as $path) {
            if ($request->is($path)) {
                return true;
            }
        }

        return false;
    }
}
