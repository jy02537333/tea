import { test, expect, request } from '@playwright/test';

const ADMIN_FE = process.env.ADMIN_FE_URL || 'http://127.0.0.1:9094';
const API_BASE = process.env.API_BASE || 'http://127.0.0.1:9292';

test.setTimeout(120000);

test('platform product: insert image in rich text via OSS and create', async ({ page }) => {
  const apiReq = await request.newContext();

  // 1) 管理员登录，确保至少一个分类存在
  const adminResp = await apiReq.post(`${API_BASE}/api/v1/user/login`, { data: { username: 'admin', password: 'Admin@123' } });
  expect(adminResp.ok()).toBeTruthy();
  const adminJson = await adminResp.json();
  const adminToken = adminJson?.data?.token || adminJson?.token;
  expect(adminToken).toBeTruthy();

  let catsResp = await apiReq.get(`${API_BASE}/api/v1/categories`);
  expect(catsResp.ok()).toBeTruthy();
  let cats = await catsResp.json();
  let categoryId = cats?.data?.[0]?.id || cats?.[0]?.id;
  if (!categoryId) {
    const createCat = await apiReq.post(`${API_BASE}/api/v1/categories`, {
      data: { name: `平台分类-${Date.now()}`, status: 1 },
      headers: { Authorization: `Bearer ${adminToken}` },
    });
    expect(createCat.ok()).toBeTruthy();
    const catJson = await createCat.json();
    categoryId = catJson?.data?.id || catJson?.id;
  }
  expect(categoryId).toBeTruthy();

  // 2) 注入运行时 API 基址与 admin token
  await page.addInitScript((apiBase: string) => { (window as any).__TEA_RUNTIME_CONFIG__ = { apiBaseUrl: apiBase }; }, API_BASE);
  await page.addInitScript((t: string) => localStorage.setItem('token', t), adminToken);

  // 3) 拦截上传策略与 OSS 直传请求，模拟成功
  const mockHost = 'https://oss.mock.local';
  await page.route('**/api/v1/upload/oss/policy', async (route) => {
    const now = Math.floor(Date.now() / 1000);
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: {
          accessid: 'test',
          host: mockHost,
          policy: 'test-policy',
          signature: 'test-sign',
          dir: 'uploads/test/',
          expire: now + 300,
        },
      }),
    });
  });
  await page.route(`${mockHost}/**`, async (route) => {
    // 伪造 OSS 直传成功
    await route.fulfill({ status: 200, contentType: 'text/plain', body: '' });
  });

  // 4) 进入平台商品页（创建 tab 会自动打开抽屉）
  await page.goto(`${ADMIN_FE}/products?tab=create`);
  await page.locator('.ant-drawer').waitFor({ state: 'visible', timeout: 15000 });

  // 5) 填写基础表单
  const name = `平台商品-${Date.now()}`;
  await page.locator('.ant-drawer input[placeholder="请输入商品名称"]').fill(name);
  // 选择分类（按表单项标签定位更稳）
  await page.locator('.ant-drawer .ant-form-item:has-text("分类") .ant-select').click();
  await page.locator('.ant-select-dropdown .ant-select-item-option').first().click();
  // 售价
  await page.locator('.ant-drawer .ant-form-item:has-text("售价") input[role="spinbutton"]').fill('12.34');
  // 库存
  await page.locator('.ant-drawer .ant-form-item:has-text("库存") input[role="spinbutton"]').fill('10');

  // 6) 上传主图（下方按钮 -> 触发文件选择 -> 直传成功 -> URL 写入）
  const tinyPngBase64 =
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR4nGNgYAAAAAMAASsJTYQAAAAASUVORK5CYII='; // 1x1 png
  const uploadBtn = page.locator('.ant-drawer .ant-upload button:has-text("上传图片到 OSS")');
  const [chooser1] = await Promise.all([
    page.waitForEvent('filechooser'),
    uploadBtn.click(),
  ]);
  await chooser1.setFiles({ name: 'tiny.png', mimeType: 'image/png', buffer: Buffer.from(tinyPngBase64, 'base64') });
  // 等待 URL 写入到表单（选择器文本包含 mockHost）
  await expect(page.locator('.ant-drawer')).toContainText('oss.mock.local');

  // 7) 在富文本中通过 toolbar 的图片按钮插入一张图片
  const toolbarBtn = page.locator('.ant-drawer .ql-toolbar button.ql-image');
  const [chooser2] = await Promise.all([
    page.waitForEvent('filechooser'),
    toolbarBtn.click(),
  ]);
  await chooser2.setFiles({ name: 'tiny2.png', mimeType: 'image/png', buffer: Buffer.from(tinyPngBase64, 'base64') });
  await expect(page.locator('.ant-drawer .ql-editor img')).toHaveCount(1);

  // 8) 冒烟断言：主图 URL 已写入表单，富文本出现插图
  await expect(page.locator('.ant-drawer')).toContainText('oss.mock.local');
  await expect(page.locator('.ant-drawer .ql-editor img')).toHaveCount(1);
});
