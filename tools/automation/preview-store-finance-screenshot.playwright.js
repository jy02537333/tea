#!/usr/bin/env node
/*
 Playwright-based screenshot for wx-fe store finance page.
 Steps: launch local static server (external), inject token + current_store_id, navigate to finance page, wait for header, capture screenshot.
 Env:
  - PREVIEW_URL: base URL for static server (default http://127.0.0.1:10114/)
  - FINANCE_STORE_ID: store id to open (default 1)
*/
const fs = require('fs');
const path = require('path');
const { chromium } = require('playwright');

function findBrowserPath() {
  const envPaths = [
    process.env.PLAYWRIGHT_EXECUTABLE_PATH,
    process.env.CHROMIUM_BIN,
    process.env.CHROME_BIN,
    process.env.GOOGLE_CHROME_BIN,
  ].filter(Boolean);
  const candidates = envPaths.concat([
    '/usr/bin/chromium-browser',
    '/usr/bin/chromium',
    '/usr/bin/google-chrome',
    '/usr/bin/google-chrome-stable',
    '/snap/bin/chromium',
  ]);
  for (const p of candidates) {
    try { if (p && fs.existsSync(p)) return p; } catch (_) {}
  }
  return null;
}

function readJSONSafe(p, fallback = null) {
  try { return JSON.parse(fs.readFileSync(p, 'utf-8')); } catch (_) { return fallback; }
}

(async () => {
  const repoRoot = path.resolve(__dirname, '../../');
  const logsDir = path.join(repoRoot, 'build-ci-logs');
  const screenshotsDir = path.join(logsDir, 'screenshots');
  if (!fs.existsSync(screenshotsDir)) fs.mkdirSync(screenshotsDir, { recursive: true });

  const previewURL = process.env.PREVIEW_URL || 'http://127.0.0.1:10114/';
  const storeIdEnv = process.env.FINANCE_STORE_ID;
  const storeId = Number(storeIdEnv || 1);
  const token = readJSONSafe(path.join(logsDir, 'admin_login_response.json'), {}).data?.token || '';

  const executablePath = findBrowserPath();
  const launchOpts = { headless: true, args: ['--no-sandbox', '--disable-setuid-sandbox'] };
  if (executablePath) launchOpts.executablePath = executablePath;

  console.log('Launching Chromium. executablePath=', executablePath || '(bundled/default)');
  const browser = await chromium.launch(launchOpts);
  const context = await browser.newContext({ viewport: { width: 420, height: 880, deviceScaleFactor: 1 } });

  // Pre-inject token and current_store_id before any page loads
  await context.addInitScript((t, sid) => {
    try { localStorage.setItem('token', t); } catch(_) {}
    try { localStorage.setItem('current_store_id', String(sid)); } catch(_) {}
  }, token, storeId);

  const page = await context.newPage();
  await page.goto(previewURL, { waitUntil: 'load' });

  const base = previewURL.replace(/#.*$/, '').replace(/\/$/, '');
  const target = `${base}/pages/store-finance/index?store_id=${storeId}`;
  console.log('Navigating to', target);
  await page.goto(target, { waitUntil: 'load' });

  try { await page.getByText('门店财务流水').waitFor({ timeout: 15000 }); } catch (_) {}
  try { await page.getByRole('button', { name: '返回门店详情' }).waitFor({ timeout: 5000 }); } catch (_) {}
  try { await page.getByRole('button', { name: '回到门店列表' }).first().waitFor({ timeout: 5000 }); } catch (_) {}

  const out = path.join(screenshotsDir, 'wx_store_finance.png');
  await page.screenshot({ path: out, fullPage: true });
  console.log('Saved screenshot:', out);

  await browser.close();
})().catch((err) => {
  console.error('[playwright-store-finance] failed:', err);
  process.exit(1);
});
