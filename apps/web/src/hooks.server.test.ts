import { describe, expect, it, vi } from 'vitest';
import { handle } from './hooks.server';

function token(claims: object): string {
	const b64 = Buffer.from(JSON.stringify(claims)).toString('base64url');
	return `h.${b64}.s`;
}

function call(path: string, cookie?: string) {
	const jar = new Map(cookie ? [['hris_session', JSON.stringify({ token: cookie, role: 'employee' })]] : []);
	const event = {
		url: new URL('http://x' + path),
		locals: {} as App.Locals,
		cookies: {
			get: (n: string) => jar.get(n),
			set: vi.fn(),
			delete: vi.fn((n: string) => jar.delete(n))
		}
	};
	const resolve = vi.fn(async () => new Response('ok'));
	const run = () => handle({ event, resolve } as never);
	return { event, resolve, run };
}

async function redirectOf(run: () => unknown) {
	try {
		await run();
	} catch (e) {
		return e as { status: number; location: string };
	}
	throw new Error('expected redirect');
}

const employee = token({ user_id: 'u', tenant_id: 't', email: 'e@x', roles: ['employee'] });
const manager = token({ user_id: 'u', tenant_id: 't', email: 'e@x', roles: ['manager'] });

describe('hooks guard', () => {
	it('lets public paths through and guards look-alike prefixes', async () => {
		const pub = call('/register');
		await pub.run();
		expect(pub.resolve).toHaveBeenCalled();

		const r = await redirectOf(call('/registerx').run);
		expect(r.location).toBe('/login?redirect=%2Fregisterx');
	});

	it('takes user and roles from the token', async () => {
		const c = call('/dashboard', employee);
		await c.run();
		expect(c.event.locals.user).toEqual({ user_id: 'u', tenant_id: 't', email: 'e@x', roles: ['employee'] });
	});

	it('sends non-approvers on /admin to 403, allows managers', async () => {
		const r = await redirectOf(call('/admin/leaves', employee).run);
		expect(r.location).toBe('/403?requested=%2Fadmin%2Fleaves');

		const ok = call('/admin/leaves', manager);
		await ok.run();
		expect(ok.resolve).toHaveBeenCalled();
	});

	it('clears an undecodable session', async () => {
		const c = call('/dashboard', 'not-a-jwt');
		const r = await redirectOf(c.run);
		expect(r.location).toContain('error=invalid_session');
		expect(c.event.cookies.delete).toHaveBeenCalled();
	});
});
