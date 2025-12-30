import { test, expect, request } from '@playwright/test';

const ADMIN_FE = process.env.ADMIN_FE_URL || 'http://127.0.0.1:5173';
const API_BASE = process.env.API_BASE || 'http://127.0.0.1:9393';

test.setTimeout(60000);

test('orders page adjust price flow (modal submit)', async ({ page }) => {
  const ctx = page.context();
  const enableTrace = process.env.TRACE !== '0';
  if (enableTrace) {
    await ctx.tracing.start({ screenshots: true, snapshots: true });
  }

  // dev-login to get token
  const apiReq = await request.newContext();
  const resp = await apiReq.post(`${API_BASE}/api/v1/user/dev-login`, { data: { openid: 'admin_openid' } });
  expect(resp.ok()).toBeTruthy();
  const body = await resp.json();
  const token = body?.data?.token || body?.token;
  expect(token).toBeTruthy();

  // inject runtime config and token
  await page.addInitScript((apiBase: string) => { (window as any).__TEA_RUNTIME_CONFIG__ = { apiBaseUrl: apiBase }; }, API_BASE);
  await page.addInitScript((t: string) => localStorage.setItem('token', t), token);

  await page.goto(`${ADMIN_FE}/orders`);
  await page.waitForSelector('.ant-table', { timeout: 20000 }).catch(() => {});

  // click 调价 on first row that has it
  const adjustBtn = page.locator('table >> button:has-text("调价")').first();
  await adjustBtn.waitFor({ state: 'visible', timeout: 10000 });
  await adjustBtn.click({ force: true });

  // modal should appear with input for 新支付金额
  await page.waitForSelector('.ant-modal:has-text("订单调价")', { timeout: 8000 });
  const amount = page.getByRole('spinbutton', { name: '* 新支付金额' });
  await amount.fill('88');
  await amount.blur();
  await page.getByRole('textbox', { name: '调价原因' }).fill('测试调价');

  const ok = page.getByRole('button', { name: '确认调价' });
  await expect(ok).toBeEnabled();

  const [adjustResponse] = await Promise.all([
    page.waitForResponse(r => /\/api\/v1\/orders\/\d+\/adjust$/.test(r.url()) && r.request().method() === 'POST'),
    ok.click(),
  ]);
  expect(adjustResponse.ok()).toBeTruthy();

  // optional: success message appears
  await expect(page.locator('.ant-message')).toContainText('调价成功');

  if (enableTrace) {
    await ctx.tracing.stop({ path: '/home/frederic/project/tea/build-ci-logs/playwright/order-adjust-trace.zip' });
  }
});
