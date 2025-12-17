const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

(async () => {
  const ADMIN_FE_URL = process.env.ADMIN_FE_URL || 'http://localhost:8000/index.html';
  const API_BASE = process.env.API_BASE || 'http://localhost:9292/api/v1';
  const REPORT_DIR = process.env.REPORT_DIR || path.join(__dirname, 'reports');
  const HEADLESS = (process.env.HEADLESS || '1') !== '0';

  if (!fs.existsSync(REPORT_DIR)) fs.mkdirSync(REPORT_DIR, { recursive: true });

  const browser = await chromium.launch({ headless: HEADLESS });
  const context = await browser.newContext();

  // try to get token via API to force full menu render
  try {
    const capResp = await fetch(API_BASE.replace(/\/api\/v1$/, '') + '/api/v1/auth/captcha');
    const capJson = await capResp.json();
    if (capJson && capJson.id) {
      const body = { openid: 'admin_openid', captcha_id: capJson.id, captcha_code: capJson.code };
      const loginResp = await fetch(API_BASE.replace(/\/api\/v1$/, '') + '/api/v1/auth/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
      const lj = await loginResp.json();
      const token = lj && lj.data && lj.data.token ? lj.data.token : null;
      if (token) {
        await context.addInitScript((t) => { try { window.localStorage.setItem('tea_admin_token', t.token); window.localStorage.setItem('tea_api_base', t.apiBase); } catch(_){} }, { token, apiBase: API_BASE });
      }
    }
  } catch (e) { /* ignore */ }

  const page = await context.newPage();
  await page.goto(ADMIN_FE_URL, { waitUntil: 'networkidle', timeout: 30000 });
  await page.waitForTimeout(1200);
  try { await page.evaluate(() => { try { if (typeof fetchAndRenderMenus === 'function') fetchAndRenderMenus(); if (typeof renderSidebarMenu === 'function') renderSidebarMenu(); } catch(_){} }); } catch(_){}
  await page.waitForTimeout(800);
  // If there's a testing menu item, click it to activate the API 测试管理 view
  try {
    await page.evaluate(() => {
      try {
        const el = Array.from((document.getElementById('menu-root')||{querySelectorAll:()=>[]}).querySelectorAll('.menu-item')).find(e => (e.dataset && e.dataset.tab) === 'testing');
        if (el) el.click();
      } catch(_){}
    });
    await page.waitForTimeout(600);
  } catch(_){}

  const result = await page.evaluate(() => {
    const root = document.getElementById('menu-root');
    const first = root ? root.querySelector('.menu-item') : null;
    const all = root ? Array.from(root.querySelectorAll('.menu-item')).map(e=>({tab: e.dataset.tab, text: (e.textContent||'').trim()})) : [];
    const rect = first ? first.getBoundingClientRect() : null;
    const style = first ? window.getComputedStyle(first) : null;
    return { hasRoot: !!root, items: all, firstRect: rect ? { x: rect.x, y: rect.y, width: rect.width, height: rect.height } : null, firstStyle: style ? { color: style.color, background: style.backgroundColor, display: style.display, visibility: style.visibility, opacity: style.opacity } : null };
  });

  // screenshot the left sidebar
  const sidebar = await page.$('#left-sidebar');
  const screenshotPath = path.join(REPORT_DIR, `menu-inspect-${Date.now()}.png`);
  if (sidebar) await sidebar.screenshot({ path: screenshotPath });

  const out = { timestamp: new Date().toISOString(), url: ADMIN_FE_URL, diag: result, screenshot: path.basename(screenshotPath) };
  const outFile = path.join(REPORT_DIR, `menu-inspect-${Date.now()}.json`);
  fs.writeFileSync(outFile, JSON.stringify(out, null, 2), 'utf8');
  console.log('Wrote inspect report:', outFile);
  console.log('Screenshot:', screenshotPath);

  await browser.close();
})();
