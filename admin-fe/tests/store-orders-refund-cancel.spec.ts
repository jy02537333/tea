import { test, expect, request } from '@playwright/test';

const ADMIN_FE = process.env.ADMIN_FE_URL || 'http://127.0.0.1:5173';
const API_BASE = process.env.API_BASE || 'http://127.0.0.1:9393';

// 覆盖门店订单页的“立即退款”原因输入与提交流程（管理员取消已由 Orders 页面用例覆盖）

test.setTimeout(60000);

test('store orders: refund reason modal flow', async ({ page }) => {
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

  // 直接进入门店订单页（门店ID=1）
  await page.goto(`${ADMIN_FE}/stores/1/orders`);
  await page.waitForSelector('.ant-table', { timeout: 20000 }).catch(() => {});

  // 退款：点击表格中第一个“立即退款”，弹出原因输入框并提交
  const refundBtn = page.locator('table >> button:has-text("立即退款")').first();
  await refundBtn.waitFor({ state: 'visible', timeout: 10000 });
  await refundBtn.click({ force: true });
  // 先确认 Popconfirm，再出现原因输入弹窗
  await page.waitForSelector('.ant-popover', { timeout: 5000 });
  const refundConfirm = page.locator('.ant-popover:has-text("确认执行立即退款？") .ant-btn-primary, .ant-popconfirm:has-text("确认执行立即退款？") .ant-btn-primary').first();
  await refundConfirm.waitFor({ state: 'visible', timeout: 3000 });
  await refundConfirm.click({ force: true });
  await page.waitForSelector('.ant-modal:has-text("填写原因")', { timeout: 8000 });
  await page.locator('.ant-modal textarea').first().fill('测试立即退款原因');
  const refundResp = page.waitForResponse(r => r.url().includes('/api/v1/orders/') && r.url().includes('/refund')); // 匹配 refund 请求
  await page.locator('.ant-modal .ant-btn-primary:has-text("确认提交")').click({ force: true });
  const rr = await refundResp;
  expect(rr.ok()).toBeTruthy();
  await expect(page.locator('.ant-message')).toContainText('操作成功');

  // 管理员取消：该页面的 Popconfirm 在可见性上较难自动化稳定，取消流程已在 Orders 页面用例覆盖
  // 此处仅覆盖“门店订单页”的退款原因输入与提交流程。
