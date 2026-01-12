#!/usr/bin/env node
// Capture product-list filters block screenshot via Playwright
const fs = require('fs');
const path = require('path');

function requirePlaywright() {
  try { return require('playwright'); } catch (_) {}
  try { return require(path.resolve(__dirname, '../../admin-fe/node_modules/playwright')); } catch (e) {
    throw new Error('playwright not found. Run pnpm install at repo root.');
  }
}

(async () => {
  const { chromium } = requirePlaywright();
  const repoRoot = path.resolve(__dirname, '../../');
  const logsDir = path.join(repoRoot, 'build-ci-logs');
  const shotsDir = path.join(logsDir, 'screenshots');
  if (!fs.existsSync(shotsDir)) fs.mkdirSync(shotsDir, { recursive: true });
  const outPath = path.join(shotsDir, 'product-list-filters.png');

  const preview = process.env.PREVIEW_URL || 'http://127.0.0.1:9093';
  const url = preview.replace(/#.*$/, '') + '/#/pages/product-list/index';

  const browser = await chromium.launch({ headless: true, args: ['--no-sandbox', '--disable-setuid-sandbox'] });
  const context = await browser.newContext({ viewport: { width: 420, height: 880, deviceScaleFactor: 1 } });
  const page = await context.newPage();

  // Basic console logging for diagnostics
  page.on('pageerror', (e) => console.error('[pageerror]', e?.message || String(e)));
  page.on('requestfailed', (r) => console.warn('[requestfailed]', r.method(), r.url()));

  await page.goto(url, { waitUntil: 'load' });

  // Wait briefly for UI to stabilize; then capture full page for robustness
  await page.waitForTimeout(1500);
  await page.screenshot({ path: outPath, fullPage: true });

  await browser.close();
  console.log('[screenshot] saved:', outPath);
})();
