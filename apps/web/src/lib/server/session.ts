import { dev } from '$app/environment';
import { env } from '$env/dynamic/private';
import { createSessionStore } from '@stery/kit/sveltekit/session';

/** Base URL of the Go API (without /api/v1). Server-only: the browser goes through /api. */
export const apiUrl = () => env.API_URL ?? 'http://localhost:8000';

/** The access token lives in this HttpOnly cookie and never reaches client JS. */
export const session = createSessionStore({ cookie: 'hris_session', secure: !dev });
