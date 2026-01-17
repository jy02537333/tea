// Verify category page hides unsupported filters when store_id is present
// Usage:
//   node tools/automation/wx-fe-verify-category-hide-filters.playwright.js http://127.0.0.1:9093 STORE_ID=414 TOKEN=xxx

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
  const url = `${urlBase}/#/pages/category/index?store_id=${encodeURIComponent(storeId)}${token ? '&tk=' + encodeURIComponent(token) : ''}`;

  const browser = await chromium.launch({ headless: true });
  const ctx = await browser.newContext();
  const page = await ctx.newPage();

  await page.goto(url, { waitUntil: 'domcontentloaded' });
  // Wait basic layout
  await page.waitForSelector('text=选择门店', { timeout: 10000 }).catch(() => {});
  await page.waitForTimeout(300);

  async function exists(text) {
    const el = await page.$(`text=${text}`);
    return !!el;
  }

  const hasLabel = await exists('筛选与排序');
  const hasOrigin = await exists('产地');
  const hasPackaging = await exists('包装');
  const hasPrice = await exists('价格');
  const hasSort = await exists('排序');

  console.log(`hidden_ok=${!(hasLabel || hasOrigin || hasPackaging || hasPrice || hasSort)}`);
  console.log(`label=${hasLabel} origin=${hasOrigin} packaging=${hasPackaging} price=${hasPrice} sort=${hasSort}`);

  await browser.close();
}

main().catch((e) => {
  console.error('verify category hide filters failed', e);
  process.exit(1);
});
