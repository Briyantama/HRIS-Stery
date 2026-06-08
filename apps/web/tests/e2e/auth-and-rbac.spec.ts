import { expect, test } from '@playwright/test';
import { login, logout, testUsers, getAuthCookies, clearAuthCookies, verifyProtectedRoute } from './helpers';

/**
 * E2E tests for authentication and RBAC enforcement.
 *
 * Tests:
 * 1. Login with valid credentials
 * 2. Login with invalid credentials
 * 3. Auth cookies are set correctly
 * 4. Session persists across page refresh
 * 5. Unauthenticated users cannot access protected routes
 * 6. Logout clears session and redirects
 */

test.describe('Authentication Flow', () => {
	test('login succeeds with valid employee credentials', async ({ page, context }) => {
		await login(page, testUsers.employee);

		// Verify redirect to dashboard
		expect(page.url()).toMatch(/\/dashboard/);

		// Verify auth cookies are set
		const { authToken, authUser } = await getAuthCookies(page);
		expect(authToken).toBeDefined();
		expect(authUser).toBeDefined();

		// Verify auth_user contains expected fields
		const userObj = JSON.parse(authUser?.value || '{}');
		expect(userObj.email).toBe(testUsers.employee.email);
		expect(userObj.roles).toContain('employee');
	});

	test('login succeeds with valid manager credentials', async ({ page }) => {
		await login(page, testUsers.manager);

		// Verify redirect to dashboard
		expect(page.url()).toMatch(/\/dashboard/);

		// Verify auth_user contains manager role
		const { authUser } = await getAuthCookies(page);
		const userObj = JSON.parse(authUser?.value || '{}');
		expect(userObj.roles).toContain('manager');
	});

	test('login fails with invalid credentials', async ({ page }) => {
		await page.goto('/login');

		// Fill in incorrect password
		await page.fill('input[type="email"]', testUsers.employee.email);
		await page.fill('input[type="password"]', 'wrongpassword');

		// Submit form
		await page.click('button[type="submit"]');

		// Should remain on login page with error message
		await page.waitForTimeout(500); // Wait for error display
		expect(page.url()).toMatch(/\/login/);

		// Verify error message is visible
		const errorMessage = page.locator('text=/authentication failed|invalid/i');
		expect(await errorMessage.isVisible()).toBe(true);
	});

	test('session persists across page refresh', async ({ page }) => {
		await login(page, testUsers.employee);

		// Verify logged in
		expect(page.url()).toMatch(/\/dashboard/);

		// Get auth cookies before refresh
		const { authToken: tokenBefore } = await getAuthCookies(page);
		expect(tokenBefore).toBeDefined();

		// Refresh page
		await page.reload();

		// Verify still logged in (no redirect to login)
		expect(page.url()).toMatch(/\/dashboard/);

		// Verify cookies still present
		const { authToken: tokenAfter } = await getAuthCookies(page);
		expect(tokenAfter).toBeDefined();
		expect(tokenAfter?.value).toBe(tokenBefore?.value);
	});

	test('logout clears session and redirects to login', async ({ page }) => {
		// First login
		await login(page, testUsers.employee);
		expect(page.url()).toMatch(/\/dashboard/);

		// Verify logged in
		const { authToken: tokenBefore } = await getAuthCookies(page);
		expect(tokenBefore).toBeDefined();

		// Logout
		await logout(page);

		// Verify redirect to login
		expect(page.url()).toMatch(/\/login/);

		// Verify cookies are cleared
		const { authToken, authUser } = await getAuthCookies(page);
		expect(authToken).toBeUndefined();
		expect(authUser).toBeUndefined();
	});
});

test.describe('Route Protection', () => {
	test('unauthenticated users cannot access /dashboard', async ({ page, context }) => {
		// Clear any existing cookies
		await context.clearCookies();

		// Navigate to protected route
		await verifyProtectedRoute(page, '/dashboard');

		// Should be on login page
		expect(page.url()).toMatch(/\/login/);
	});

	test('unauthenticated users cannot access /leaves', async ({ page, context }) => {
		await context.clearCookies();
		await verifyProtectedRoute(page, '/leaves');
		expect(page.url()).toMatch(/\/login/);
	});

	test('unauthenticated users cannot access /leaves/apply', async ({ page, context }) => {
		await context.clearCookies();
		await verifyProtectedRoute(page, '/leaves/apply');
		expect(page.url()).toMatch(/\/login/);
	});

	test('authenticated user can access /dashboard', async ({ page }) => {
		await login(page, testUsers.employee);

		// Navigate to dashboard
		await page.goto('/dashboard');

		// Should remain on dashboard (not redirect to login)
		expect(page.url()).toMatch(/\/dashboard/);
		expect(page.locator('text=/leave balance|dashboard/i').first()).toBeVisible();
	});

	test('authenticated user can access /leaves', async ({ page }) => {
		await login(page, testUsers.employee);

		// Navigate to leaves
		await page.goto('/leaves');

		// Should remain on leaves page
		expect(page.url()).toMatch(/\/leaves/);
	});

	test('authenticated user can access /leaves/apply', async ({ page }) => {
		await login(page, testUsers.employee);

		// Navigate to apply page
		await page.goto('/leaves/apply');

		// Should remain on apply page
		expect(page.url()).toMatch(/\/leaves\/apply/);
		expect(page.locator('button[type="submit"]')).toBeVisible();
	});
});

test.describe('Role-Based Access Control', () => {
	test('employee cannot access /admin/leaves', async ({ page }) => {
		await login(page, testUsers.employee);

		// Try to navigate to admin page
		await page.goto('/admin/leaves');

		// Should redirect to 403
		await page.waitForURL(/\/403/);
		expect(page.url()).toMatch(/\/403/);

		// Verify 403 error message
		expect(page.locator('text=/access denied|permission/i').first()).toBeVisible();
	});

	test('manager can access /admin/leaves', async ({ page }) => {
		await login(page, testUsers.manager);

		// Navigate to admin page
		await page.goto('/admin/leaves');

		// Should remain on admin page
		expect(page.url()).toMatch(/\/admin\/leaves/);
		expect(page.locator('button', { hasText: /approve|reject/i }).first()).toBeVisible();
	});

	test('hr_admin can access /admin/leaves', async ({ page }) => {
		await login(page, testUsers.admin);

		// Navigate to admin page
		await page.goto('/admin/leaves');

		// Should remain on admin page
		expect(page.url()).toMatch(/\/admin\/leaves/);
		expect(page.locator('button', { hasText: /approve|reject/i }).first()).toBeVisible();
	});

	test('direct URL navigation to /admin/leaves is enforced for non-managers', async ({ page, context }) => {
		// Don't login, just navigate directly to admin page
		await context.clearCookies();
		await page.goto('/admin/leaves');

		// Should redirect to login (unauthenticated)
		await page.waitForURL(/\/login/);
		expect(page.url()).toMatch(/\/login/);
	});
});
