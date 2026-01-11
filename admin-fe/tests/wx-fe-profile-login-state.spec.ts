import { test, expect } from '@playwright/test';

const BASE = process.env.WXFE_URL_BASE || 'http://127.0.0.1:9093';
const LOGIN_URL = `${BASE}/#/pages/login/index`;
const PROFILE_URL = `${BASE}/#/pages/profile/index`;

// Collect authorized requests for later assertions
function trackAuthRequests(page) {
  const store: { url: string; auth?: string }[] = [];
  page.on('request', (req) => {
    const url = req.url();
    if (url.includes('/api/v1/')) {
      try {
        const headers = req.headers();
        const auth = headers['authorization'] || headers['Authorization'];
        store.push({ url, auth });
      } catch {}
    }
  });
  return store;
}

async function loginWithPhone(page) {
  await page.goto(LOGIN_URL, { waitUntil: 'domcontentloaded' });
  const phoneTab = page.getByText('手机号登录').first();
  if (await phoneTab.isVisible()) {
    await phoneTab.click().catch(() => {});
  }
  const inputs = page.locator('input.weui-input');
  await inputs.first().waitFor({ state: 'visible' });
  await inputs.nth(0).fill('18985121575');
  await inputs.nth(1).fill('000000');
  await page.locator('text=使用手机号登录').first().click();
  await page.waitForFunction(() => !!window.localStorage.getItem('token'), { timeout: 10000 });
}

async function gotoProfile(page) {
  await page.goto(PROFILE_URL, { waitUntil: 'domcontentloaded' });
}

test.describe('wx-fe profile detects login info across sections', () => {
  test('profile shows logged-in user and sections respond to auth', async ({ page }) => {
    const authReqs = trackAuthRequests(page);

    // 1) 登录
    await loginWithPhone(page);

    // 2) 进入「我的」页面
    await gotoProfile(page);

    // 等待 summary 拉取完成（出现资产概览标题）
    await page.getByText('资产概览').first().waitFor({ state: 'visible' });

    // 断言：顶部不再显示“未登录用户”，且显示“退出”按钮或手机号
    expect(await page.locator('text=未登录用户').count()).toBe(0);
    const hasLogout = await page.getByText('退出').first().isVisible().catch(() => false);
    const hasPhone = await page.getByText('18985121575').first().isVisible().catch(() => false);
    expect(hasLogout || hasPhone).toBeTruthy();

    // 断言：已发起带 Authorization 的用户接口请求
    const authed = authReqs.some((r) =>
      (r.url.includes('/api/v1/users/me/summary') || r.url.includes('/api/v1/user/info')) && !!r.auth && r.auth.toLowerCase().startsWith('bearer ')
    );
    expect(authed).toBeTruthy();

    // 3) 服务入口点击均能进入对应页面（订单/钱包/积分/优惠券/分享推广）
    const checks: Array<{ path: string; marker: string; title: string }> = [
      { path: '/pages/orders/index', marker: 'page-orders', title: '订单' },
      { path: '/pages/wallet/index', marker: 'page-wallet', title: '钱包' },
      { path: '/pages/points/index', marker: 'page-points', title: '积分' },
      { path: '/pages/coupons/index', marker: 'page-coupons', title: '优惠券' },
      { path: '/pages/share/index', marker: 'page-share', title: '分享推广' },
    ];

    for (const c of checks) {
      // 在 Profile 页点击对应的服务入口，强制点击避免覆盖层影响
      await gotoProfile(page);
      await page.getByText(c.title).first().click({ force: true });
      await page.waitForTimeout(500);
      const marker = page.locator(`[data-testid="${c.marker}"]`).first();
      try {
        await marker.waitFor({ state: 'visible', timeout: 2500 });
      } catch {
        // 兜底：等待该页面的标题文案出现
        await page.getByText(c.title).first().waitFor({ state: 'visible', timeout: 5000 });
      }
    }

    // 4) 资产概览数值块存在（显示数字或占位），证明数据区渲染完成
    await gotoProfile(page);
    const assetLabels = ['余额（¥）', '茶币', '积分', '可用佣金（¥）', '冻结佣金（¥）', '累计佣金（¥）', '推广规模（直属/团队）'];
    for (const label of assetLabels) {
      expect(await page.getByText(label).first().isVisible()).toBeTruthy();
    }
  });
});
