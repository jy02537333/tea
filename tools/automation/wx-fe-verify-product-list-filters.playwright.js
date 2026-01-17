#!/usr/bin/env node
// Verify product-list filters layout (left-right) using Playwright
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
  const preview = process.env.PREVIEW_URL || 'http://127.0.0.1:9093';
  const tokenEnv = process.env.TOKEN || process.env.TK || '';
  const base = preview.replace(/#.*$/, '') + '/#/pages/product-list/index';
  const url = tokenEnv ? `${base}?tk=${encodeURIComponent(tokenEnv)}` : base;

  const browser = await chromium.launch({ headless: true, args: ['--no-sandbox', '--disable-setuid-sandbox'] });
  const context = await browser.newContext({ viewport: { width: 420, height: 880 } });
  const page = await context.newPage();
  if (tokenEnv) {
    await page.addInitScript((t) => window.localStorage?.setItem('token', t), tokenEnv);
  }

  page.on('pageerror', (e) => console.error('[pageerror]', e?.message || String(e)));
  page.on('requestfailed', (r) => console.warn('[requestfailed]', r.method(), r.url()));

  await page.goto(url, { waitUntil: 'load' });

  // Wait for key labels
  await page.getByText('选择门店').first().waitFor({ timeout: 15000 }).catch(() => {});
  await page.getByText('筛选与排序').first().waitFor({ timeout: 15000 }).catch(() => {});
  await page.waitForTimeout(500);

  async function rowInfo(labelText) {
    const label = page.getByText(labelText).first();
    const labelBox = await label.boundingBox();
    // the control box is expected to be the next sibling with class row-control
    const control = page.locator('.row-control').first();
    const controlBox = await control.boundingBox();
    return { labelText, labelBox, controlBox };
  }

  const infos = [];
  for (const text of ['选择门店', '产地', '包装', '排序']) {
    try { infos.push(await rowInfo(text)); } catch (e) { infos.push({ labelText: text, error: String(e) }); }
  }

  for (const info of infos) {
    console.log('[row]', info.labelText, 'labelBox=', info.labelBox, 'controlBox=', info.controlBox);
    if (info.labelBox && info.controlBox) {
      const leftRight = info.labelBox.x + info.labelBox.width <= info.controlBox.x + 2;
      console.log('  layout=', leftRight ? 'left-right OK' : 'stacked/NOK');
    }
  }

  await browser.close();
})();
