import type { Page } from '@playwright/test';

/**
 * E2E test helpers for authentication and user flows.
 * Provides utilities for login, logout, and session management.
 */

export interface TestUser {
	email: string;
	password: string;
	role: 'employee' | 'manager' | 'hr_admin';
	name: string;
}

// Test user credentials (match backend test fixtures)
export const testUsers = {
	employee: {
		email: 'employee@example.com',
		password: 'password123',
		role: 'employee',
		name: 'John Employee'
	} as TestUser,
	manager: {
		email: 'manager@example.com',
		password: 'password123',
		role: 'manager',
		name: 'Jane Manager'
	} as TestUser,
	admin: {
		email: 'admin@example.com',
		password: 'password123',
		role: 'hr_admin',
		name: 'Admin User'
	} as TestUser
};

/**
 * Login a user by navigating to /login and submitting credentials.
 *
 * In test mode (E2E_REAL_BACKEND not set), injects auth tokens directly.
 * In production mode, submits form to real backend.
 */
export async function login(page: Page, user: TestUser) {
	// Test mode: inject auth directly via cookies
	if (!process.env.E2E_REAL_BACKEND) {
		const userObj = {
			user_id: `user-${user.role}-${Date.now()}`,
			tenant_id: 'tenant-test-001',
			email: user.email,
			roles: [user.role]
		};

		// Set auth cookies directly (bypassing login form)
		await page.context().addCookies([
			{
				name: 'auth_token',
				value: `test-token-${user.role}-${Date.now()}`,
				domain: new URL(page.url() || 'http://localhost:4173').hostname || 'localhost',
				path: '/',
				httpOnly: true,
				secure: false,
				sameSite: 'Lax'
			},
			{
				name: 'auth_user',
				value: JSON.stringify(userObj),
				domain: new URL(page.url() || 'http://localhost:4173').hostname || 'localhost',
				path: '/',
				httpOnly: false,
				secure: false,
				sameSite: 'Lax'
			}
		]);

		// Navigate to dashboard (will not redirect to login due to valid session)
		await page.goto('/dashboard');
		await page.waitForURL(/\/dashboard/);
		return;
	}

	// Real backend mode: submit actual login form
	await page.goto('/login');

	// Wait for login form to be visible
	await page.waitForSelector('form');

	// Fill in credentials
	await page.fill('input[type="email"]', user.email);
	await page.fill('input[type="password"]', user.password);

	// Submit form
	await page.click('button[type="submit"]');

	// Wait for redirect to dashboard (or other protected page)
	await page.waitForURL(/^\/(dashboard|admin|leaves)/);
}

/**
 * Logout a user by finding and clicking the logout form action.
 */
export async function logout(page: Page) {
	// Find logout button/form in layout
	// This assumes a logout button exists in the app layout
	const logoutButton = page.locator('button, form').filter({ hasText: /logout|sign out/i }).first();

	if (await logoutButton.isVisible()) {
		await logoutButton.click();
	} else {
		// If no logout button, try to access logout via form action
		await page.evaluate(() => {
			const form = document.querySelector('form[action*="logout"]');
			if (form instanceof HTMLFormElement) {
				form.submit();
			}
		});
	}

	// Wait for redirect to login
	await page.waitForURL(/\/login/);
}

/**
 * Get auth cookies from page storage.
 */
export async function getAuthCookies(page: Page) {
	const cookies = await page.context().cookies();
	return {
		authToken: cookies.find(c => c.name === 'auth_token'),
		authUser: cookies.find(c => c.name === 'auth_user')
	};
}

/**
 * Clear auth cookies.
 */
export async function clearAuthCookies(page: Page) {
	await page.context().clearCookies({ name: 'auth_token' });
	await page.context().clearCookies({ name: 'auth_user' });
}

/**
 * Navigate to a protected page and verify redirect to login if not authenticated.
 */
export async function verifyProtectedRoute(page: Page, path: string) {
	await page.goto(path);
	// Should redirect to login
	await page.waitForURL(/\/login/);
}

/**
 * Navigate to a role-restricted page and verify redirect to 403 if insufficient role.
 */
export async function verifyRoleRestrictedRoute(page: Page, path: string, requiredRole: string) {
	await page.goto(path);
	// Should redirect to 403 for insufficient role
	await page.waitForURL(/\/403/);
}

/**
 * Fill and submit a leave request form.
 */
export async function submitLeaveRequest(
	page: Page,
	data: {
		leaveType: string;
		startDate: string;
		endDate: string;
		reason?: string;
	}
) {
	// Navigate to leave application page
	await page.goto('/leaves/apply');

	// Fill in form
	await page.selectOption('select[name="leave_type_id"]', { label: data.leaveType });
	await page.fill('input[type="date"][id*="start"]', data.startDate);
	await page.fill('input[type="date"][id*="end"]', data.endDate);

	if (data.reason) {
		await page.fill('textarea[name="reason"]', data.reason);
	}

	// Submit form
	await page.click('button[type="submit"]');

	// Wait for success (toast or redirect)
	await page.waitForTimeout(1000); // Wait for mutation to complete
}

/**
 * Approve a leave request from the admin panel.
 */
export async function approveLeaveRequest(page: Page, leaveId: string) {
	// Navigate to admin panel
	await page.goto('/admin/leaves');

	// Find the leave request and click approve
	const leaveCard = page.locator(`[data-leave-id="${leaveId}"]`).first();
	await leaveCard.click();

	const approveButton = leaveCard.locator('button', { hasText: /approve/i });
	await approveButton.click();

	// Wait for success
	await page.waitForTimeout(1000);
}

/**
 * Reject a leave request from the admin panel.
 */
export async function rejectLeaveRequest(page: Page, leaveId: string, reason: string) {
	// Navigate to admin panel
	await page.goto('/admin/leaves');

	// Find the leave request and click reject
	const leaveCard = page.locator(`[data-leave-id="${leaveId}"]`).first();
	const rejectButton = leaveCard.locator('button', { hasText: /reject/i });
	await rejectButton.click();

	// Fill in rejection reason in dialog
	await page.fill('textarea[name="reason"]', reason);

	// Submit dialog
	const submitButton = page.locator('button').filter({ hasText: /confirm|submit/ });
	await submitButton.click();

	// Wait for success
	await page.waitForTimeout(1000);
}
