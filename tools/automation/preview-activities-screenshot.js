#!/usr/bin/env node
/*
 Automates H5 preview to: inject token, open activities page, fill fields,
 click register+pay, wait for success marker, and save a screenshot.
*/
const fs = require('fs');
const path = require('path');
const puppeteer = require('puppeteer');

async function readJSON(p) {
  return JSON.parse(fs.readFileSync(p, 'utf-8'));
}

async function main() {
  const repoRoot = path.resolve(__dirname, '../../');
  const logsDir = path.join(repoRoot, 'build-ci-logs');
  const screenshotsDir = path.join(logsDir, 'screenshots');
  if (!fs.existsSync(screenshotsDir)) fs.mkdirSync(screenshotsDir, { recursive: true });

  const previewURL = process.env.PREVIEW_URL || 'http://127.0.0.1:10093/';

  // Pull token and store id from prior integration logs
  const loginFile = path.join(logsDir, 'admin_login_response.json');
  const storeFile = path.join(logsDir, 'activity_demo_create_store.json');
  const actFile = path.join(logsDir, 'activity_demo_create_activity.json');
  const token = (await readJSON(loginFile)).data?.token || '';
  const storeId = (await readJSON(storeFile)).data?.id || 0;
  const activityId = (await readJSON(actFile)).data?.id || 0;
  if (!token || !storeId || !activityId) {
    throw new Error('Missing token/storeId/activityId. Please run backend setup first.');
  }

  const name = process.env.ACT_NAME || '张三';
  const phone = process.env.ACT_PHONE || '18000000001';
  const fee = process.env.ACT_FEE || '9.9';

  const browser = await puppeteer.launch({
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox'],
  });
  const page = await browser.newPage();
  await page.setViewport({ width: 420, height: 880, deviceScaleFactor: 1 });

  // 1) Open preview root and inject token
  await page.goto(previewURL, { waitUntil: 'networkidle0', timeout: 60000 });
  await page.evaluate((t) => localStorage.setItem('token', t), token);
  await page.reload({ waitUntil: 'networkidle0' });

  // 2) Navigate to activities page (browser history route for Taro H5)
  const actURL = new URL(previewURL);
  const base = actURL.toString().replace(/#.*$/, '').replace(/\/$/, '');
  const qName = encodeURIComponent(name);
  const qPhone = encodeURIComponent(phone);
  const qFee = encodeURIComponent(String(fee));
  const target = `${base}/pages/activities/index?store_id=${storeId}&activity_id=${activityId}&name=${qName}&phone=${qPhone}&fee=${qFee}&auto=1`;
  await page.goto(target, { waitUntil: 'networkidle2' });

  // 3) Fill inputs by placeholder
  const typeIn = async (placeholder, value) => {
    const sel = `input[placeholder="${placeholder}"]`;
    await page.waitForSelector(sel, { timeout: 15000 });
    await page.click(sel, { delay: 50 });
    await page.type(sel, value, { delay: 20 });
  };
  try { await typeIn('请输入姓名', name); } catch (_) {}
  try { await typeIn('请输入手机号', phone); } catch (_) {}
  try { await typeIn('不填则视为免费活动', String(fee)); } catch (_) {}

  // 4) Click the first visible button containing “报名”
  // When auto=1, the page should auto-submit. Keep a minimal fallback.
  try {
    const btns = await page.$$('button');
    for (const b of btns) {
      const text = (await page.evaluate(el => el.textContent || '', b)).trim();
      if (/报名/.test(text)) { await b.click({ delay: 50 }); break; }
    }
  } catch (_) {}

  // 5) Wait for success marker “已报名成功” to appear
  await page.waitForFunction(
    () => !!document.body.textContent && document.body.textContent.includes('已报名成功'),
    { timeout: 60000 }
  );

  // 6) Screenshot
  const out = path.join(screenshotsDir, 'activities_paid_success.png');
  await page.screenshot({ path: out, fullPage: true });
  console.log('Saved screenshot:', out);

  await browser.close();
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
