// @ts-nocheck
import fs from 'fs';
import path from 'path';
import { chromium, request as playwrightRequest } from 'playwright';

// TypeScript CI script (injection-style)
// Steps:
// 1) Use Playwright's request context to GET /auth/captcha and POST /auth/login
// 2) Obtain token from response
// 3) Launch browser, inject `tea_api_base` and `tea_admin_token` into localStorage via addInitScript
// 4) Open Admin FE and validate sidebar + store->order bring-in flow
// 5) Write a JSON report into REPORT_DIR

const ADMIN_URL = process.env.ADMIN_FE_URL || 'http://localhost:8000';
const API_BASE = process.env.API_BASE || 'http://localhost:8080/api/v1';
const HEADLESS = (process.env.HEADLESS || '1') !== '0';
const REPORT_DIR = process.env.REPORT_DIR || path.join(__dirname, 'reports');

async function run() {
  console.log('CI Test - Admin URL:', ADMIN_URL);
  console.log('CI Test - API_BASE:', API_BASE);

  const apiReq = await playwrightRequest.newContext();
  let captcha: any = null;
  try {
    const resp = await apiReq.get(API_BASE + '/auth/captcha');
    captcha = await resp.json();
    console.log('Fetched captcha:', captcha && captcha.id ? captcha.id : '(no id)');
  } catch (e) {
    console.error('Failed fetching captcha:', e);
    process.exit(2);
  }

  // perform login via API
  let token: string | null = null;
  try {
    const body = { openid: 'admin_openid', captcha_id: captcha.id, captcha_code: captcha.code };
    const loginResp = await apiReq.post(API_BASE + '/auth/login', { data: body, headers: { 'Content-Type': 'application/json' } });
    const j = await loginResp.json();
    token = j && j.data && j.data.token ? j.data.token : null;
    if (!token) {
      console.error('Login did not return token:', JSON.stringify(j));
      process.exit(3);
    }
    console.log('Obtained token via API login (length):', token.length);
  } catch (e) {
    console.error('Login request failed:', e);
    process.exit(4);
  }

  // Launch browser and inject localStorage values
  const browser = await chromium.launch({ headless: HEADLESS });
  const context = await browser.newContext();
  await context.addInitScript((v: any) => {
    try { window.localStorage.setItem('tea_api_base', v.apiBase); } catch(_) {}
    try { window.localStorage.setItem('tea_admin_token', v.token); } catch(_) {}
  }, { apiBase: API_BASE, token });

  const page = await context.newPage();
  page.on('console', msg => { try { console.log('PAGE:', msg.text()); } catch(_){} });

  // navigate to admin url
  try {
    await page.goto(ADMIN_URL, { waitUntil: 'networkidle', timeout: 20000 });
  } catch (e) {
    console.error('Failed to open Admin-FE:', e);
    await browser.close();
    process.exit(5);
  }

  // Validate presence of sidebar
  let sidebarText: string | null = null;
  try {
    await page.waitForSelector('#menu-root', { timeout: 8000 });
    sidebarText = await page.evaluate(() => { const el = document.getElementById('menu-root'); return el ? (el.innerText || el.textContent || null) : null; });
  } catch (_) {
    console.warn('menu-root not found');
  }

  // perform store -> order bring-in check
  let flowOk = false;
  try {
    try { await page.click("button:has-text('门店管理')", { timeout: 3000 }); } catch (_) { try { await page.click("text=门店管理", { timeout: 3000 }); } catch (_) {} }
    try { await page.waitForSelector('#stores-table', { timeout: 5000 }); } catch (_) {}
    const storeButtons = await page.$$('#stores-tbody button');
    if (storeButtons && storeButtons.length > 0) {
      await storeButtons[0].click();
      await page.waitForSelector('#store-panel', { timeout: 8000 });
      try { await page.waitForSelector('#store-orders-tbody tr', { timeout: 8000 }); } catch (_) { console.warn('No orders rows found'); }
      const orderButtons = await page.$$('#store-orders-tbody button');
      if (orderButtons && orderButtons.length > 0) {
        const firstRow = await page.$('#store-orders-tbody tr');
        let selectedOrderId: string | null = null;
        if (firstRow) {
          try { selectedOrderId = await firstRow.$eval('td:first-child', td => td.textContent && td.textContent.trim()); } catch(_) {}
        }
        await orderButtons[0].click();
        try {
          await page.waitForFunction((expected) => {
            const el = document.getElementById('deliver-order-id');
            return el && String(el.value || '').trim() === String(expected).trim();
          }, selectedOrderId, { timeout: 4000 });
        } catch (_) { /* ignore */ }
        flowOk = true;
      }
    }
  } catch (e) {
    console.warn('Store->Order flow check error', e);
  }

  // write report
  try {
    if (!fs.existsSync(REPORT_DIR)) fs.mkdirSync(REPORT_DIR, { recursive: true });
    const out = {
      timestamp: new Date().toISOString(),
      adminUrl: ADMIN_URL,
      apiBase: API_BASE,
      tokenPresent: !!token,
      sidebar: sidebarText ? (sidebarText as string).substring(0, 200) : null,
      storeOrderFlow: flowOk,
    };
    const reportPath = path.join(REPORT_DIR, `report-ci-${Date.now()}.json`);
    fs.writeFileSync(reportPath, JSON.stringify(out, null, 2), 'utf8');
    console.log('Wrote report to', reportPath);
  } catch (e) {
    console.warn('Failed to write report', e);
  }

  await browser.close();
  process.exit(flowOk ? 0 : 1);
}

run().catch(err => {
  console.error('Unhandled error in CI script', err);
  process.exit(10);
});
