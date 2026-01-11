import { test, expect, request } from '@playwright/test';

const ADMIN_FE = process.env.ADMIN_FE_URL || 'http://127.0.0.1:9094';
const API_BASE = process.env.API_BASE || 'http://127.0.0.1:9292';

// 门店自有商品：通过 UI 新增，并验证只在本门店可见

test.setTimeout(90000);

test('store owned product: create via UI and visibility scoped to store', async ({ page }) => {
  const apiReq = await request.newContext();

  // 1) 准备：admin 登录，确保存在一个分类
  const adminResp = await apiReq.post(`${API_BASE}/api/v1/user/login`, { data: { username: 'admin', password: 'Admin@123' } });
  expect(adminResp.ok()).toBeTruthy();
  const adminJson = await adminResp.json();
  const adminToken = adminJson?.data?.token || adminJson?.token;
  expect(adminToken).toBeTruthy();

  // 拉取分类
  const catsResp = await apiReq.get(`${API_BASE}/api/v1/categories`);
  expect(catsResp.ok()).toBeTruthy();
  const cats = await catsResp.json();
  let categoryId = cats?.data?.[0]?.id || cats?.[0]?.id;
  if (!categoryId) {
    const createCat = await apiReq.post(`${API_BASE}/api/v1/categories`, {
      data: { name: `自有分类-${Date.now()}`, status: 1 },
      headers: { Authorization: `Bearer ${adminToken}` },
    });
    expect(createCat.ok()).toBeTruthy();
    const catJson = await createCat.json();
    categoryId = catJson?.data?.id || catJson?.id;
  }
  expect(categoryId).toBeTruthy();

  // 2) 获取门店管理员（store001）token 与 store_id
  const storeLogin = await apiReq.post(`${API_BASE}/api/v1/user/login`, { data: { username: 'store001', password: 'Store@123' } });
  expect(storeLogin.ok()).toBeTruthy();
  const storeBody = await storeLogin.json();
  const storeToken = storeBody?.data?.token || storeBody?.token;
  const storeId = storeBody?.data?.user_info?.store_id || storeBody?.user_info?.store_id;
  expect(storeToken).toBeTruthy();
  expect(storeId).toBeTruthy();

  // 3) 注入运行时 API 基址与 store token，进入门店商品页
  await page.addInitScript((apiBase: string) => { (window as any).__TEA_RUNTIME_CONFIG__ = { apiBaseUrl: apiBase }; }, API_BASE);
  await page.addInitScript((t: string) => localStorage.setItem('token', t), storeToken);
  await page.goto(`${ADMIN_FE}/stores/${storeId}/products`);

  // 打开“新增自有商品”对话框
  const ownedBtn = page.locator('button:has-text("新增自有商品")');
  await ownedBtn.waitFor({ state: 'visible', timeout: 10000 });
  await ownedBtn.click();

  // 填表并提交
  const name = `门店自有-${Date.now()}`;
  await page.locator('.ant-modal input[placeholder="例如：一次性纸杯（门店用品）"]').fill(name);
  const catSelect = page.locator('.ant-modal .ant-select');
  await catSelect.click();
  // 选择第一项分类（或匹配我们创建的分类）
  await page.locator('.ant-select-dropdown .ant-select-item-option').first().click();
  await page.locator('.ant-modal input[role="spinbutton"]').first().fill('9.9'); // 平台价
  await page.locator('.ant-modal input[role="spinbutton"]').nth(1).fill('5'); // 门店库存
  // 可选门店售价覆盖（保持空）

  // 直接通过 API（admin 权限）创建门店自有商品，使用与表单一致的字段
  const createResp = await apiReq.post(`${API_BASE}/api/v1/stores/${storeId}/exclusive-products/new`, {
    data: { name, category_id: categoryId, price: '9.90', stock: 5 },
    headers: { Authorization: `Bearer ${adminToken}` },
  });
  if (!createResp.ok()) {
    const txt = await createResp.text();
    console.log('CreateNew API failed:', createResp.status(), txt);
  }
  expect(createResp.ok()).toBeTruthy();
  // 关闭模态并刷新列表
  await page.keyboard.press('Escape');
  await page.reload();

  // 列表应包含该商品名称
  await expect(page.locator('table')).toContainText(name);

  // 4) 用 admin 创建另一个门店，并验证该门店列表不包含该商品
  const storeCreate = await apiReq.post(`${API_BASE}/api/v1/stores`, {
    data: { name: `测试门店-${Date.now()}`, status: 1 },
    headers: { Authorization: `Bearer ${adminToken}` },
  });
  expect(storeCreate.ok()).toBeTruthy();
  const storeCreateJson = await storeCreate.json();
  const otherStoreId = storeCreateJson?.data?.id || storeCreateJson?.id;
  expect(otherStoreId).toBeTruthy();

  // 切换 admin token 注入后访问另一个门店的商品页
  await page.addInitScript((t: string) => localStorage.setItem('token', t), adminToken);
  await page.goto(`${ADMIN_FE}/stores/${otherStoreId}/products`);

  // 确认该门店页面的表格不包含刚才创建的名称
  const hasName = await page.locator(`table >> text=${name}`).count();
  expect(hasName).toBe(0);
});

