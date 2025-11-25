const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

const ADMIN_URL = process.env.ADMIN_FE_URL || 'http://localhost:8000';
const API_BASE = process.env.API_BASE || 'http://localhost:8080/api/v1';
const HEADLESS = (process.env.HEADLESS || '1') !== '0';

(async () => {
  console.log('Admin FE URL:', ADMIN_URL);
  console.log('API BASE:', API_BASE);
  // Step 1: fetch captcha from backend
  let captcha = null;
  try {
    const resp = await fetch(API_BASE.replace(/\/api\/v1$/, '') + '/api/v1/auth/captcha');
    captcha = await resp.json();
    console.log('Got captcha', captcha);
  } catch (e) {
    console.error('Failed to fetch captcha:', e.message || e);
    process.exit(2);
  }

  // Step 2: call login
  let token = null;
  try {
    const body = { openid: 'admin_openid', captcha_id: captcha.id, captcha_code: captcha.code };
    const resp = await fetch(API_BASE.replace(/\/api\/v1$/, '') + '/api/v1/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
    const j = await resp.json();
    token = (j && j.data && j.data.token) ? j.data.token : null;
    console.log('Login response token present?', !!token);
    if (!token) {
      console.error('Login did not return token', JSON.stringify(j));
      process.exit(3);
    }
  } catch (e) {
    console.error('Login request failed:', e.message || e);
    process.exit(4);
  }

  // Step 3: launch browser and set localStorage before page load
  const browser = await chromium.launch({ headless: HEADLESS });
  const context = await browser.newContext();
  await context.addInitScript((t) => {
    try { window.localStorage.setItem('tea_api_base', t.apiBase); } catch (_) {}
    try { window.localStorage.setItem('tea_admin_token', t.token); } catch (_) {}
  }, { apiBase: API_BASE, token });

  const page = await context.newPage();
  // Capture page console and errors
  page.on('console', msg => { try { console.log('PAGE LOG:', msg.text()); } catch(_){} });
  page.on('pageerror', err => { console.log('PAGE ERROR:', err && err.message ? err.message : err); });

  // Capture network requests and responses to help debug 400 errors
  page.on('request', req => {
    try { console.log('REQ ', req.method(), req.url()); } catch (_) {}
  });
  page.on('response', res => {
    try {
      const status = res.status();
      if (status >= 400) {
        console.log('RES ', status, res.url());
        // attempt to print small response body for debugging
        res.text().then(t => {
          if (t && t.length > 0) {
            const snippet = t.length > 1000 ? t.slice(0,1000) + '... (truncated)' : t;
            console.log('RESP BODY:', snippet);
          }
        }).catch(() => {});
      }
    } catch (_) {}
  });

  try {
    await page.goto(ADMIN_URL, { waitUntil: 'networkidle', timeout: 15000 });
  } catch (e) {
    console.error('Failed to open Admin-FE at', ADMIN_URL, e.message);
    await browser.close();
    process.exit(5);
  }

  // Wait for token-driven UI to appear (menu-root or deliver-order-id)
  try {
    await page.waitForFunction(() => !!window.localStorage.getItem('tea_admin_token'), { timeout: 10000 });
    console.log('Token present in localStorage');
  } catch (e) {
    console.warn('Token not present in localStorage after injection');
  }

  // Check sidebar/menu
  const menuText = await page.evaluate(() => {
    const el = document.getElementById('menu-root');
    return el ? (el.innerText || el.textContent) : null;
  });
  console.log('Sidebar snapshot (first 300 chars):', menuText ? menuText.substring(0,300) : '(none)');

  // Perform store -> order -> bring-in check (reuse logic from original script)
  let flowOk = false;
  try {
    try { await page.click("button:has-text('门店管理')", { timeout: 3000 }); } catch (e) { try { await page.click("text=门店管理", { timeout: 3000 }); } catch (_) {} }
    try { await page.waitForSelector('#stores-table', { timeout: 5000 }); } catch (_) {}
    const storeButtons = await page.$$('#stores-tbody button');
    if (storeButtons && storeButtons.length > 0) {
      await storeButtons[0].click();
      console.log('Clicked first store select button');
      await page.waitForSelector('#store-panel', { timeout: 8000 });
      // give some extra time for async queries to complete
      try { await page.waitForSelector('#store-orders-tbody tr', { timeout: 10000 }); } catch (_) { console.warn('No orders rows found'); }
      const orderButtons = await page.$$('#store-orders-tbody button');
      if (orderButtons && orderButtons.length > 0) {
        const firstRow = await page.$('#store-orders-tbody tr');
        let selectedOrderId = null;
        if (firstRow) {
          try { selectedOrderId = await firstRow.$eval('td:first-child', td => td.textContent && td.textContent.trim()); } catch (_) {}
        }
        await orderButtons[0].click();
        console.log('Clicked first order bring-in button');
        try {
          await page.waitForFunction((expected) => {
            const el = document.getElementById('deliver-order-id');
            return el && String(el.value || '').trim() === String(expected).trim();
          }, selectedOrderId, { timeout: 4000 });
        } catch (_) { console.warn('deliver-order-id did not match'); }
        flowOk = true;
      } else {
        console.warn('No order buttons found');
      }
    } else {
      console.warn('No stores select buttons found');
    }
  } catch (e) {
    console.warn('Store->Order flow check error:', e && e.message ? e.message : e);
  }

  // write short report
  try {
    const out = {
      timestamp: new Date().toISOString(),
      adminUrl: ADMIN_URL,
      apiBase: API_BASE,
      tokenPresent: !!token,
      sidebar: menuText ? menuText.substring(0,500) : null,
      storeOrderFlow: flowOk
    };
    const reportPath = path.join(__dirname, `report-inject-${Date.now()}.json`);
    fs.writeFileSync(reportPath, JSON.stringify(out, null, 2), 'utf8');
    console.log('Wrote report to', reportPath);
  } catch (e) {
    console.warn('Failed to write report:', e.message || e);
  }

  await browser.close();
  process.exit(flowOk ? 0 : 0);
})();
