// Verify that navigating from homepage to dine-in requests products with store_id filter
// Usage:
//   node tools/automation/wx-fe-verify-dinein-store-filter.playwright.js http://127.0.0.1:9093 STORE_ID=414 TOKEN=xxx

const fs = require('fs');
const path = require('path');
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
  if (!storeId) throw new Error('STORE_ID is required');
  const urlBase = base.replace(/\/$/, '');
  const homeUrl = `${urlBase}/#/pages/index/index${token ? '?tk=' + encodeURIComponent(token) : ''}`;

  const outDir = path.resolve(process.cwd(), 'build-ci-logs/verify');
  fs.mkdirSync(outDir, { recursive: true });
  const logPath = path.join(outDir, 'dinein-store-filter.txt');
  const shotPath = path.join(outDir, 'dinein-store-filter.png');

  const browser = await chromium.launch({ headless: true });
  const ctx = await browser.newContext();
  const page = await ctx.newPage();

  // Inject token and store
  await page.addInitScript((tk, sid) => {
    try { localStorage.setItem('token', tk); } catch (_) {}
    try { localStorage.setItem('current_store_id', String(sid)); } catch (_) {}
  }, token || '', storeId);

  const requests = [];
  await page.route('**/api/v1/**', async (route, req) => {
    const u = req.url();
    if (/\/api\/v1\/(products|stores\/.+\/exclusive-products)/.test(u)) {
      requests.push(u);
    }
    await route.continue();
  });

  await page.goto(homeUrl, { waitUntil: 'domcontentloaded' });
  // Click dine-in tile
  await page.click('text=堂食');
  // Wait a bit for requests to fire
  await page.waitForTimeout(1200);

  // Gather any /products requests
  const matched = requests.filter((u) => /\/api\/v1\/(products|stores\/.+\/exclusive-products)/.test(u));
  const exclusive = matched.find((u) => new RegExp(`/api/v1/stores/${storeId}/exclusive-products`).test(u));
  const withStore = matched.find((u) => /\/api\/v1\/products/.test(u) && /[?&]store_id=/.test(u));
  const storeOk = !!exclusive || (withStore && new URL(withStore).searchParams.get('store_id') === String(storeId));

  const lines = [];
  lines.push(`requests_count=${matched.length}`);
  lines.push(`exclusive_url=${exclusive || ''}`);
  lines.push(`with_store_id=${withStore || ''}`);
  lines.push(`store_ok=${!!storeOk}`);
  fs.writeFileSync(logPath, lines.join('\n'));
  await page.screenshot({ path: shotPath, fullPage: true });
  console.log(lines.join('\n'));
  console.log(`[screenshot] ${shotPath}`);

  await browser.close();
}

main().catch((e) => {
  console.error('verify dine-in store filter failed', e);
  process.exit(1);
});
