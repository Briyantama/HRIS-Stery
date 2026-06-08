import { redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';

/**
 * Server-side authentication handler for login page.
 * Manages session establishment and cookie-based auth for server-side route guards.
 */

export const load: PageServerLoad = async ({ cookies, url }) => {
	// If already authenticated, redirect to dashboard
	const authToken = cookies.get('auth_token');
	if (authToken) {
		const redirectUrl = url.searchParams.get('redirect') || '/dashboard';
		throw redirect(307, redirectUrl);
	}

	return {
		error: url.searchParams.get('error'),
		redirect: url.searchParams.get('redirect')
	};
};

export const actions = {
	login: async ({ cookies, request }) => {
		const formData = await request.formData();
		const email = formData.get('email') as string;
		const password = formData.get('password') as string;

		// Validate inputs
		if (!email || !password) {
			return {
				success: false,
				error: 'Email and password are required'
			};
		}

		try {
			// Call auth API (mocked for now - implement with actual backend)
			// In production, this would call: POST /api/auth/login
			const response = await fetch('http://localhost:3000/api/auth/login', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email, password })
			});

			if (!response.ok) {
				const error = await response.json();
				return {
					success: false,
					error: error.message || 'Authentication failed'
				};
			}

			const { access_token, user } = await response.json();

			// Set auth cookies for server-side route guards
			// Using httpOnly + Secure for production security
			cookies.set('auth_token', access_token, {
				path: '/',
				httpOnly: true,
				secure: process.env.NODE_ENV === 'production',
				sameSite: 'lax',
				maxAge: 60 * 60 * 24 * 7 // 7 days
			});

			cookies.set('auth_user', JSON.stringify(user), {
				path: '/',
				httpOnly: false, // Client-side code needs to read this
				secure: process.env.NODE_ENV === 'production',
				sameSite: 'lax',
				maxAge: 60 * 60 * 24 * 7
			});

			// Redirect to dashboard or requested page
			throw redirect(307, '/dashboard');
		} catch (error) {
			if (error instanceof Error && error.message.startsWith('Redirect')) {
				throw error;
			}
			return {
				success: false,
				error: error instanceof Error ? error.message : 'An unexpected error occurred'
			};
		}
	}
};