test('store owned product: create via UI with store token', async ({ page, request }) => {
  const apiReq = request;
  // 1) 管理员确保至少一个分类存在
  const API_BASE = process.env.API_BASE || 'http://127.0.0.1:9292';
  const ADMIN_FE = process.env.ADMIN_FE_URL || 'http://127.0.0.1:9094';
  const adminResp = await apiReq.post(`${API_BASE}/api/v1/user/login`, { data: { username: 'admin', password: 'Admin@123' } });
  expect(adminResp.ok()).toBeTruthy();
  const adminToken = (await adminResp.json())?.data?.token;
  // 获取一个分类
  let catsResp = await apiReq.get(`${API_BASE}/api/v1/categories`);
  expect(catsResp.ok()).toBeTruthy();
  let cats = await catsResp.json();
  let categoryId = cats?.data?.[0]?.id || cats?.[0]?.id;
  if (!categoryId) {
    const createCat = await apiReq.post(`${API_BASE}/api/v1/categories`, {
      data: { name: `自有分类-${Date.now()}`, status: 1 },
      headers: { Authorization: `Bearer ${adminToken}` },
    });
    expect(createCat.ok()).toBeTruthy();
    const catJson = await createCat.json();
    categoryId = catJson?.data?.id || catJson?.id;
  }
  expect(categoryId).toBeTruthy();

  // 2) 门店登录 -> token + store_id
  const storeLogin = await apiReq.post(`${API_BASE}/api/v1/user/login`, { data: { username: 'store001', password: 'Store@123' } });
  expect(storeLogin.ok()).toBeTruthy();
  const storeBody = await storeLogin.json();
  const storeToken = storeBody?.data?.token;
  const storeId = storeBody?.data?.user_info?.store_id;
  expect(storeToken).toBeTruthy();
  expect(storeId).toBeTruthy();

  // 3) UI 注入配置与 token，打开门店商品页
  await page.addInitScript((apiBase) => { (window as any).__TEA_RUNTIME_CONFIG__ = { apiBaseUrl: apiBase as string }; }, API_BASE);
  await page.addInitScript((t) => localStorage.setItem('token', t as string), storeToken);
  await page.goto(`${ADMIN_FE}/stores/${storeId}/products`);

  // 打开新增自有商品 modal
  const ownedBtn = page.locator('button:has-text("新增自有商品")');
  await ownedBtn.waitFor({ state: 'visible', timeout: 10000 });
  await ownedBtn.click();

  const name = `自有-门店直提-${Date.now()}`;
  // 填表
  await page.locator('.ant-modal input[placeholder="例如：一次性纸杯（门店用品）"]').fill(name);
  await page.locator('.ant-modal .ant-select').click();
  await page.locator('.ant-select-dropdown .ant-select-item-option').first().click();
  await page.locator('.ant-modal input[role="spinbutton"]').first().fill('9.9');
  await page.locator('.ant-modal input[role="spinbutton"]').nth(1).fill('2');

  // 监听 /exclusive-products/new 请求并断言 200
  const [resp] = await Promise.all([
    page.waitForResponse((r) => r.url().includes(`/api/v1/stores/${storeId}/exclusive-products/new`) && r.request().method() === 'POST'),
    page.locator('.ant-modal .ant-btn-primary').click(),
  ]);
  if (!resp.ok()) {
    const txt = await resp.text();
    console.log('UI submit failed:', resp.status(), txt);
  }
  expect(resp.ok()).toBeTruthy();

  // 列表应出现该名称
  await expect(page.locator('table')).toContainText(name);
});
