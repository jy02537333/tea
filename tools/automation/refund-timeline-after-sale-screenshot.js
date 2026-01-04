const fs = require('fs');
const path = require('path');
const { chromium } = require('playwright');

async function main() {
  const serverScript = path.resolve(__dirname, './static-server.js');
  const screenshotDir = path.resolve(__dirname, '../../build-ci-logs/screenshots');
  if (!fs.existsSync(screenshotDir)) fs.mkdirSync(screenshotDir, { recursive: true });

  const tokenPath = path.resolve(__dirname, '../../build-ci-logs/local_api_token.txt');
  const token = fs.existsSync(tokenPath) ? fs.readFileSync(tokenPath, 'utf8').trim() : '';

  // Start static server on 10112 to avoid conflicts
  const { spawn } = require('child_process');
  const altScript = path.resolve(__dirname, './static-server-10112.js');
  fs.writeFileSync(altScript, fs.readFileSync(serverScript, 'utf8').replace('const port = 10111;', 'const port = 10112;'));
  const server = spawn('node', [altScript], { stdio: 'inherit' });
  await new Promise((r) => setTimeout(r, 1200));
  const serverPort = 10112;

  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  const url = `http://127.0.0.1:${serverPort}/pages/after-sale/index.html?mock_refund=1&tk=${encodeURIComponent(token)}`;
  await page.goto(url, { waitUntil: 'networkidle' });

  // Wait for cards
  await page.waitForTimeout(1200);
  const file = path.join(screenshotDir, 'refund_timeline_after_sale.png');
  await page.screenshot({ path: file, fullPage: true });

  await browser.close();
  try { server.kill('SIGTERM'); } catch(_) {}
  console.log('Saved screenshot:', file);
}

main().catch((err) => {
  console.error('screenshot failed', err);
  process.exit(1);
});
