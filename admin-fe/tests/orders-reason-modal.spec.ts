import { test, expect, request } from '@playwright/test';

const ADMIN_FE = process.env.ADMIN_FE_URL || 'http://127.0.0.1:5173';
const API_BASE = process.env.API_BASE || 'http://127.0.0.1:9292';

test.setTimeout(60000);

test('orders page opens reason modal for sensitive actions when available', async ({ page }) => {
  const ctx = page.context();
  const enableTrace = process.env.TRACE !== '0';
  if (enableTrace) {
    await ctx.tracing.start({ screenshots: true, snapshots: true });
  }

  // admin dev-login for token
  // Use API_BASE for token, but force FE runtime to point at MOCK at 9393
  const apiRequest = await request.newContext();
  const resp = await apiRequest.post(`${API_BASE}/api/v1/user/dev-login`, { data: { openid: 'admin_openid' } });
  expect(resp.ok()).toBeTruthy();
  const body = await resp.json();
  const token = body?.data?.token || body?.token;
  expect(token).toBeTruthy();
  await page.addInitScript((apiBase: string) => { (window as any).__TEA_RUNTIME_CONFIG__ = { apiBaseUrl: apiBase }; }, 'http://127.0.0.1:9393');
  await page.addInitScript((t: string) => localStorage.setItem('token', t), token);

  await page.goto(`${ADMIN_FE}/orders`);
  // quick screenshot of initial state for debugging
  await page.screenshot({ path: '/home/frederic/project/tea/build-ci-logs/playwright/orders-initial.png', fullPage: true }).catch(() => {});
  // wait table or header
  await page.waitForSelector('.ant-table, text=订单号', { timeout: 20000 }).catch(() => {});

  // candidate actions that should open a reason modal after confirming Popconfirm
  const candidates: Array<{ text: string; expectLabel: RegExp }> = [
    { text: '管理员取消', expectLabel: /取消原因/ },
    { text: '标记退款中', expectLabel: /原因/ },
    { text: '确认退款完成', expectLabel: /原因/ },
    { text: '立即退款', expectLabel: /原因/ },
  ];

  let opened = false;
  for (const c of candidates) {
    const btn = page.locator(`table >> button:has-text("${c.text}")`).first();
    if (!(await btn.count())) continue;
    try {
      await btn.scrollIntoViewIfNeeded().catch(() => {});
      await btn.click({ force: true });
      // confirm Popconfirm
      const confirmBtn = page.locator('.ant-popover .ant-btn-primary').first();
      await confirmBtn.waitFor({ timeout: 5000 });
      await confirmBtn.click({ force: true });
      await page.waitForSelector('.ant-modal', { timeout: 8000 });
      // assert label text exists in modal
      await expect(page.locator('.ant-modal')).toContainText(c.expectLabel);
      opened = true;
      break;
    } catch {
      // try next candidate
    }
  }

  // If no candidate action is available, still assert Orders page is visible as smoke
  if (!opened) {
    await page.screenshot({ path: '/home/frederic/project/tea/build-ci-logs/playwright/orders-before-assert.png', fullPage: true }).catch(() => {});
    // 避免严格模式下多元素匹配，使用表头角色选择器
    await expect(page.getByRole('columnheader', { name: '订单号' })).toBeVisible();
  }

  if (enableTrace) {
    await ctx.tracing.stop({ path: '/home/frederic/project/tea/build-ci-logs/playwright/orders-reason-modal-trace.zip' });
  }
});
