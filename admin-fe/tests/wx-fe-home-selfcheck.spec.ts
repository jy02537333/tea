import { test, expect } from '@playwright/test';

const WX_FE_ORIGIN = process.env.WX_FE_ORIGIN || 'http://127.0.0.1:9093';
const WX_FE_INDEX_URL = process.env.WX_FE_INDEX_URL || `${WX_FE_ORIGIN}/#/pages/index/index`;

function fmt(msg: unknown) {
  try {
    if (typeof msg === 'string') return msg;
    return JSON.stringify(msg);
  } catch {
    return String(msg);
  }
}

test('wx-fe index resource self-check has no chunk errors', async ({ page }, testInfo) => {
  const logs: string[] = [];
  const failures: string[] = [];

  page.on('console', (m) => {
    const entry = `[console.${m.type()}] ${m.text()}`;
    logs.push(entry);
  });
  page.on('pageerror', (err) => {
    const entry = `[pageerror] ${err?.name || 'Error'}: ${err?.message || String(err)}`;
    logs.push(entry);
    failures.push(entry);
  });
  page.on('requestfailed', (req) => {
    // requestfailed 通常拿不到可用的 response（即便有，Playwright 的 req.response() 也是 async），这里只记录关键信息。
    const entry = `[requestfailed] ${req.method()} ${req.url()} failure=${req.failure()?.errorText || ''}`.trim();
    logs.push(entry);
    // 页面内的资源自检会用 HEAD 做探测并在失败时回退 GET；HEAD 的 requestfailed 不应当作硬错误。
    if (req.method() === 'HEAD') return;
    // 仅把静态资源 GET 失败视为硬错误
    if (/\/chunk\/|\/js\/|\/css\//.test(req.url())) failures.push(entry);
  });

  await page.goto(WX_FE_INDEX_URL, { waitUntil: 'domcontentloaded' });

  // 页面可能因 runtime error 无法渲染；这里尽量“等一等”但不 hard-fail
  await page
    .locator('[data-testid="page-index"]')
    .first()
    .waitFor({ timeout: 15_000, state: 'attached' })
    .catch(() => {});

  // 给自检逻辑一些时间跑完（即便没渲染出来也能抓到 console/pageerror）
  await page.waitForTimeout(6_000);
  await page.waitForLoadState('networkidle', { timeout: 20_000 }).catch(() => {});

  try {
    await testInfo.attach('wx-fe-index.png', {
      body: await page.screenshot({ fullPage: true }),
      contentType: 'image/png',
    });
  } catch (e) {
    logs.push(`[attach-error] screenshot: ${fmt((e as any)?.message || e)}`);
  }

  const logText = logs.join('\n') + '\n';
  await testInfo.attach('wx-fe-console.log', {
    body: Buffer.from(logText, 'utf-8'),
    contentType: 'text/plain; charset=utf-8',
  });

  // 断言：不出现 ChunkLoadError / private field 解析错误
  const all = logs.join('\n');
  expect(all).not.toMatch(/ChunkLoadError/i);
  expect(all).not.toMatch(/Private field/i);
  expect(all).not.toMatch(/Loading chunk\s+\d+\s+failed/i);
  // 若静态资源请求失败，给出更直观的错误信息
  expect(failures, `Static/network failures detected:\n${failures.join('\n')}`).toEqual([]);
});
