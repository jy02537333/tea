// Verify category page shows origin/packaging/price filters when no store_id (platform mode)
// Usage:
//   node tools/automation/wx-fe-verify-category-show-filters.playwright.js http://127.0.0.1:9093 TOKEN=xxx

const { chromium } = require('playwright');

function parseArgs(argv) {
  const args = { base: 'http://127.0.0.1:9093', token: undefined };
  if (argv[2] && /^https?:/.test(argv[2])) args.base = argv[2];
  for (const a of argv.slice(2)) {
    if (/^TOKEN=/.test(a)) args.token = a.split('=')[1];
  }
  return args;
}

async function main() {
  const { base, token } = parseArgs(process.argv);
  const urlBase = base.replace(/\/$/, '');
  const url = `${urlBase}/#/pages/category/index${token ? '?tk=' + encodeURIComponent(token) : ''}`;

  const browser = await chromium.launch({ headless: true });
  const ctx = await browser.newContext();
  const page = await ctx.newPage();

  await page.goto(url, { waitUntil: 'domcontentloaded' });
  // 等待页面稳定并确保“选择门店”渲染完成
  await page.waitForSelector('text=选择门店', { timeout: 10000 }).catch(() => {});
  await page.waitForTimeout(800);

  async function exists(text) {
    const el = await page.$(`text=${text}`);
    return !!el;
  }

  const chooseStore = await exists('选择门店');
  const label = await exists('筛选与排序');
  const origin = await exists('产地');
  const packaging = await exists('包装');
  const priceRow = await exists('价格');
  const priceMin = await exists('最低价');
  const priceMax = await exists('最高价');

  const price = priceRow || priceMin || priceMax;
  const visibleOk = (origin || packaging || price);
  console.log(`visible_ok=${visibleOk}`);
  console.log(`choose_store=${chooseStore} label=${label} origin=${origin} packaging=${packaging} price=${price} (row=${priceRow}, min=${priceMin}, max=${priceMax})`);

  await browser.close();
}

main().catch((e) => {
  console.error('verify category show filters failed', e);
  process.exit(1);
});
