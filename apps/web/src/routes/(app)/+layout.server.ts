import type { LayoutServerLoad } from './$types';

/** hooks.server.ts guarantees locals.user on every (app) route. */
export const load: LayoutServerLoad = async ({ locals }) => {
	const user = locals.user ?? { user_id: '', tenant_id: '', email: '', roles: [] };
	return {
		user,
		isManager: user.roles.includes('manager'),
		isAdmin: user.roles.includes('hr_admin'),
		isAuthenticated: !!locals.user
	};
};
