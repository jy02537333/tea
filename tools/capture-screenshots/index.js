const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

(async () => {
  const outDir = path.resolve(process.cwd(), 'screenshots');
  if (!fs.existsSync(outDir)) fs.mkdirSync(outDir, { recursive: true });

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1280, height: 900 } });
  const page = await context.newPage();

  // Targets can be adjusted
  const targets = [
    { url: 'http://localhost:8000/index.html', name: 'admin_index' },
    { url: 'http://localhost:8000/login.html', name: 'admin_login' },
    { url: 'http://localhost:9292/api/v1/health', name: 'api_health' },
    { url: 'http://localhost:9292/api/v1/admin/menus', name: 'api_admin_menus' }
  ];

  for (const t of targets) {
    try {
      console.log(`Navigating ${t.url}`);
      const response = await page.goto(t.url, { waitUntil: 'networkidle', timeout: 30000 });
      const file = path.join(outDir, `${t.name}.png`);
      await page.screenshot({ path: file, fullPage: true });
      console.log(`Saved ${file} (status: ${response ? response.status() : 'no response'})`);
    } catch (err) {
      console.error(`Failed to capture ${t.url}:`, err.message);
      fs.writeFileSync(path.join(outDir, `${t.name}.error.txt`), String(err.stack || err));
    }
  }

  await browser.close();
  console.log('All done. Screenshots saved to:', outDir);
})();
