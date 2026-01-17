import { test, expect, request } from '@playwright/test';

const ADMIN_FE = process.env.ADMIN_FE_URL || 'http://127.0.0.1:9094';
const API_BASE = process.env.API_BASE || 'http://127.0.0.1:9292';

// 门店收款账户：通过 UI 新增并删除（门店管理员 role=store）

test.setTimeout(60000);

test('store accounts: create and delete via UI', async ({ page }) => {
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

  // 进入门店收款账户页
  await page.goto(`${ADMIN_FE}/store-accounts`);

  // 打开新增收款账户弹窗
  const createBtn = page.locator('button:has-text("新建收款账户")');
  await createBtn.waitFor({ state: 'visible', timeout: 10000 });
  await createBtn.click();

  const name = `测试账户-${Date.now()}`;
  const no = `${Math.floor(Math.random() * 1e12)}`; // 模拟账号/收款号

  // 填写表单（账户类型默认银行卡，可保持默认）
  await page.locator('.ant-modal input[placeholder="开户名，例如：杭州茶心阁门店"]').fill(name);
  await page.locator('.ant-modal input[placeholder="银行卡号 / 支付宝账号 / 微信收款码标识等"]').fill(no);
  await page.locator('.ant-modal input[placeholder="如：中国银行杭州分行 / 支付宝 / 微信支付"]').fill('中国银行杭州分行');
  // 设为默认账户（可选）
  const defaultSwitch = page.locator('.ant-modal .ant-switch');
  await defaultSwitch.first().click();

  // 监听创建接口请求
  const createRespPromise = page.waitForResponse(r => r.url().includes(`/api/v1/stores/${storeId}/accounts`) && r.request().method() === 'POST');
  await page.locator('.ant-modal .ant-btn-primary').click({ force: true });
  const createResp = await createRespPromise;
  expect(createResp.ok()).toBeTruthy();
  await expect(page.locator('.ant-message')).toContainText('收款账户已创建');

  // 列表应出现新记录
  const row = page.locator(`table >> tr:has-text("${name}")`).first();
  await row.waitFor({ state: 'visible', timeout: 10000 });

  // 再创建一个非默认账户用于过滤断言
  await createBtn.click();
  const name2 = `测试账户-非默认-${Date.now()}`;
  const no2 = `${Math.floor(Math.random() * 1e12)}`;
  await page.locator('.ant-modal input[placeholder="开户名，例如：杭州茶心阁门店"]').fill(name2);
  await page.locator('.ant-modal input[placeholder="银行卡号 / 支付宝账号 / 微信收款码标识等"]').fill(no2);
  await page.locator('.ant-modal input[placeholder="如：中国银行杭州分行 / 支付宝 / 微信支付"]').fill('支付宝');
  // 保持“设为默认账户”关闭状态
  const createRespPromise2 = page.waitForResponse(r => r.url().includes(`/api/v1/stores/${storeId}/accounts`) && r.request().method() === 'POST');
  await page.locator('.ant-modal .ant-btn-primary').click({ force: true });
  const createResp2 = await createRespPromise2;
  expect(createResp2.ok()).toBeTruthy();
  await expect(page.locator('.ant-message')).toContainText('收款账户已创建');
  const row2 = page.locator(`table >> tr:has-text("${name2}")`).first();
  await row2.waitFor({ state: 'visible', timeout: 10000 });

  // 过滤：开启“只看默认账户”，默认账户应可见，非默认账户不可见
  const filterSwitch = page.locator('div.ant-card:has-text("只看默认账户") .ant-switch').first();
  await filterSwitch.click();
  await expect(page.locator('table')).toContainText(name);
  const nonDefaultVisible = await page.locator(`table >> text=${name2}`).count();
  expect(nonDefaultVisible).toBe(0);
  // 关闭过滤，两个账户都应可见
  await filterSwitch.click();
  await expect(page.locator('table')).toContainText(name);
  await expect(page.locator('table')).toContainText(name2);

  // 编辑默认账户：修改银行名称并断言更新消息
  const editBtn = row.locator('text=编辑').first();
  await editBtn.click();
  const bankInput = page.locator('.ant-modal input[placeholder="如：中国银行杭州分行 / 支付宝 / 微信支付"]');
  await bankInput.fill('中国银行西湖支行');
  const updateRespPromise = page.waitForResponse(r => r.url().includes(`/api/v1/stores/${storeId}/accounts/`) && r.request().method() === 'PUT');
  await page.locator('.ant-modal .ant-btn-primary').click({ force: true });
  const updateResp = await updateRespPromise;
  expect(updateResp.ok()).toBeTruthy();
  await expect(page.locator('.ant-message')).toContainText('收款账户已更新');
  await expect(row).toContainText('中国银行西湖支行');

  // 结束：保留账户以便后续测试复用（删除操作另行覆盖）

  if (enableTrace) {
    await ctx.tracing.stop({ path: '/home/frederic/project/tea/build-ci-logs/playwright/store-accounts-trace.zip' });
  }
});
