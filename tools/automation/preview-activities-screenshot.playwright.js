#!/usr/bin/env node
/*
 Playwright-based automation without browser download.
 Attempts to launch a system-installed Chromium/Chrome via executablePath.
 Steps: inject token -> open activities -> fill -> click -> wait success -> screenshot.
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

function readJSON(p) { return JSON.parse(fs.readFileSync(p, 'utf-8')); }

(async () => {
  const repoRoot = path.resolve(__dirname, '../../');
  const logsDir = path.join(repoRoot, 'build-ci-logs');
  const screenshotsDir = path.join(logsDir, 'screenshots');
  if (!fs.existsSync(screenshotsDir)) fs.mkdirSync(screenshotsDir, { recursive: true });

  const previewURL = process.env.PREVIEW_URL || 'http://127.0.0.1:10093/';
  const token = readJSON(path.join(logsDir, 'admin_login_response.json')).data?.token || '';
  const storeId = readJSON(path.join(logsDir, 'activity_demo_create_store.json')).data?.id || 0;
  const activityId = readJSON(path.join(logsDir, 'activity_demo_create_activity.json')).data?.id || 0;
  if (!token || !storeId || !activityId) throw new Error('Missing token/storeId/activityId');

  const name = process.env.ACT_NAME || '张三';
  const phone = process.env.ACT_PHONE || '18000000001';
  const fee = process.env.ACT_FEE || '9.9';

  const executablePath = findBrowserPath();
  const launchOpts = { headless: true, args: ['--no-sandbox', '--disable-setuid-sandbox'] };
  if (executablePath) launchOpts.executablePath = executablePath;

  console.log('Launching Chromium. executablePath=', executablePath || '(bundled/default)');
  const browser = await chromium.launch(launchOpts);
  const context = await browser.newContext({ viewport: { width: 420, height: 880, deviceScaleFactor: 1 } });

  // Pre-inject token before any page loads
  await context.addInitScript((t) => { try { localStorage.setItem('token', t); } catch(_) {} }, token);

  const page = await context.newPage();
  await page.goto(previewURL, { waitUntil: 'load' });

  // Go to activities page via browser history route
  const actURL = new URL(previewURL);
  const base = actURL.toString().replace(/#.*$/, '').replace(/\/$/, '');
  const qName = encodeURIComponent(name);
  const qPhone = encodeURIComponent(phone);
  const qFee = encodeURIComponent(String(fee));
  const target = `${base}/pages/activities/index?store_id=${storeId}&activity_id=${activityId}&name=${qName}&phone=${qPhone}&fee=${qFee}&auto=1`;
  await page.goto(target, { waitUntil: 'load' });

  // Ensure activities page rendered
  try {
    await page.getByText('活动报名').waitFor({ timeout: 20000 });
  } catch (_) {
    // try an alternative anchor
    await page.waitForTimeout(1000);
  }

  async function fillByPlaceholderOrIndex(selectorPlaceholder, index, value) {
    const cssSel = `input[placeholder="${selectorPlaceholder}"]`;
    const hasByPlaceholder = await page.locator(cssSel).count();
    if (hasByPlaceholder > 0) {
      await page.locator(cssSel).first().fill(value);
      return;
    }
    const inputs = page.locator('input');
    const cnt = await inputs.count();
    if (cnt > index) {
      await inputs.nth(index).fill(value);
      return;
    }
    throw new Error(`No input found for ${selectorPlaceholder}`);
  }

  // Fill fields by placeholder
  // With auto=1, the page should already be pre-filled and auto-submitted.
  // Keep a light fallback in case auto injection didn't occur.
  try { await fillByPlaceholderOrIndex('请输入姓名', 0, name); } catch (_) {}
  try { await fillByPlaceholderOrIndex('请输入手机号', 1, phone); } catch (_) {}
  try { await fillByPlaceholderOrIndex('不填则视为免费活动', 2, String(fee)); } catch (_) {}

  // Click first button containing 报名
  await page.getByRole('button', { name: /报名/ }).first().click();

  // Wait for success mark
  await page.getByText('已报名成功').waitFor({ timeout: 60000 });

  const out = path.join(screenshotsDir, 'activities_paid_success.png');
  await page.screenshot({ path: out, fullPage: true });
  console.log('Saved screenshot:', out);

  await browser.close();
})().catch((err) => {
  console.error('[playwright-screenshot] failed:', err);
  process.exit(1);
});
