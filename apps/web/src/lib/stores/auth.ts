import { writable, derived } from 'svelte/store';

/**
 * Authentication store manages JWT token and user session state.
 *
 * Stores:
 * - token: JWT access token
 * - user: Decoded user claims (user_id, tenant_id, roles, email)
 * - isAuthenticated: Computed boolean indicating active session
 */

export interface AuthUser {
	user_id: string;
	tenant_id: string;
	email: string;
	roles: string[];
	sub?: string;
}

interface AuthState {
	token: string | null;
	user: AuthUser | null;
}

// Initialize from localStorage if available
function createAuthStore() {
	const initialState: AuthState = {
		token: null,
		user: null
	};

	// Try to restore from localStorage on load
	if (typeof window !== 'undefined') {
		const storedToken = localStorage.getItem('auth_token');
		const storedUser = localStorage.getItem('auth_user');

		if (storedToken && storedUser) {
			try {
				initialState.token = storedToken;
				initialState.user = JSON.parse(storedUser);
			} catch (e) {
				// Invalid stored data, clear it
				localStorage.removeItem('auth_token');
				localStorage.removeItem('auth_user');
			}
		}
	}

	const { subscribe, set, update } = writable<AuthState>(initialState);

	return {
		subscribe,

		/**
		 * Set authenticated user session from login response.
		 * Stores token and user claims in both store and localStorage.
		 */
		setSession: (token: string, user: AuthUser) => {
			const state = { token, user };
			if (typeof window !== 'undefined') {
				localStorage.setItem('auth_token', token);
				localStorage.setItem('auth_user', JSON.stringify(user));
			}
			set(state);
		},

		/**
		 * Clear session and logout user.
		 * Removes token and user from store and localStorage.
		 */
		clearSession: () => {
			if (typeof window !== 'undefined') {
				localStorage.removeItem('auth_token');
				localStorage.removeItem('auth_user');
			}
			set({ token: null, user: null });
		},

		/**
		 * Update token (e.g., after refresh).
		 */
		setToken: (token: string) => {
			update((state) => {
				if (typeof window !== 'undefined') {
					localStorage.setItem('auth_token', token);
				}
				return { ...state, token };
			});
		}
	};
}

export const auth = createAuthStore();

/**
 * Derived store: isAuthenticated
 * True if token exists and user is set
 */
export const isAuthenticated = derived(auth, ($auth) => {
	return !!$auth.token && !!$auth.user;
});

/**
 * Derived store: user info
 * Returns the current user or null
 */
export const user = derived(auth, ($auth) => $auth.user);

/**
 * Derived store: roles
 * Returns user's roles array or empty array
 */
export const roles = derived(auth, ($auth) => $auth.user?.roles ?? []);

/**
 * Derived store: hasRole
 * Returns function to check if user has a specific role
 */
export const hasRole = derived(roles, ($roles) => {
	return (role: string) => $roles.includes(role);
});

/**
 * Derived store: isManager
 * True if user has manager or hr_admin role
 */
export const isManager = derived(
	roles,
	($roles) => $roles.includes('manager') || $roles.includes('hr_admin')
);

/**
 * Derived store: isAdmin
 * True if user has hr_admin role
 */
export const isAdmin = derived(roles, ($roles) => $roles.includes('hr_admin'));

/**
 * Derived store: tenantId
 * Returns current tenant ID
 */
export const tenantId = derived(auth, ($auth) => $auth.user?.tenant_id ?? null);
