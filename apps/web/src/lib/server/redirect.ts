/** Returns `target` only if it is a same-origin absolute path; otherwise `fallback`. */
export function safeRedirect(target: string | null | undefined, fallback = '/dashboard'): string {
	if (!target || !/^\/(?![/\\])/.test(target)) return fallback;
	// eslint-disable-next-line no-control-regex
	if (/[\u0000-\u001f\u007f]/.test(target)) return fallback;
	return target;
}
