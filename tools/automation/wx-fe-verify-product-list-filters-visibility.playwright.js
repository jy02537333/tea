// Verify product-list shows filters in platform mode and hides in store/exclusive mode
// Usage:
//   node tools/automation/wx-fe-verify-product-list-filters-visibility.playwright.js http://127.0.0.1:9093 STORE_ID=414 TOKEN=xxx

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

async function exists(page, text) {
  try {
    const loc = page.getByText(text).first();
    await loc.waitFor({ timeout: 1500 });
    return await loc.isVisible();
  } catch (_) {
    return false;
  }
}

async function check(page, url) {
  await page.goto(url, { waitUntil: 'load' });
  // wait base container
  try { await page.locator('.page-product-list').waitFor({ timeout: 5000 }); } catch (_) {}
  await page.waitForTimeout(800);
  const label = await exists(page, '筛选与排序');
  const origin = await exists(page, '产地');
  const packaging = await exists(page, '包装');
  const price = await exists(page, '价格');
  const priceMinInput = await page.$('input[placeholder="例如 10"]');
  const priceMaxInput = await page.$('input[placeholder="例如 100"]');
  const hasPriceInputs = !!(priceMinInput || priceMaxInput);
  // also detect filters-body presence
  const filtersBodies = await page.$$('.filters .filters-body');
  const hasFiltersBody = filtersBodies.length > 0;
  return { label, origin, packaging, price, hasPriceInputs, hasFiltersBody };
}

const fs = require('fs');
const path = require('path');

async function main() {
  const { base, storeId, token } = parseArgs(process.argv);
  const urlBase = base.replace(/\/$/, '');
  const outDir = path.resolve(process.cwd(), 'build-ci-logs/verify');
  fs.mkdirSync(outDir, { recursive: true });

  const browser = await chromium.launch({ headless: true });
  const ctx = await browser.newContext();
  const page = await ctx.newPage();

  // Ensure platform check runs without lingering store_id from previous sessions
  await page.addInitScript((tk) => {
    try { localStorage.setItem('token', tk); } catch (_) {}
    try { localStorage.removeItem('current_store_id'); } catch (_) {}
  }, token || '');

  // Platform mode (no store_id): expect filters visible
  const urlPlatform = `${urlBase}/#/pages/product-list/index${token ? '?tk=' + encodeURIComponent(token) : ''}`;
  const a = await check(page, urlPlatform);
  await page.screenshot({ path: path.join(outDir, 'product-list-platform.png'), fullPage: true });
  const platformOk = a.origin || a.packaging || a.price || a.label || a.hasPriceInputs || a.hasFiltersBody;

  // Store mode (with store_id): expect filters hidden
  const urlStore = `${urlBase}/#/pages/product-list/index?store_id=${encodeURIComponent(storeId || '1')}${token ? '&tk=' + encodeURIComponent(token) : ''}`;
  const b = await check(page, urlStore);
  await page.screenshot({ path: path.join(outDir, 'product-list-store.png'), fullPage: true });
  const storeOk = !(b.origin || b.packaging || b.price || b.label || b.hasPriceInputs || b.hasFiltersBody);

  console.log(`platform_visible_ok=${platformOk}`);
  console.log(`store_hidden_ok=${storeOk}`);
  console.log(`[platform] label=${a.label} origin=${a.origin} packaging=${a.packaging} price=${a.price} priceInputs=${a.hasPriceInputs} filtersBody=${a.hasFiltersBody}`);
  console.log(`[store]    label=${b.label} origin=${b.origin} packaging=${b.packaging} price=${b.price} priceInputs=${b.hasPriceInputs} filtersBody=${b.hasFiltersBody}`);

  await browser.close();
}

main().catch((e) => {
  console.error('verify product-list filters visibility failed', e);
  process.exit(1);
});
