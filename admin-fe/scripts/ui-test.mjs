import { chromium } from 'playwright';

const APP_URL = process.env.APP_URL || 'http://127.0.0.1:5173';
const ADMIN_TOKEN = process.env.ADMIN_TOKEN || '';

async function run() {
  if (!ADMIN_TOKEN) {
    console.error('ADMIN_TOKEN not provided');
    process.exit(2);
  }

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await context.newPage();

  try {
    console.log('Opening app root to set token...');
    await page.goto(APP_URL, { waitUntil: 'networkidle' });
    await page.evaluate((t) => { localStorage.setItem('token', t); }, ADMIN_TOKEN);
    await page.reload({ waitUntil: 'networkidle' });

    // Partners page
    console.log('Navigating to /partners');
    await page.goto(`${APP_URL}/partners`, { waitUntil: 'networkidle' });
    await page.screenshot({ path: 'ui-partners.png' });
    const partnersHeader = await page.locator('text=合伙人管理').first().count();
    console.log('Partners header found:', partnersHeader > 0);

    // Partner withdrawals
    console.log('Navigating to /partner-withdrawals');
    await page.goto(`${APP_URL}/partner-withdrawals`, { waitUntil: 'networkidle' });
    await page.screenshot({ path: 'ui-withdrawals-list.png' });

    // Try to click first 审核 button if present
    const auditBtn = page.locator('button:has-text("审核")').first();
    if (await auditBtn.count() > 0) {
      console.log('Found 审核 button, opening modal...');
      await auditBtn.click();
      await page.waitForTimeout(500);
      await page.screenshot({ path: 'ui-withdrawal-modal.png' });

      // fill remark
      const ta = page.locator('textarea');
      if (await ta.count() > 0) {
        await ta.fill('自动化测试备注');
      }

      // Try clicking 受理 if exists
      const accept = page.locator('button:has-text("受理")');
      if (await accept.count() > 0) {
        console.log('Clicking 受理');
        await accept.click();
        await page.waitForTimeout(800);
        await page.screenshot({ path: 'ui-after-accept.png' });
      } else {
        console.log('No 受理 button found');
      }

    } else {
      console.log('No 审核 buttons found on withdrawals page');
    }

    console.log('UI automation finished successfully');
    await browser.close();
    process.exit(0);
  } catch (err) {
    console.error('UI automation error:', err);
    await page.screenshot({ path: 'ui-error.png' }).catch(() => {});
    await browser.close();
    process.exit(3);
  }
}

run();
