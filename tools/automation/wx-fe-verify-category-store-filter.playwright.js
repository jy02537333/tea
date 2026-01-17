// Verify category page uses exclusive-products when store_id is present
// Usage:
//   node tools/automation/wx-fe-verify-category-store-filter.playwright.js http://127.0.0.1:9093 STORE_ID=414 TOKEN=xxx

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
  const url = `${urlBase}/#/pages/category/index?store_id=${encodeURIComponent(storeId)}${token ? '&tk=' + encodeURIComponent(token) : ''}`;

  const outDir = path.resolve(process.cwd(), 'build-ci-logs/verify');
  fs.mkdirSync(outDir, { recursive: true });
  const logPath = path.join(outDir, 'category-store-filter.txt');

  const browser = await chromium.launch({ headless: true });
  const ctx = await browser.newContext();
  const page = await ctx.newPage();

  const requests = [];
  await page.route('**/api/v1/**', async (route, req) => {
    const u = req.url();
    if (/\/api\/v1\/(products|stores\/.+\/exclusive-products)/.test(u)) {
      requests.push(u);
    }
    await route.continue();
  });

  await page.goto(url, { waitUntil: 'domcontentloaded' });
  // 尝试触发商品请求：
  // 1) 等待“选择门店”或搜索框渲染
  try { await page.getByText('选择门店').first().waitFor({ timeout: 8000 }); } catch (_) {}
  // 2) 在搜索框输入并回车，触发 handleKeywordConfirm -> fetchProducts(reset)
  try {
    const search = await page.$('input[placeholder="搜索商品关键字"]');
    if (search) {
      await search.fill('茶');
      await search.press('Enter');
    }
  } catch (_) {}
  // 3) 点击“全部”或第一个分类 pill
  try {
    const allPill = await page.getByText('全部').first();
    if (allPill) await allPill.click();
  } catch (_) {}
  // 4) 如果存在“加载更多”，点击一次
  try {
    const more = await page.getByText('加载更多').first();
    await more.click({ timeout: 2000 });
  } catch (_) {}
  // 5) 滚动页面，给请求一些时间
  await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
  await page.waitForTimeout(2000);

  // 6) 再次触发一轮：等待更久，确保选中门店状态生效
  await page.waitForTimeout(2000);
  try {
    const search2 = await page.$('input[placeholder="搜索商品关键字"]');
    if (search2) {
      await search2.fill('绿茶');
      await search2.press('Enter');
    }
  } catch (_) {}
  await page.waitForTimeout(1500);
  try {
    const more2 = await page.getByText('加载更多').first();
    await more2.click({ timeout: 2000 });
  } catch (_) {}
  await page.waitForTimeout(1500);

  // 7) 刷新一次页面再等待，进一步兜底
  try {
    await page.reload({ waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(2000);
    const search3 = await page.$('input[placeholder="搜索商品关键字"]');
    if (search3) {
      await search3.fill('红茶');
      await search3.press('Enter');
    }
    await page.waitForTimeout(1500);
  } catch (_) {}

  const matched = requests.filter((u) => /\/api\/v1\/(products|stores\/.+\/exclusive-products)/.test(u));
  const exclusive = matched.find((u) => new RegExp(`/api/v1/stores/${storeId}/exclusive-products`).test(u));
  const withStore = matched.find((u) => /\/api\/v1\/products/.test(u) && /[?&]store_id=/.test(u));
  const storeOk = !!exclusive || (withStore && new URL(withStore).searchParams.get('store_id') === String(storeId));

  const lines = [];
  lines.push(`requests_count=${matched.length}`);
  lines.push(`exclusive_url=${exclusive || ''}`);
  lines.push(`with_store_id=${withStore || ''}`);
  lines.push(`store_ok=${!!storeOk}`);
  // 追加部分请求样本，便于排查
  lines.push('samples=');
  for (const u of matched.slice(0, 5)) lines.push(u);
  fs.writeFileSync(logPath, lines.join('\n'));

  console.log(lines.join('\n'));
  await browser.close();
}

main().catch((e) => {
  console.error('verify category store filter failed', e);
  process.exit(1);
});
