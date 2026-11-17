import { derived } from 'svelte/store';
import { page } from '$app/stores';

export interface AuthUser {
	user_id: string;
	tenant_id: string;
	email: string;
	roles: string[];
}

/** The signed-in user, from the (app) layout's server load (the session itself is an HttpOnly cookie). */
export const user = derived(page, ($page) => ($page.data.user as AuthUser | undefined) ?? null);

/** True for manager or hr_admin. */
export const isManager = derived(user, ($user) =>
	($user?.roles ?? []).some((r) => r === 'manager' || r === 'hr_admin')
);
