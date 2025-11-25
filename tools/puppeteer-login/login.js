const puppeteer = require('puppeteer');
const path = require('path');

// 配置：根据你的本地前端静态服务器地址调整
const ADMIN_FE_URL = process.env.ADMIN_FE_URL || 'http://localhost:8000/index.html';
const USERNAME = process.env.ADMIN_USERNAME || 'admin';
const PASSWORD = process.env.ADMIN_PASSWORD || 'pass';

(async () => {
  const browser = await puppeteer.launch({ headless: true });
  const page = await browser.newPage();
  page.on('console', msg => console.log('PAGE:', msg.text()));

  console.log('打开页面:', ADMIN_FE_URL);
  await page.goto(ADMIN_FE_URL, { waitUntil: 'networkidle2' });

  // 打开登录模态
  await page.evaluate(() => { if (typeof openLoginModal === 'function') openLoginModal(); });
  await page.waitForTimeout(300);

  // 填写用户名/密码并提交
  await page.type('#login-username', USERNAME);
  await page.type('#login-password', PASSWORD);
  await page.click('.modal-content .btn-add');

  // 等待登录完成：观察 localStorage 的变化
  await page.waitForFunction(() => !!localStorage.getItem('tea_admin_token'), { timeout: 5000 })
    .catch(() => {});

  const token = await page.evaluate(() => localStorage.getItem('tea_admin_token'));
  const role = await page.evaluate(() => localStorage.getItem('admin_role'));
  const name = await page.evaluate(() => localStorage.getItem('admin_name'));

  console.log('=== 登录结果 ===');
  console.log('token:', token ? token.slice(0, 40) + '...' : '(none)');
  console.log('role:', role);
  console.log('name:', name);

  // 可选：测试侧栏菜单是否加载（如果页面有 fetchAndRenderMenus 函数）
  const menus = await page.evaluate(async () => {
    try {
      if (typeof fetchAndRenderMenus === 'function') {
        // 触发一次菜单加载并等待短暂时间
        await fetchAndRenderMenus();
        await new Promise(r => setTimeout(r, 500));
        // 尝试从 DOM 中读取菜单项文本
        const items = Array.from(document.querySelectorAll('#menu-root .menu-item')).map(el => el.textContent.trim());
        return { fromDom: items };
      }
      return { fromDom: null };
    } catch (e) {
      return { error: e.message };
    }
  });

  console.log('menus:', menus);

  await browser.close();
})();
