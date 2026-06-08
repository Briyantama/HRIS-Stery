import { chromium } from '@playwright/test';
import type { FullConfig } from '@playwright/test';

/**
 * Global setup for E2E tests.
 *
 * This runs once before all tests. We use it to:
 * 1. Verify the frontend is running
 * 2. Set environment variables for the test session
 * 3. Pre-configure test mode
 */

async function globalSetup(config: FullConfig) {
	const baseURL = process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:4173';

	console.log(`\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`);
	console.log(`E2E Test Suite Configuration`);
	console.log(`━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`);
	console.log(`Frontend URL:     ${baseURL}`);
	console.log(`Test Mode:        ${process.env.E2E_TEST_MODE || 'mock-auth (no backend required)'}`);
	console.log(`Backend Mode:     ${process.env.E2E_REAL_BACKEND ? 'Real backend (requires make dev + services running)' : 'Mock auth (frontend-only tests)'}`);
	console.log(`CI Environment:   ${process.env.CI ? 'yes' : 'no'}`);
	console.log(`━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n`);

	// Verify frontend is reachable
	const browser = await chromium.launch();
	const page = await browser.newPage();

	try {
		const response = await page.goto(`${baseURL}/`, { waitUntil: 'domcontentloaded' });
		if (!response || !response.ok()) {
			throw new Error(`Frontend not accessible at ${baseURL}`);
		}
		console.log(`✅ Frontend verified at ${baseURL}`);
	} catch (error) {
		console.error(`❌ Frontend verification failed:`, error instanceof Error ? error.message : String(error));
		console.error(`\nMake sure the frontend is running:`);
		console.error(`  cd apps/web && npm run preview -- --port 4173 --host`);
		process.exit(1);
	} finally {
		await browser.close();
	}

	// Note about backend
	if (!process.env.E2E_REAL_BACKEND) {
		console.log(`\n📌 Backend Integration Tests:`);
		console.log(`   Tests use mocked authentication (direct cookie injection).`);
		console.log(`   For full E2E with real backend, run:`);
		console.log(`   E2E_REAL_BACKEND=1 make test-e2e`);
		console.log(`\n   This requires:`);
		console.log(`   - make dev              (start Docker infrastructure)`);
		console.log(`   - make migrate-up       (run database migrations)`);
		console.log(`   - make dev-laravel      (start Laravel gateway)`);
		console.log(`   - make dev-svelte       (start SvelteKit dev server)`);
		console.log(`   - make dev-apps (after services bootstrapped)`);
		console.log();
	}

	return {};
}

export default globalSetup;
