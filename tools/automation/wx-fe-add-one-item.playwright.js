// Add one product to cart from a given storeId (or platform when omitted)
// Usage:
//   node tools/automation/wx-fe-add-one-item.playwright.js http://127.0.0.1:9093 STORE_ID=413 TOKEN=xxx

const { chromium } = require('playwright');

function parseArgs(argv) {
  const args = { base: 'http://127.0.0.1:9093', storeId: undefined, token: undefined };
  if (argv[2] && /^https?:/.test(argv[2])) args.base = argv[2];
  for (const a of argv.slice(2)) {
    if (/^STORE_ID=/.test(a)) args.storeId = a.split('=')[1];
    if (/^TOKEN=/.test(a)) args.token = a.split('=')[1];
  }
  return args;
}

async function main() {
  const { base, storeId, token } = parseArgs(process.argv);
  const urlBase = base.replace(/\/$/, '');
  const qs = [];
  if (storeId) qs.push(`store_id=${encodeURIComponent(storeId)}`);
  if (token) qs.push(`tk=${encodeURIComponent(token)}`);
  const url = `${urlBase}/#/pages/product-list/index${qs.length ? '?' + qs.join('&') : ''}`;

  const browser = await chromium.launch({ headless: true });
  const ctx = await browser.newContext();
  const page = await ctx.newPage();

  await page.addInitScript((tk, sid) => {
    try { localStorage.setItem('token', tk); } catch (_) {}
    if (sid) {
      try { localStorage.setItem('current_store_id', String(sid)); } catch (_) {}
    }
  }, token || '', storeId);

  await page.goto(url, { waitUntil: 'domcontentloaded' });
  await page.waitForSelector('.page-product-list .grid .card-product', { timeout: 10000 });
  const btn = await page.$('.page-product-list .grid .card-product .btn-add');
  await btn.click();
  await page.waitForTimeout(500);
  await browser.close();
}

main().catch((e) => {
  console.error('add-one-item failed', e);
  process.exit(1);
});
