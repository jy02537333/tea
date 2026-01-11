import { test, expect, request } from '@playwright/test';

const ADMIN_FE = process.env.ADMIN_FE_URL || 'http://127.0.0.1:9094';
const API_BASE = process.env.API_BASE || 'http://127.0.0.1:9292';

// 门店桌号管理：新增/列表/删除（门店管理员 role=store）

test.setTimeout(60000);

test('store settings: tables CRUD via UI', async ({ page }) => {
  const ctx = page.context();
  const enableTrace = process.env.TRACE !== '0';
  if (enableTrace) {
    await ctx.tracing.start({ screenshots: true, snapshots: true });
  }

  // 使用门店管理员账号登录以获取 token（store001 / Store@123）
  const apiReq = await request.newContext();
  const loginResp = await apiReq.post(`${API_BASE}/api/v1/user/login`, {
    data: { username: 'store001', password: 'Store@123' },
  });
  expect(loginResp.ok()).toBeTruthy();
  const loginBody = await loginResp.json();
  const token = loginBody?.data?.token || loginBody?.token;
  const storeId = loginBody?.data?.user_info?.store_id || loginBody?.user_info?.store_id;
  expect(token).toBeTruthy();
  expect(storeId).toBeTruthy();

  // 注入运行时 API 基址与 token
  await page.addInitScript((apiBase: string) => { (window as any).__TEA_RUNTIME_CONFIG__ = { apiBaseUrl: apiBase }; }, API_BASE);
  await page.addInitScript((t: string) => localStorage.setItem('token', t), token);

  // 进入门店设置页
  await page.goto(`${ADMIN_FE}/store-settings`);

  // 等待桌号管理区域出现
  await page.waitForSelector('div.ant-card:has-text("门店桌号管理")');

  // 填写新增表单（桌号必填，其余为可选）
  await page.locator('form >> input[placeholder="例如 A12 或 5 号桌"]').fill('A01');

  // 监听创建接口请求
  const createRespPromise = page.waitForResponse(r => r.url().includes(`/api/v1/stores/${storeId}/tables`) && r.request().method() === 'POST');
  await page.locator('form >> text=新增桌号').click({ force: true });
  const createResp = await createRespPromise;
  expect(createResp.ok()).toBeTruthy();
  await expect(page.locator('.ant-message')).toContainText('新增桌号成功');

  // 列表应出现新记录，抓到第一行的删除按钮
  const deleteBtn = page.locator('table >> text=删除').first();
  await deleteBtn.waitFor({ state: 'visible', timeout: 10000 });

  // 删除并校验接口
  const deleteRespPromise = page.waitForResponse(r => r.url().includes(`/api/v1/stores/${storeId}/tables/`) && r.request().method() === 'DELETE');
  await deleteBtn.click({ force: true });
  const popOk = page.locator('.ant-popconfirm >> .ant-btn-primary');
  await popOk.scrollIntoViewIfNeeded();
  await popOk.click({ force: true });
  const delResp = await deleteRespPromise;
  expect(delResp.ok()).toBeTruthy();
  await expect(page.locator('.ant-message')).toContainText('删除成功');

  if (enableTrace) {
    await ctx.tracing.stop({ path: '/home/frederic/project/tea/build-ci-logs/playwright/store-tables-trace.zip' });
  }
});
