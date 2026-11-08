import { goto } from '$app/navigation';
import { createClient, type RequestOptions } from '@stery/kit';

/**
 * Browser client for the BFF proxy: /api/<path> reaches the Go API's /api/v1/<path> with the
 * session's bearer token added server-side. Every call returns the envelope's `data`; failures are
 * ApiError. A 401 means the session is gone, so the user goes back to the login page.
 */
const client = createClient('/api', {
	onUnauthorized: () => {
		if (typeof window !== 'undefined') {
			// no base path; the query string makes resolve() unusable here
			// eslint-disable-next-line svelte/no-navigation-without-resolve
			void goto('/login?redirect=' + encodeURIComponent(location.pathname + location.search));
		}
	}
});

export const apiGet = <T>(path: string, options?: RequestOptions) => client.get<T>(path, options);
export const apiPost = <T>(path: string, body?: unknown, options?: RequestOptions) =>
	client.post<T>(path, body, options);

/** A message fit for the user. */
export function getErrorMessage(error: unknown): string {
	if (error instanceof Error && error.message) return error.message;
	return 'An unexpected error occurred';
}
