import { fail, redirect } from '@sveltejs/kit';
import { ApiError, unwrap } from '@stery/kit';
import { apiUrl, session } from '$lib/server/session';
import { safeRedirect } from '$lib/server/redirect';
import type { Actions, PageServerLoad } from './$types';

interface LoginData {
	access_token: string;
	claims: { user_id: string; tenant_id: string; email: string; roles: string[] };
}

export const load: PageServerLoad = async ({ cookies, url }) => {
	const target = safeRedirect(url.searchParams.get('redirect'));
	if (session.read(cookies)) redirect(303, target);
	return { error: url.searchParams.get('error'), redirect: target };
};

export const actions = {
	login: async ({ cookies, request, fetch }) => {
		const form = await request.formData();
		const email = String(form.get('email') ?? '').trim();
		const password = String(form.get('password') ?? '');
		const tenantSlug = String(form.get('tenant_slug') ?? '').trim();
		if (!email || !password || !tenantSlug) {
			return fail(400, { error: 'Organization, email and password are required' });
		}

		let data: LoginData;
		try {
			const res = await fetch(`${apiUrl()}/api/v1/auth/login`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email, password, tenant_slug: tenantSlug })
			});
			data = await unwrap<LoginData>(res);
		} catch (err) {
			const status = err instanceof ApiError ? err.status : 0;
			if (status === 0 || status >= 500) return fail(502, { error: 'Cannot reach the HRIS API' });
			return fail(status, { error: (err as ApiError).message });
		}

		session.write(cookies, { token: data.access_token, role: data.claims.roles.join(',') });
		redirect(303, safeRedirect(String(form.get('redirect') ?? '')));
	}
} satisfies Actions;
