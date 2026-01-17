// Verify checkout page shows items, images, amount, and supports
// local pre-check + quick clean for non-store items.
// Usage:
//   node tools/automation/wx-fe-verify-checkout-clean.playwright.js http://127.0.0.1:9093 STORE_ID=414 TOKEN=xxx

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
  const envTokens = loadTokensEnv();
  if (!args.token && process.env.TOKEN) args.token = process.env.TOKEN;
  if (!args.token && envTokens.AUTH_TOKEN) args.token = envTokens.AUTH_TOKEN;
  return args;
}

function loadTokensEnv() {
  const p = path.resolve(process.cwd(), 'build-ci-logs/tokens.env');
  const out = {};
  try {
    const txt = fs.readFileSync(p, 'utf-8');
    for (const line of txt.split(/\r?\n/)) {
      const m = line.match(/^([A-Z0-9_]+)=(.*)$/);
      if (m) out[m[1]] = m[2];
    }
  } catch (_) {}
  return out;
}

async function addOneProduct(page, base, { storeId, token }) {
  const urlBase = base.replace(/\/$/, '');
  const qs = [];
  if (storeId) qs.push(`store_id=${encodeURIComponent(storeId)}`);
  if (token) qs.push(`tk=${encodeURIComponent(token)}`);
  const url = `${urlBase}/#/pages/product-list/index${qs.length ? '?' + qs.join('&') : ''}`;
  await page.goto(url, { waitUntil: 'domcontentloaded' });
  await page.waitForSelector('.page-product-list .grid .card-product', { timeout: 10000 });
  const firstAdd = await page.$('.page-product-list .grid .card-product .btn-add');
  if (firstAdd) {
    await firstAdd.click();
    await page.waitForTimeout(300);
  }
}

async function main() {
  const { base, storeId, token } = parseArgs(process.argv);
  if (!storeId) throw new Error('STORE_ID is required');
  const urlBase = base.replace(/\/$/, '');

  const outDir = path.resolve(process.cwd(), 'build-ci-logs/verify');
  fs.mkdirSync(outDir, { recursive: true });
  const shotPath = path.join(outDir, 'checkout-clean.png');
  const logPath = path.join(outDir, 'checkout-clean.txt');

  const browser = await chromium.launch({ headless: true });
  const ctx = await browser.newContext();
  const page = await ctx.newPage();

  // Inject token and current store id
  await page.addInitScript((tk, sid) => {
    try { localStorage.setItem('token', tk); } catch (_) {}
    try { localStorage.setItem('current_store_id', String(sid)); } catch (_) {}
  }, token || '', storeId);

  // Step 1: add a platform product (no storeId)
  await addOneProduct(page, base, { storeId: undefined, token });
  // Step 2: add a store product
  await addOneProduct(page, base, { storeId, token });

  // Step 3: go to checkout
  const checkoutUrl = `${urlBase}/#/pages/checkout/index${token ? '?tk=' + encodeURIComponent(token) : ''}`;
  await page.goto(checkoutUrl, { waitUntil: 'domcontentloaded' });
  await page.waitForSelector('text=确认订单', { timeout: 10000 });

  // Check warning + quick clean button presence
  const warning = await page.$('text=购物车包含平台或其他门店商品');
  const cleanBtn = await page.$('text=仅保留本店并重试');
  const submitBtn = await page.$('text=提交订单');
  const beforeDisabled = submitBtn ? !(await submitBtn.isEnabled()) : null;

  // If clean button exists, click it
  if (cleanBtn) {
    await cleanBtn.click();
    await page.waitForTimeout(800);
    // Confirm modal
    const confirm = await page.$('text=移除');
    if (confirm) {
      await confirm.click();
      await page.waitForTimeout(1200);
    }
  }

  // After clean, check that warning disappeared and submit unlocked
  const warningAfter = await page.$('text=购物车包含平台或其他门店商品');
  const submitBtnAfter = await page.$('text=提交订单');
  const afterDisabled = submitBtnAfter ? !(await submitBtnAfter.isEnabled()) : null;

  // Verify list items contain images and amount
  const itemRows = await page.$$('.taro-img, img');
  const hasImages = itemRows.length > 0;
  const totalText = await page.textContent('text=金额合计');

  const lines = [];
  lines.push(`warning_present_before=${!!warning}`);
  lines.push(`submit_disabled_before=${beforeDisabled}`);
  lines.push(`clean_button_present=${!!cleanBtn}`);
  lines.push(`warning_present_after=${!!warningAfter}`);
  lines.push(`submit_disabled_after=${afterDisabled}`);
  lines.push(`has_images=${hasImages}`);
  lines.push(`total_text=${(totalText || '').trim()}`);
  fs.writeFileSync(logPath, lines.join('\n'));

  // Screenshot
  await page.screenshot({ path: shotPath, fullPage: true });
  console.log(lines.join('\n'));
  console.log(`[screenshot] ${shotPath}`);

  await browser.close();
}

main().catch((e) => {
  console.error('verify checkout failed', e);
  process.exit(1);
});
