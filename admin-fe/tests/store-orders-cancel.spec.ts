import { test, expect, request } from '@playwright/test';

const ADMIN_FE = process.env.ADMIN_FE_URL || 'http://127.0.0.1:5173';
const API_BASE = process.env.API_BASE || 'http://127.0.0.1:9393';

// 门店订单页：管理员取消原因输入与提交流程

test.setTimeout(60000);

test('store orders: admin-cancel reason modal flow', async ({ page }) => {
  const ctx = page.context();
  const enableTrace = process.env.TRACE !== '0';
  if (enableTrace) {
    await ctx.tracing.start({ screenshots: true, snapshots: true });
  }

  // dev-login 获取 token
  const apiReq = await request.newContext();
  const resp = await apiReq.post(`${API_BASE}/api/v1/user/dev-login`, { data: { openid: 'admin_openid' } });
  expect(resp.ok()).toBeTruthy();
  const body = await resp.json();
  const token = body?.data?.token || body?.token;
  expect(token).toBeTruthy();

  // 注入运行时 API 基址与 token
  await page.addInitScript((apiBase: string) => { (window as any).__TEA_RUNTIME_CONFIG__ = { apiBaseUrl: apiBase }; }, API_BASE);
  await page.addInitScript((t: string) => localStorage.setItem('token', t), token);

  // 进入门店订单页（门店ID=1）
  await page.goto(`${ADMIN_FE}/stores/1/orders`);
  await page.waitForSelector('.ant-table', { timeout: 20000 }).catch(() => {});

  // 点击“管理员取消” -> Popconfirm 确认（通过 data-testid）-> 原因输入弹窗 -> 提交
  const cancelBtn = page.locator('table >> button:has-text("管理员取消")').first();
  await cancelBtn.waitFor({ state: 'visible', timeout: 10000 });
  await cancelBtn.click({ force: true });
  const confirmOk = page.locator('[data-testid="store-cancel-popconfirm-ok"]');
  await confirmOk.waitFor({ state: 'visible', timeout: 5000 });
  await confirmOk.click({ force: true });

  await page.waitForSelector('.ant-modal:has-text("取消订单")', { timeout: 8000 });
  await page.locator('.ant-modal textarea').first().fill('测试管理员取消原因');
  const cancelResp = page.waitForResponse(r => r.url().includes('/api/v1/orders/') && r.url().includes('/admin-cancel'));
  await page.locator('.ant-modal .ant-btn-primary:has-text("确认取消")').click({ force: true });
  const cr = await cancelResp;
  expect(cr.ok()).toBeTruthy();
  await expect(page.locator('.ant-message')).toContainText('操作成功');

  if (enableTrace) {
    await ctx.tracing.stop({ path: '/home/frederic/project/tea/build-ci-logs/playwright/store-orders-cancel-trace.zip' });
  }
});
