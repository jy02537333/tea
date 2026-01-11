// Simple Puppeteer script to screenshot refund timeline on order-detail page
const fs = require('fs');
const path = require('path');
const { chromium } = require('playwright');

async function main() {
  const outputRoot = 'dist-fixed';
  const dist = path.resolve(__dirname, `../../wx-fe/${outputRoot}`);
  const serverScript = path.resolve(__dirname, './static-server.js');
  const screenshotDir = path.resolve(__dirname, '../../build-ci-logs/screenshots');
  if (!fs.existsSync(screenshotDir)) fs.mkdirSync(screenshotDir, { recursive: true });

  const tokenPath = path.resolve(__dirname, '../../build-ci-logs/local_api_token.txt');
  const orderInfoPath = path.resolve(__dirname, '../../build-ci-logs/payment_order_after_callback.json');
  const token = fs.existsSync(tokenPath) ? fs.readFileSync(tokenPath, 'utf8').trim() : '';
  const orderData = fs.existsSync(orderInfoPath) ? JSON.parse(fs.readFileSync(orderInfoPath, 'utf8')) : {};
  const orderId = (orderData && orderData.data && orderData.data.order && orderData.data.order.id) || 1;

  // Start static server for H5 dist on port 10112 (avoid conflicts)
  const { spawn } = require('child_process');
  const altScript = path.resolve(__dirname, './static-server-10112.js');
  fs.writeFileSync(altScript, fs.readFileSync(serverScript, 'utf8').replace('const port = 10111;', 'const port = 10112;'));
  const server = spawn('node', [altScript], { stdio: 'inherit' });
  await new Promise((r) => setTimeout(r, 1200));
  const serverPort = 10112;

  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  const url = `http://127.0.0.1:${serverPort}/pages/order-detail/index.html?id=${orderId}&mock_refund=1&tk=${encodeURIComponent(token)}`;
  await page.goto(url, { waitUntil: 'networkidle' });

  // Wait for refund card
  await page.waitForTimeout(1200);
  const file = path.join(screenshotDir, 'refund_timeline_order_detail.png');
  await page.screenshot({ path: file, fullPage: true });

  await browser.close();
  try { server.kill('SIGTERM'); } catch(_) {}
  console.log('Saved screenshot:', file);
}

main().catch((err) => {
  console.error('screenshot failed', err);
  process.exit(1);
});
