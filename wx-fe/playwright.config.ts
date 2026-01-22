import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 30_000,
  retries: 0,
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:10086',
    headless: true
  },
  webServer: {
    command: 'pnpm run dev:h5',
    url: 'http://localhost:10086',
    reuseExistingServer: false,
    timeout: 120_000
  }
});
