const { chromium } = require('playwright');
const { spawn } = require('child_process');
const fs = require('fs');
const path = require('path');

async function ensureDir(p) {
  await fs.promises.mkdir(p, { recursive: true }).catch(() => {});
}

async function main() {
  const outputDir = path.resolve(__dirname, '../../build-ci-logs/screenshots');
  await ensureDir(outputDir);
  const outPath = path.join(outputDir, 'admin_store_finance.png');

  const server = spawn('node', [path.resolve(__dirname, './admin-static-server.js')], {
    stdio: 'inherit',
  });
  // wait a moment for server to start
  await new Promise((r) => setTimeout(r, 800));

  const browser = await chromium.launch();
  const context = await browser.newContext({ viewport: { width: 1360, height: 900 } });
  const page = await context.newPage();
  const url = 'http://127.0.0.1:10113/store-finance?tk=dummy';
  await page.goto(url, { waitUntil: 'networkidle' });

  // Try to wait for key UI parts to appear (non-fatal timeouts)
  try {
    await page.waitForSelector('text=门店财务提现', { timeout: 2000 });
  } catch {}
  try {
    await page.waitForSelector('text=资金流水（支付/退款/提现）', { timeout: 2000 });
  } catch {}

  await page.screenshot({ path: outPath, fullPage: true });
  await browser.close();

  // Stop server
  server.kill('SIGTERM');
  console.log('Saved screenshot:', outPath);
}

main().catch((err) => {
  console.error('screenshot failed:', err);
  process.exit(1);
});
