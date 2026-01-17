#!/usr/bin/env node
/*
 Playwright screenshot for wx-fe login page.
 Uses system Chromium if available; otherwise falls back to bundled.
 - PREVIEW_URL: base URL (default http://127.0.0.1:10088/)
 - OUTPUT: filename under build-ci-logs/screenshots (default wx-fe-login.png)
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

(async () => {
  const repoRoot = path.resolve(__dirname, '../../');
  const logsDir = path.join(repoRoot, 'build-ci-logs');
  const screenshotsDir = path.join(logsDir, 'screenshots');
  if (!fs.existsSync(screenshotsDir)) fs.mkdirSync(screenshotsDir, { recursive: true });

  const baseUrl = (process.env.PREVIEW_URL || 'http://127.0.0.1:10088/').replace(/\/$/, '');
  const output = process.env.OUTPUT || 'wx-fe-login.png';

  const executablePath = findBrowserPath();
  const launchOpts = { headless: true, args: ['--no-sandbox', '--disable-setuid-sandbox'] };
  if (executablePath) launchOpts.executablePath = executablePath;

  console.log('Launching Chromium. executablePath=', executablePath || '(bundled/default)');
  const browser = await chromium.launch(launchOpts);
  const context = await browser.newContext({ viewport: { width: 420, height: 880, deviceScaleFactor: 1 } });
  const page = await context.newPage();

  // Prefer hash route to bypass server-side fallback requirements
  const hashUrl = `${baseUrl}#/pages/login/index`;
  const browserUrl = `${baseUrl}/pages/login/index`;
  const indexHtmlHashUrl = `${baseUrl}/index.html#/pages/login/index`;

  // Try hash first
  try {
    await page.goto(hashUrl, { waitUntil: 'load' });
  } catch (_) {
    await page.goto(baseUrl + '/', { waitUntil: 'load' });
  }

  // Ensure render by looking for login markers
  const markers = [
    '微信一键登录',
    '使用微信授权登录',
    '开发者 OpenID 登录',
    '请输入 openid',
  ];

  let ok = false;
  for (const m of markers) {
    try {
      await page.getByText(m).first().waitFor({ timeout: 8000 });
      ok = true;
      break;
    } catch (_) {}
  }

  // Fallback to browser route if hash didn’t render markers
  if (!ok) {
    try {
      await page.goto(browserUrl, { waitUntil: 'load' });
      for (const m of markers) {
        try { await page.getByText(m).first().waitFor({ timeout: 5000 }); ok = true; break; } catch (_) {}
      }
    } catch (e) {
      console.warn('Fallback browser route failed:', e.message);
    }
  }

  // Final fallback: explicitly load index.html with hash
  if (!ok) {
    try {
      await page.goto(indexHtmlHashUrl, { waitUntil: 'load' });
      for (const m of markers) {
        try { await page.getByText(m).first().waitFor({ timeout: 5000 }); ok = true; break; } catch (_) {}
      }
    } catch (e) {
      console.warn('Fallback index.html# route failed:', e.message);
    }
  }

  const outPath = path.join(screenshotsDir, output);
  await page.screenshot({ path: outPath, fullPage: true });
  console.log('Saved screenshot:', outPath, 'visible markers ok =', ok);

  await browser.close();

  if (!ok) {
    console.warn('Markers were not detected; the page may not have rendered fully.');
    process.exitCode = 2;
  }
})().catch((err) => {
  console.error('[wx-fe-screenshot-login] failed:', err);
  process.exit(1);
});
