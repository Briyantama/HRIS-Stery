import { redirect } from '@sveltejs/kit';
import type { Handle } from '@sveltejs/kit';

/**
 * Server-side authentication and authorization hooks.
 *
 * This hook enforces:
 * 1. Authentication check: redirects unauthenticated users to /login
 * 2. Route protection: (app) group requires active session
 * 3. Role-based access: /admin/leaves requires manager or hr_admin role
 * 4. Session restoration: preserves auth state across page refreshes
 */

export const handle: Handle = async ({ event, resolve }) => {
	// Public routes that don't require authentication
	const publicRoutes = ['/login', '/register', '/health'];
	const currentPath = event.url.pathname;

	// Check if current route is public
	const isPublicRoute = publicRoutes.some((route) => currentPath.startsWith(route));

	// If public route, skip authentication check
	if (isPublicRoute) {
		return resolve(event);
	}

	// Verify authentication for protected routes
	const authToken = event.cookies.get('auth_token');
	const authUserStr = event.cookies.get('auth_user');

	// If accessing protected route without token, redirect to login
	if (!authToken || !authUserStr) {
		throw redirect(307, `/login?redirect=${encodeURIComponent(currentPath)}`);
	}

	// Parse and validate user object
	let authUser: any;
	try {
		authUser = JSON.parse(authUserStr);
	} catch {
		// Invalid auth data, clear session and redirect to login
		event.cookies.delete('auth_token', { path: '/' });
		event.cookies.delete('auth_user', { path: '/' });
		throw redirect(307, `/login?error=invalid_session&redirect=${encodeURIComponent(currentPath)}`);
	}

	// Validate required fields exist
	if (!authUser.user_id || !authUser.tenant_id) {
		event.cookies.delete('auth_token', { path: '/' });
		event.cookies.delete('auth_user', { path: '/' });
		throw redirect(307, `/login?error=invalid_session&redirect=${encodeURIComponent(currentPath)}`);
	}

	// Inject auth data into event for use in +page.server.ts files
	event.locals.token = authToken;
	event.locals.user = authUser;

	// Role-based route protection: /admin/* routes require manager or hr_admin role
	if (currentPath.startsWith('/admin')) {
		const userRoles = authUser.roles || [];
		const hasRequiredRole = userRoles.some((role: string) =>
			role.toLowerCase() === 'manager' || role.toLowerCase() === 'hr_admin'
		);

		if (!hasRequiredRole) {
			// User is authenticated but lacks required role, redirect to access denied
			throw redirect(307, `/403?requested=${encodeURIComponent(currentPath)}`);
		}
	}

	return resolve(event);
};
