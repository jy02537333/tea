// login-test-pure-ui.js
// Pure UI Playwright script (copy of original) — performs login via the page UI and writes a report.
// Usage:
// 1) npm install
// 2) npx playwright install
// 3) ADMIN_FE_URL=http://localhost:8000 API_BASE=http://localhost:8080/api/v1 node login-test-pure-ui.js

const { chromium } = require('playwright');

const ADMIN_URL = process.env.ADMIN_FE_URL || 'http://localhost:8000';
const USERNAME = process.env.TEST_USER || 'admin';
const PASSWORD = process.env.TEST_PASS || 'pass';
const HEADLESS = (process.env.HEADLESS || '1') !== '0';
const API_BASE = process.env.API_BASE || (typeof window === 'undefined' ? 'http://localhost:8080/api/v1' : null);
const REPORT_DIR = process.env.REPORT_DIR || `${__dirname}`;

(async () => {
  console.log('Admin FE URL:', ADMIN_URL);
  const browser = await chromium.launch({ headless: HEADLESS });
  const context = await browser.newContext();
  const page = await context.newPage();
  // Capture page console and errors for debugging
  page.on('console', msg => {
    try { console.log('PAGE LOG:', msg.text()); } catch(_) {}
  });
  page.on('pageerror', err => { console.log('PAGE ERROR:', err && err.message ? err.message : err); });
  // Ensure Admin-FE uses our local API server for tests (if API_BASE is set)
  try {
    const apiToSet = process.env.API_BASE || 'http://localhost:8080/api/v1';
    await context.addInitScript((v) => { try { localStorage.setItem('tea_api_base', v); } catch(_) {} }, apiToSet);
  } catch (_) {}

    try {
      // Navigate directly to the standalone login page, with next param to return to index
      const loginUrl = ADMIN_URL.replace(/\/$/, '') + '/login.html?next=index.html';
      console.log('Navigating to', loginUrl);
      await page.goto(loginUrl, { waitUntil: 'networkidle', timeout: 20000 });
      // Ensure login fields are present
      await page.waitForSelector('#login-username', { timeout: 15000 });
    } catch (e) {
      console.error('Failed to open Admin-FE login page at', ADMIN_URL, e.message);
      await browser.close();
      process.exit(2);
    }

  // Try to open login modal by clicking login button if present
  try {
    const loginBtn = await page.$('#login-btn');
    if (loginBtn) {
      await loginBtn.click();
      console.log('Clicked #login-btn');
    } else {
      console.log('#login-btn not found; assuming login modal may already be visible');
    }
  } catch (e) {
    console.warn('Error clicking login button:', e.message);
  }

  // Fill username/password fields
  try {
    await page.fill('#login-username', USERNAME);
    await page.fill('#login-password', PASSWORD);
  } catch (e) {
    console.error('Login form fields not found:', e.message);
  }
  // Read captcha displayed on page (development helper) and fill it if present.
  // If missing, fetch captcha from backend using page context and populate fields (keeps pure-UI flow).
  try {
    // allow some time for captcha code display to populate
    try { await page.waitForSelector('#captcha-code-display', { timeout: 3000 }); } catch (_) {}
    // read the captcha displayed on page (if any)
    let captchaText = await page.evaluate(() => {
      const el = document.getElementById('captcha-code-display');
      return el ? (el.innerText || el.textContent || '') : null;
    });
    if (!captchaText || String(captchaText).trim() === '' || captchaText.trim() === '-') {
      // fetch captcha through the page (so CORS uses same origin and stored API_BASE)
      try {
        await page.evaluate(async () => {
          const base = (localStorage.getItem('tea_api_base') || 'http://localhost:8080/api/v1').replace(/\/api\/v1\/?$/, '');
          const resp = await fetch(base + '/api/v1/auth/captcha');
          const j = await resp.json();
          if (j && j.id) {
            document.getElementById('captcha-id').value = j.id;
            document.getElementById('captcha-code-display').textContent = j.code || '';
            document.getElementById('captcha-input').value = j.code || '';
          }
        });
        captchaText = await page.evaluate(() => {
          const el = document.getElementById('captcha-code-display');
          return el ? (el.innerText || el.textContent || '') : null;
        });
        console.log('Fetched captcha via API and filled page:', captchaText && captchaText.trim());
      } catch (e) {
        console.warn('Failed to fetch captcha via page context:', e.message || e);
      }
    } else {
      try { await page.fill('#captcha-input', captchaText.trim()); } catch (e) { console.warn('Failed to fill #captcha-input:', e.message); }
      try { await page.fill('#captcha-id', (document.getElementById('captcha-id') && document.getElementById('captcha-id').value) || ''); } catch(_) {}
      console.log('Filled captcha from page display:', captchaText.trim());
    }
  } catch (e) {
    console.warn('Error ensuring captcha on page:', e.message);
  }

  // Click the login submit button. Several sites use different selectors; try multiple fallbacks.
  const clickSelectors = [
    "#login-modal .btn.btn-add",
    "#login-modal button.btn-add",
    "#login-modal button:has-text('登录')",
    "button.btn-add:has-text('登录')",
    "button:has-text('登录')"
  ];

  let clicked = false;
  for (const sel of clickSelectors) {
    try {
      const el = await page.$(sel);
      if (el) {
        await el.click();
        console.log('Clicked login button using selector:', sel);
        clicked = true;
        break;
      }
    } catch (e) {
      // ignore and try next
    }
  }
  // Also try direct #do-login button on standalone page with retry
  if (!clicked) {
    for (let attempt = 0; attempt < 3 && !clicked; attempt++) {
      try { const direct = await page.$('#do-login'); if (direct) { await direct.click(); console.log('Clicked #do-login'); clicked = true; break; } } catch (e) { console.warn('click #do-login attempt', attempt, e && e.message); }
      await page.waitForTimeout(500);
    }
  }
  if (!clicked) console.warn('Could not find a known login button selector; the page may submit automatically or require different flow');

  // Attach network listeners to capture XHR/fetch and failures
  const netLogs = [];
  page.on('request', req => {
    const r = { id: req._guid || req.url(), url: req.url(), method: req.method(), headers: req.headers(), ts: Date.now() };
    netLogs.push({ type: 'request', data: r });
  });
  page.on('response', async res => {
    try {
      const url = res.url();
      const status = res.status();
      const headers = res.headers();
      let body = null;
      // try to get JSON/text for /auth/login or /admin/menus
      if (url.includes('/auth/login') || url.includes('/admin/menus')) {
        try { body = await res.text(); } catch (e) { body = '<binary or unavailable>'; }
      }
      netLogs.push({ type: 'response', data: { url, status, headers, body, ts: Date.now() } });
    } catch (e) {
      // ignore
    }
  });
  page.on('requestfailed', rf => {
    netLogs.push({ type: 'requestfailed', data: { url: rf.url(), errorText: rf.failure() ? rf.failure().errorText : 'unknown', ts: Date.now() } });
  });

  // Wait for auth/login response specifically and retry clicking login up to 3 times if needed
  let loginResponse = null;
  for (let attempt = 1; attempt <= 5; attempt++) {
    try {
      loginResponse = await page.waitForResponse(r => r.url().includes('/auth/login'), { timeout: 20000 });
      console.log('Detected /auth/login response: status=', loginResponse.status());
      break;
    } catch (e) {
      console.warn(`Attempt ${attempt}: no /auth/login response within timeout.`);
      // retry clicking the submit button if present
      for (const sel of clickSelectors) {
        try {
          const el = await page.$(sel);
          if (el) {
            await el.click();
            console.log('Retry clicked login using selector:', sel);
            break;
          }
        } catch (_) {}
      }
      // give short pause between attempts to allow UI/backend to recover
      try { await page.waitForTimeout(1500); } catch (_) {}
      // on last attempt, give up
      if (attempt === 5) break;
    }
  }

  // After observing response (or exhausting retries), wait for token in localStorage up to 20s
  try {
    await page.waitForFunction(() => !!window.localStorage.getItem('tea_admin_token'), { timeout: 20000 });
    const token = await page.evaluate(() => window.localStorage.getItem('tea_admin_token'));
    console.log('\n=== LOGIN SUCCESS: token saved in localStorage.tea_admin_token ===\n');
    console.log(token);

    const menuHtml = await page.evaluate(() => {
      const root = document.getElementById('menu-root');
      if (root) return root.innerText || root.textContent || null;
      return null;
    });
    console.log('\nSidebar menu (text):', menuHtml ? menuHtml.substring(0, 800) : '(no menu-root found)');

    // === 新增：门店 -> 订单 -> 带入并查看 流程验证 ===
    let selectedStoreId = null;
    let selectedOrderId = null;
    try {
      // 切换到 门店管理 标签
      try { await page.click("button:has-text('门店管理')", { timeout: 5000 }); } catch (e) { try { await page.click("text=门店管理", { timeout: 5000 }); } catch (_) {} }
      // 等待门店表格加载
      try { await page.waitForSelector('#stores-table', { timeout: 8000 }); } catch (_) { /* ignore */ }

      // 选择第一个门店（点击第一行的 '选择' 按钮）
      const storeButtons = await page.$$('#stores-tbody button');
      if (storeButtons && storeButtons.length > 0) {
        await storeButtons[0].click();
        console.log('Clicked first store select button');
      } else {
        console.warn('No store select buttons found');
      }

        // 等待门店面板显示并加载订单列表
        await page.waitForSelector('#store-panel', { timeout: 10000 });
        // 等待至少一条订单行出现（后端 mock 可能返回空）
        try {
          await page.waitForSelector('#store-orders-tbody tr', { timeout: 10000 });
        } catch (_) {
          console.warn('No orders rows found within timeout');
        }

      // 如果有订单行，点击第一个行内的“带入并查看”按钮
      const orderButtons = await page.$$('#store-orders-tbody button');
      if (orderButtons && orderButtons.length > 0) {
        // 在点击前读取第一行的订单ID（列 1）
        const firstRow = await page.$('#store-orders-tbody tr');
        if (firstRow) {
          const idText = await firstRow.$eval('td:first-child', td => td.textContent && td.textContent.trim());
          selectedOrderId = idText || null;
        }
        await orderButtons[0].click();
        console.log('Clicked first order "带入并查看"');
        // 等待 deliver-order-id 填充为选中订单ID
        try {
          await page.waitForFunction((expected) => {
            const el = document.getElementById('deliver-order-id');
            return el && String(el.value || '').trim() === String(expected).trim();
          }, selectedOrderId, { timeout: 4000 });
        } catch (_) {
          console.warn('deliver-order-id did not match expected within timeout');
        }

        // 等待右侧订单详情区包含订单 ID 或订单号文本
        try {
          await page.waitForFunction((oid) => {
            const box = document.getElementById('order-inspect');
            if (!box) return false;
            const txt = (box.innerText || box.textContent || '').toString();
            if (!txt) return false;
            if (oid && oid !== 'null') return txt.includes(String(oid));
            return /订单 #[0-9]+/.test(txt) || /订单号/.test(txt);
          }, selectedOrderId, { timeout: 5000 });
        } catch (_) {
          console.warn('Order inspect did not render expected content within timeout');
        }
        } else {
          console.warn('No "带入并查看" buttons found in orders table');
        }
    } catch (err) {
      console.warn('Store->Order flow check failed:', err && err.message ? err.message : err);
    }
  } catch (e) {
    console.error('Timed out waiting for tea_admin_token in localStorage:', e && e.message);
    const keys = await page.evaluate(() => Object.keys(window.localStorage));
    console.log('localStorage keys:', keys);
    console.log('\n=== Collected network logs (recent 80 entries) ===');
    const recent = netLogs.slice(-80);
    for (const entry of recent) {
      if (entry.type === 'request') console.log('[REQ] ', entry.data.method, entry.data.url);
      else if (entry.type === 'response') console.log('[RES] ', entry.data.status, entry.data.url);
      else if (entry.type === 'requestfailed') console.log('[FAIL]', entry.data.errorText, entry.data.url);
    }
  }

  // write a small JSON report to disk for CI/manual review
  try {
    const fs = require('fs');
    const path = require('path');
    // Compose richer report with network logs and pass/fail
    const savedToken = (await page.evaluate(() => window.localStorage.getItem('tea_admin_token'))) || null;
    const sidebarText = await page.evaluate(() => { const r = document.getElementById('menu-root'); return r ? (r.innerText||r.textContent) : null });
    const success = !!savedToken;
    const out = {
      timestamp: new Date().toISOString(),
      adminUrl: ADMIN_URL,
      username: USERNAME,
      token: savedToken,
      sidebarText: sidebarText,
      success: success,
      networkLogs: netLogs.slice(-200),
    };
    const reportName = `report-${Date.now()}.json`;
    const reportPath = path.join(REPORT_DIR, reportName);
    fs.writeFileSync(reportPath, JSON.stringify(out, null, 2), { encoding: 'utf8' });
    console.log('Wrote report to', reportPath);
    // exit with non-zero code on failure to help CI
    await browser.close();
    process.exit(success ? 0 : 1);
  } catch (e) {
    console.warn('Failed to write report:', e.message);
  }

})();
