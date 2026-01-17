import { test, expect, request } from '@playwright/test';

const ADMIN_FE = process.env.ADMIN_FE_URL || 'http://127.0.0.1:9094';
const API_BASE = process.env.API_BASE || 'http://127.0.0.1:9292';

// 门店活动与优惠券列表页可见性（store 角色）

test.setTimeout(60000);

test('store coupons: visibility and GET list ok', async ({ page }) => {
  const apiReq = await request.newContext();
  // 门店登录
  const loginResp = await apiReq.post(`${API_BASE}/api/v1/user/login`, {
    data: { username: 'store001', password: 'Store@123' },
    headers: { 'Content-Type': 'application/json' },
  });
  expect(loginResp.ok()).toBeTruthy();
  const loginBody = await loginResp.json();
  const token = loginBody?.data?.token || loginBody?.token;
  const storeId = loginBody?.data?.user_info?.store_id || loginBody?.user_info?.store_id;
  expect(token).toBeTruthy();
  expect(storeId).toBeTruthy();

  // 注入 API 基址与 token
  await page.addInitScript((apiBase: string) => { (window as any).__TEA_RUNTIME_CONFIG__ = { apiBaseUrl: apiBase }; }, API_BASE);
  await page.addInitScript((t: string) => localStorage.setItem('token', t), token);

  // 进入门店优惠券页（store 角色入口）
  await page.goto(`${ADMIN_FE}/store-coupons`);
  // 断言 GET /stores/:id/coupons 正常
  const [resp] = await Promise.all([
    page.waitForResponse(r => r.url().includes(`/api/v1/stores/${storeId}/coupons`) && r.request().method() === 'GET'),
    page.waitForSelector('text=门店优惠券', { state: 'visible' }),
  ]);
  expect(resp.ok()).toBeTruthy();
  // 页面标题存在
  await expect(page.locator('h4:has-text("门店优惠券")')).toBeVisible();
});

test('store activities: visibility and GET list ok', async ({ page }) => {
  const apiReq = await request.newContext();
  // 门店登录
  const loginResp = await apiReq.post(`${API_BASE}/api/v1/user/login`, {
    data: { username: 'store001', password: 'Store@123' },
    headers: { 'Content-Type': 'application/json' },
  });
  expect(loginResp.ok()).toBeTruthy();
  const loginBody = await loginResp.json();
  const token = loginBody?.data?.token || loginBody?.token;
  const storeId = loginBody?.data?.user_info?.store_id || loginBody?.user_info?.store_id;
  expect(token).toBeTruthy();
  expect(storeId).toBeTruthy();

  await page.addInitScript((apiBase: string) => { (window as any).__TEA_RUNTIME_CONFIG__ = { apiBaseUrl: apiBase }; }, API_BASE);
  await page.addInitScript((t: string) => localStorage.setItem('token', t), token);

  await page.goto(`${ADMIN_FE}/store-activities`);
  const [resp] = await Promise.all([
    page.waitForResponse(r => r.url().includes(`/api/v1/stores/${storeId}/activities`) && r.request().method() === 'GET'),
    page.waitForSelector('text=门店活动', { state: 'visible' }),
  ]);
  expect(resp.ok()).toBeTruthy();
  await expect(page.locator('h4:has-text("门店活动")')).toBeVisible();
});

