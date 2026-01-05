import { test, expect, request } from '@playwright/test';

const ADMIN_FE = process.env.ADMIN_FE_URL || 'http://127.0.0.1:5173';
const API_BASE = process.env.API_BASE || 'http://127.0.0.1:9393';

test.setTimeout(60000);

test('partner withdrawal approve flow', async ({ page }) => {
  const ctx = page.context();
  // start tracing (will collect snapshots and screenshots) unless disabled
  const enableTrace = process.env.TRACE !== '0';
  if (enableTrace) {
    await ctx.tracing.start({ screenshots: true, snapshots: true });
  }
  // obtain admin token via dev-login
  const apiRequest = await request.newContext();
  const resp = await apiRequest.post(`${API_BASE}/api/v1/user/dev-login`, { data: { openid: 'admin_openid' } });
  expect(resp.ok()).toBeTruthy();
  const body = await resp.json();
  const token = body?.data?.token || body?.token;
  expect(token).toBeTruthy();

  // set API base and token into localStorage before app boot
  await page.addInitScript((apiBase: string) => { (window as any).__TEA_RUNTIME_CONFIG__ = { apiBaseUrl: apiBase }; }, API_BASE);
  await page.addInitScript((t: string) => { localStorage.setItem('token', t); }, token);

  // navigate to partner withdrawals page
  await page.goto(`${ADMIN_FE}/partner-withdrawals`);
  await page.waitForSelector('text=合伙人提现审核, >> button:has-text("审")', { timeout: 15000 }).catch(() => {});

  // open first 审核 modal - be robust: wait for a visible button in the table row
  const reviewButton = page.locator('table >> button:has-text("审")').first();
  try {
    await reviewButton.waitFor({ state: 'visible', timeout: 20000 });
    // always capture pre-click screenshot for debugging
    const outDir = '/home/frederic/project/tea/build-ci-logs/playwright';
    await page.screenshot({ path: `${outDir}/partner-withdrawal-before-click.png`, fullPage: true }).catch(() => {});
    // try a forced click in case spacing/overlay interferes
    await reviewButton.click({ force: true });
    // wait briefly for modal to appear; if not, try a DOM-eval fallback click
    let modalAppeared = false;
    try {
      await page.waitForSelector('.ant-modal', { timeout: 4000 });
      modalAppeared = true;
    } catch (e) {
      // fallback: try clicking via DOM query (handles unusual spacing/characters)
      await page.evaluate(() => {
        const candidates = Array.from(document.querySelectorAll('button'));
        for (const b of candidates) {
          try { if ((b.textContent || '').replace(/\s+/g, '').includes('审核') || (b.textContent || '').replace(/\s+/g, '').includes('审')) { (b as HTMLElement).click(); break; } } catch (err) {}
        }
      });
      try {
        await page.waitForSelector('.ant-modal', { timeout: 10000 });
        modalAppeared = true;
      } catch (ee) {
        modalAppeared = false;
      }
    }
    await page.screenshot({ path: `${outDir}/partner-withdrawal-modal.png`, fullPage: true }).catch(() => {});
  } catch (err) {
    // on failure save screenshot and stop tracing to help debugging
    const outDir = '/home/frederic/project/tea/build-ci-logs/playwright';
    await page.screenshot({ path: `${outDir}/partner-withdrawal-failure.png`, fullPage: true }).catch(() => {});
    if (enableTrace) {
      await ctx.tracing.stop({ path: `${outDir}/partner-withdrawal-trace.zip` }).catch(() => {});
    }
    throw err;
  }

  // fill remark and approve (intercept network)
  await page.fill('textarea[placeholder="备注（选填）"]', '自动化测试受理');

  const approvePromise = page.waitForResponse(resp => resp.url().includes('/api/v1/admin/withdraws/') && resp.url().includes('/approve'));
  // click the approve button within modal
  // try to find the approve button inside modal by matching text, fallback to first actionable button
  const modalButtons = page.locator('.ant-modal button');
  const btnCount = await modalButtons.count();
  let clicked = false;
  for (let i = 0; i < btnCount; i++) {
    const b = modalButtons.nth(i);
    const txt = (await b.innerText()).replace(/\s+/g, '');
    if (/受理|受/.test(txt)) {
      await b.click({ force: true }); clicked = true; break;
    }
  }
  if (!clicked && btnCount > 0) {
    // fallback: click the last primary button (often the first or second)
    await modalButtons.nth(btnCount - 1).click({ force: true });
  }
  const approveResp = await approvePromise;
  expect(approveResp.ok()).toBeTruthy();

  // stop tracing and save artifacts
  if (enableTrace) {
    await ctx.tracing.stop({ path: '/home/frederic/project/tea/build-ci-logs/playwright/partner-withdrawal-trace.zip' });
  }

  // open reject path as smoke (re-open if modal closed)
  // create a second withdrawal via script before running e2e to test reject path if needed
});
