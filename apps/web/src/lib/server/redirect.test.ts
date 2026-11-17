import { describe, expect, it } from 'vitest';
import { safeRedirect } from './redirect';

describe('safeRedirect', () => {
	it('allows local paths with query', () => {
		expect(safeRedirect('/leave?tab=2')).toBe('/leave?tab=2');
	});
	it.each([null, '', 'leave', '//evil.com', '/\\evil.com', 'https://evil.com', '/a\nb', '/a\tb', '\\\\evil'])(
		'falls back for %j',
		(t) => {
			expect(safeRedirect(t)).toBe('/dashboard');
		}
	);
	it('honors custom fallback', () => {
		expect(safeRedirect('//x', '/home')).toBe('/home');
	});
});
