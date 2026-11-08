import { redirect, type Handle } from '@sveltejs/kit';
import { session } from '$lib/server/session';

const PUBLIC = ['/login', '/register', '/health'];

function clearAndRedirect(event: Parameters<Handle>[0]['event'], path: string): never {
	session.clear(event.cookies);
	redirect(303, `/login?error=invalid_session&redirect=${encodeURIComponent(path)}`);
}

/** Claims come from the access token payload. The API verifies the signature on every call. */
function claimsFromToken(token: string): App.Locals['user'] | null {
	try {
		const seg = token.split('.')[1];
		const b64 = seg.replace(/-/g, '+').replace(/_/g, '/');
		const c = JSON.parse(new TextDecoder().decode(Uint8Array.from(atob(b64), (ch) => ch.charCodeAt(0))));
		if (!c?.user_id || !c.tenant_id) return null;
		return { user_id: c.user_id, tenant_id: c.tenant_id, email: c.email, roles: Array.isArray(c.roles) ? c.roles : [] };
	} catch {
		return null;
	}
}

/**
 * Route guard. Pages need a session (HttpOnly cookie set by the login action); /admin needs the
 * manager or hr_admin role. /api is the BFF proxy and answers 401 itself.
 */
export const handle: Handle = async ({ event, resolve }) => {
	const path = event.url.pathname;
	const isPrefix = (p: string) => path === p || path.startsWith(p + '/');
	if (isPrefix('/api') || PUBLIC.some(isPrefix)) return resolve(event);

	const current = session.read(event.cookies);
	if (!current) redirect(303, `/login?redirect=${encodeURIComponent(path + event.url.search)}`);

	const user = claimsFromToken(current.token);
	if (!user) clearAndRedirect(event, path);

	event.locals.token = current.token;
	event.locals.user = user;

	if (isPrefix('/admin') && !user.roles.some((r) => r === 'manager' || r === 'hr_admin')) {
		redirect(303, `/403?requested=${encodeURIComponent(path)}`);
	}
	return resolve(event);
};
