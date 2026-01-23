import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: '.',
  timeout: 30000,
  fullyParallel: true,
  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:8080',
    trace: 'retain-on-failure',
  },
  projects: [
    {
      name: 'api-tests',
      use: { ...devices['Desktop Chrome'] }
    }
  ],
});
