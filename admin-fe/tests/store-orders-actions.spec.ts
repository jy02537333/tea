import { test, expect, request } from '@playwright/test';

const ADMIN_FE = process.env.ADMIN_FE_URL || 'http://127.0.0.1:9094';
const API_BASE = process.env.API_BASE || 'http://127.0.0.1:9292';

test.setTimeout(60000);

test('store orders page shows table column and action buttons when eligible', async ({ page }) => {
  const apiReq = await request.newContext();

  // 1) store001 登录拿 token 与 store_id
  const storeLogin = await apiReq.post(`${API_BASE}/api/v1/user/login`, {
    data: { username: 'store001', password: 'Store@123' },
  });
  expect(storeLogin.ok()).toBeTruthy();
  const storeBody = await storeLogin.json();
  const storeToken = storeBody?.data?.token || storeBody?.token;
  const storeId = storeBody?.data?.user_info?.store_id || storeBody?.user_info?.store_id;
  expect(storeToken).toBeTruthy();
  expect(storeId).toBeTruthy();

  // 2) 注入运行时 API 基址与 token
  await page.addInitScript((apiBase: string) => {
    (window as any).__TEA_RUNTIME_CONFIG__ = { apiBaseUrl: apiBase };
  }, API_BASE);
  await page.addInitScript((t: string) => localStorage.setItem('token', t), storeToken);

  // 3) 打开门店订单列表页
  await page.goto(`${ADMIN_FE}/stores/${storeId}/orders`);

  // 桌号列应存在（说明前端为新版本）
  await expect(page.locator('table')).toBeVisible({ timeout: 15000 });
  await expect(page.locator('table')).toContainText('桌号');

  // “完成”按钮：当存在 status=6/7/3 且用户有 order:complete 权限时出现
  // 这里只做“至少出现一次”的弱断言，用于验证权限拉取与按钮渲染链路。
  const completeCount = await page.locator('a:has-text("完成"), button:has-text("完成")').count();
  expect(completeCount).toBeGreaterThan(0);
});
