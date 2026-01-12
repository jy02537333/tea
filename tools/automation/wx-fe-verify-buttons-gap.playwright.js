// Verify the pixel gap between '详情' and '加入购物车' buttons on product-list page
// Usage:
//   node tools/automation/wx-fe-verify-buttons-gap.playwright.js http://127.0.0.1:9093 STORE_ID=414 TOKEN=xxx

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
  if (!args.token && process.env.TOKEN) args.token = process.env.TOKEN;
  return args;
}

async function main() {
  const { base, storeId, token } = parseArgs(process.argv);
  const urlBase = base.replace(/\/$/, '');
  const qs = [];
  if (storeId) qs.push(`store_id=${encodeURIComponent(storeId)}`);
  if (token) qs.push(`tk=${encodeURIComponent(token)}`);
  const url = `${urlBase}/#/pages/product-list/index${qs.length ? '?' + qs.join('&') : ''}`;

  const outDir = path.resolve(process.cwd(), 'build-ci-logs/verify');
  fs.mkdirSync(outDir, { recursive: true });
  const shotPath = path.join(outDir, 'buttons-gap.png');
  const logPath = path.join(outDir, 'buttons-gap.txt');

  const browser = await chromium.launch({ headless: true });
  const ctx = await browser.newContext();
  const page = await ctx.newPage();

  // token 注入到 localStorage 兜底
  if (token) {
    await page.addInitScript((tk) => {
      try { localStorage.setItem('token', tk); } catch (e) {}
    }, token);
  }

  await page.goto(url, { waitUntil: 'domcontentloaded' });
  // 等待首个商品卡片出现
  await page.waitForSelector('.page-product-list .grid .card-product', { timeout: 10000 });

  const cards = await page.$$('.page-product-list .grid .card-product');
  const results = [];

  for (let i = 0; i < Math.min(cards.length, 3); i++) {
    const card = cards[i];
    const viewBtn = await card.$('.btn-view');
    const addBtn = await card.$('.btn-add');
    if (!viewBtn || !addBtn) continue;
    const vb = await viewBtn.boundingBox();
    const ab = await addBtn.boundingBox();
    if (!vb || !ab) continue;
    const gap = ab.x - (vb.x + vb.width);
    results.push({ index: i, gap });
    if (i === 0) {
      const clip = await card.boundingBox();
      if (clip) {
        await page.screenshot({ path: shotPath, clip });
      }
    }
  }

  const lines = results.map(r => `card#${r.index} gap_px=${r.gap.toFixed(2)}`);
  const ok = results.length > 0 && results.every(r => r.gap > 0);
  const exact5 = results.length > 0 && results.every(r => Math.abs(r.gap - 5) <= 1.0);
  lines.push(`has_gap=${ok}`);
  lines.push(`approx_5px=${exact5}`);
  fs.writeFileSync(logPath, lines.join('\n'));

  console.log(lines.join('\n'));
  console.log(`[screenshot] ${shotPath}`);

  await browser.close();
}

main().catch((e) => {
  console.error('verify failed', e);
  process.exit(1);
});
