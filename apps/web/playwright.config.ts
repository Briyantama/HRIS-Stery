import { defineConfig, devices } from '@playwright/test';
import path from 'path';

export default defineConfig({
	testDir: 'tests/e2e',
	fullyParallel: true,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 0,
	reporter: 'html',
	use: {
		baseURL: process.env.PLAYWRIGHT_BASE_URL ?? 'http://localhost:4173',
		trace: 'on-first-retry'
	},
	projects: [
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] }
		}
	],
	webServer: {
		command: 'npm run preview -- --port 4173 --host',
		url: 'http://localhost:4173',
		reuseExistingServer: !process.env.CI
	},
	globalSetup: path.resolve(__dirname, './tests/e2e/global-setup.ts')
});
