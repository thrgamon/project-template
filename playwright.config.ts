import { defineConfig } from '@playwright/test';

export default defineConfig({
	testDir: './e2e',
	fullyParallel: true,
	forbidOnly: !!process.env.CI,
	retries: 1,
	workers: process.env.CI ? 2 : 3,
	reporter: 'list',
	timeout: 45000,
	use: {
		baseURL: 'http://localhost:3000',
		trace: 'on-first-retry',
		actionTimeout: 10000,
	},
	projects: [
		{
			name: 'chromium',
			use: { browserName: 'chromium' },
		},
	],
});
