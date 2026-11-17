import { createApiProxy } from '@stery/kit/sveltekit/proxy';
import { apiUrl, session } from '$lib/server/session';

/** BFF proxy: /api/<path> goes to the Go API's /api/v1/<path> with the session's bearer token. */
export const { GET, POST, PUT, PATCH, DELETE } = createApiProxy({
	upstream: apiUrl,
	accessCookie: session.cookie,
	onUnauthorized: ({ cookies }) => session.clear(cookies)
});
