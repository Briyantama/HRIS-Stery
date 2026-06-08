import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

/**
 * Server-side layout handler for protected (app) routes.
 * Provides authenticated user data to all pages in this group.
 * Handles logout action across all protected pages.
 */

export const load: LayoutServerLoad = async ({ locals }) => {
	// This hook.server.ts ensures authenticated users reach here
	// Locals populated by hooks.server.ts with verified auth data
	const user = locals.user || { user_id: '', tenant_id: '', email: '', roles: [] };
	return {
		user,
		isManager: user.roles?.some((r: string) => r.toLowerCase() === 'manager') ?? false,
		isAdmin: user.roles?.some((r: string) => r.toLowerCase() === 'hr_admin') ?? false,
		isAuthenticated: !!locals.user
	};
};

export const actions: any = {
	logout: async ({ cookies }: any) => {
		// Clear auth cookies
		cookies.delete('auth_token', { path: '/' });
		cookies.delete('auth_user', { path: '/' });

		// Redirect to login
		throw redirect(307, '/login');
	}
};
