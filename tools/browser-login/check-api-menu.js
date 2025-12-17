const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

(async () => {
  const ADMIN_FE_URL = process.env.ADMIN_FE_URL || 'http://localhost:8000/index.html';
  const REPORT_DIR = process.env.REPORT_DIR || path.join(__dirname, 'reports');
  const HEADLESS = (process.env.HEADLESS || '1') !== '0';
  const API_BASE = process.env.API_BASE || 'http://localhost:9292/api/v1';

  if (!fs.existsSync(REPORT_DIR)) fs.mkdirSync(REPORT_DIR, { recursive: true });

  // Try to obtain an admin token via API (login inject) so the UI will render menus
  let injectedToken = null;
  try {
    const capResp = await fetch(API_BASE.replace(/\/api\/v1$/, '') + '/api/v1/auth/captcha');
    const capJson = await capResp.json();
    if (capJson && capJson.id) {
      const body = { openid: 'admin_openid', captcha_id: capJson.id, captcha_code: capJson.code };
      const loginResp = await fetch(API_BASE.replace(/\/api\/v1$/, '') + '/api/v1/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
      const lj = await loginResp.json();
      injectedToken = lj && lj.data && lj.data.token ? lj.data.token : null;
      console.log('Injected token obtained?', !!injectedToken);
    }
  } catch (e) {
    console.warn('Token injection attempt failed:', e && e.message ? e.message : e);
  }

  const browser = await chromium.launch({ headless: HEADLESS });
  const context = await browser.newContext();
  if (injectedToken) {
    await context.addInitScript((t) => {
      try { window.localStorage.setItem('tea_api_base', t.apiBase); } catch(_){}
      try { window.localStorage.setItem('tea_admin_token', t.token); } catch(_){}
      try { window.localStorage.setItem('admin_role', '9'); } catch(_){}
    }, { apiBase: API_BASE, token: injectedToken });
  }
  const page = await context.newPage();

  const out = { timestamp: new Date().toISOString(), url: ADMIN_FE_URL };
  try {
    await page.goto(ADMIN_FE_URL, { waitUntil: 'networkidle', timeout: 30000 });
    // allow scripts to run and menus to populate
    await page.waitForTimeout(1200);

    // Try to force menu rendering if page scripts provide helpers
    try {
      await page.evaluate(() => {
        try { if (typeof fetchAndRenderMenus === 'function') fetchAndRenderMenus(); } catch(_) {}
        try { if (typeof renderSidebarMenu === 'function') renderSidebarMenu(); } catch(_) {}
      });
    } catch(_) {}

    // wait a bit for menu to be populated
    await page.waitForTimeout(800);

    // diagnostic: collect debug info about menu rendering helpers and DOM
    const diag = await page.evaluate(() => {
      const root = document.getElementById('menu-root');
      return {
        hasMenuRoot: !!root,
        menuRootHTML: root ? root.innerHTML : null,
        sidebarMenuType: typeof SIDEBAR_MENU,
        sidebarMenuLen: (typeof SIDEBAR_MENU !== 'undefined' && Array.isArray(SIDEBAR_MENU)) ? SIDEBAR_MENU.length : null,
        hasRenderSidebarMenu: typeof renderSidebarMenu === 'function',
        hasFetchAndRenderMenus: typeof fetchAndRenderMenus === 'function',
        containerDisplay: (document.querySelector('.container') || {}).style ? (document.querySelector('.container').style.display || '') : ''
      };
    });
    out.diag = diag;

    const items = await page.$$eval('#menu-root .menu-item', els => els.map(e => ({ text: (e.textContent||'').trim(), tab: e.dataset ? e.dataset.tab : null })));
    out.items = items;
    const found = items.some(i => (i.tab === 'testing') || /API|api/.test(i.text));
    out.found = !!found;

    const file = path.join(REPORT_DIR, `check-api-menu-${Date.now()}.json`);
    fs.writeFileSync(file, JSON.stringify(out, null, 2), 'utf8');
    console.log('Wrote report:', file, 'found=', out.found);
    await browser.close();
    process.exit(found ? 0 : 2);
  } catch (err) {
    out.error = err && err.message ? err.message : String(err);
    const file = path.join(REPORT_DIR, `check-api-menu-error-${Date.now()}.json`);
    try { fs.writeFileSync(file, JSON.stringify(out, null, 2), 'utf8'); } catch(_){}
    console.error('Error during check, report:', file, err);
    try { await browser.close(); } catch(_){}
    process.exit(3);
  }
})();