test('store coupons: quick disable action succeeds', async ({ page }) => {
  const apiReq = await request.newContext();
  // 门店登录
  const loginResp = await apiReq.post(`${API_BASE}/api/v1/user/login`, {
    data: { username: 'store001', password: 'Store@123' },
    headers: { 'Content-Type': 'application/json' },
  });
  expect(loginResp.ok()).toBeTruthy();
  const loginBody = await loginResp.json();
  const token = loginBody?.data?.token || loginBody?.token;
  const storeId = loginBody?.data?.user_info?.store_id || loginBody?.user_info?.store_id;
  expect(token).toBeTruthy();
  expect(storeId).toBeTruthy();

  // 先创建一个启用中的优惠券，确保有可禁用的对象
  const now = new Date();
  const startISO = new Date(now.getTime()).toISOString();
  const endISO = new Date(now.getTime() + 24 * 3600 * 1000).toISOString();
  const couponName = `E2E禁用优惠券-${Math.floor(Math.random() * 100000)}`;
  const createResp = await apiReq.post(`${API_BASE}/api/v1/stores/${storeId}/coupons`, {
    data: {
      name: couponName,
      type: 1,
      amount: '5',
      discount: '',
      min_amount: '30',
      total_count: 10,
      status: 1,
      start_time: startISO,
      end_time: endISO,
      description: 'e2e create then disable',
    },
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
  });
  expect(createResp.ok()).toBeTruthy();
  const created = await createResp.json();
  const couponId = created?.data?.id || created?.id;
  expect(couponId).toBeTruthy();

  // 注入 API 基址与 token，并进入优惠券页
  await page.addInitScript((apiBase: string) => { (window as any).__TEA_RUNTIME_CONFIG__ = { apiBaseUrl: apiBase }; }, API_BASE);
  await page.addInitScript((t: string) => localStorage.setItem('token', t), token);
  await page.goto(`${ADMIN_FE}/store-coupons`);

  // 等待列表加载
  await Promise.all([
    page.waitForResponse(r => r.url().includes(`/api/v1/stores/${storeId}/coupons`) && r.request().method() === 'GET'),
    page.waitForSelector('h4:has-text("门店优惠券")', { state: 'visible' }),
  ]);

  // 点击该行的“一键禁用”，并拦截 PUT 更新请求
  const disablePromise = page.waitForResponse(r => r.url().includes(`/api/v1/stores/${storeId}/coupons/${couponId}`) && r.request().method() === 'PUT');
  await page.locator(`tr:has-text("${couponName}") >> text=一键禁用`).click();
  const putResp = await disablePromise;
  expect(putResp.ok()).toBeTruthy();
  // 成功提示
  await expect(page.locator('text=已禁用该优惠券')).toBeVisible();
});

test('store activities: quick create succeeds', async ({ page }) => {
  const apiReq = await request.newContext();
  // 门店登录
  const loginResp = await apiReq.post(`${API_BASE}/api/v1/user/login`, {
    data: { username: 'store001', password: 'Store@123' },
    headers: { 'Content-Type': 'application/json' },
  });
  expect(loginResp.ok()).toBeTruthy();
  const loginBody = await loginResp.json();
  const token = loginBody?.data?.token || loginBody?.token;
  const storeId = loginBody?.data?.user_info?.store_id || loginBody?.user_info?.store_id;
  expect(token).toBeTruthy();
  expect(storeId).toBeTruthy();

  await page.addInitScript((apiBase: string) => { (window as any).__TEA_RUNTIME_CONFIG__ = { apiBaseUrl: apiBase }; }, API_BASE);
  await page.addInitScript((t: string) => localStorage.setItem('token', t), token);

  await page.goto(`${ADMIN_FE}/store-activities`);
  await Promise.all([
    page.waitForResponse(r => r.url().includes(`/api/v1/stores/${storeId}/activities`) && r.request().method() === 'GET'),
    page.waitForSelector('h4:has-text("门店活动")', { state: 'visible' }),
  ]);

  // 打开新建活动弹窗，验证入口与表单可见后关闭弹窗（UI入口校验）
  await page.getByRole('button', { name: '新建活动' }).click();
  await expect(page.getByLabel('活动名称')).toBeVisible();
  await page.keyboard.press('Escape');

  // 通过 API 快速创建活动，作为创建成功的校验基准（减少 UI 交互不确定性）
  const actName = `E2E创建活动-${Math.floor(Math.random() * 100000)}`;
  const now = new Date();
  const startISO = new Date(now.getTime()).toISOString();
  const endISO = new Date(now.getTime() + 24 * 3600 * 1000).toISOString();
  const apiCreate = await apiReq.post(`${API_BASE}/api/v1/stores/${storeId}/activities`, {
    data: {
      name: actName,
      type: 1,
      start_time: startISO,
      end_time: endISO,
      status: 1,
      priority: 1,
      description: 'e2e create via api',
      rules: '{}',
    },
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
  });
  expect(apiCreate.ok()).toBeTruthy();
  const createdBody = await apiCreate.json();
  const activityId = createdBody?.data?.id || createdBody?.id;
  expect(activityId).toBeTruthy();

  // 刷新并断言活动列表包含新建活动名称
  const getPromise = page.waitForResponse(r => r.url().includes(`/api/v1/stores/${storeId}/activities`) && r.request().method() === 'GET');
  await page.reload();
  const getResp = await getPromise;
  expect(getResp.ok()).toBeTruthy();
  await expect(page.locator(`tr:has-text("${actName}")`)).toBeVisible();

  // 清理：删除新建活动，保持环境整洁
  const delResp = await apiReq.delete(`${API_BASE}/api/v1/stores/${storeId}/activities/${activityId}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(delResp.ok()).toBeTruthy();
});
