#!/usr/bin/env node
/*
 wx-fe feature verification via Playwright
 - Verifies H5 pages render and key UI exist: login, profile ("我的工单"), feedback (Ticket ID modal capable), store-detail (minimal actions)
 - Saves screenshots to build-ci-logs/screenshots and a JSON summary to build-ci-logs/wx-fe-verify-result.json
 - Uses system Chromium if available to avoid downloads
 Env:
   PREVIEW_URL: Base URL of H5 preview (default http://127.0.0.1:10088/)
*/
const fs = require('fs');
const http = require('http');
const path = require('path');
const { chromium } = require('playwright');
const { URL } = require('url');

function startStaticServer(rootDir) {
  return new Promise((resolve, reject) => {
    const server = http.createServer((req, res) => {
      const reqPath = (req.url || '/').split('?')[0];
      const safePath = reqPath.replace(/\0/g, '');
      const candidate = path.join(rootDir, safePath);
      const sendFile = (fp) => fs.readFile(fp, (err, data) => {
        if (err) {
          res.writeHead(404);
          res.end('not found');
          return;
        }
        res.writeHead(200);
        res.end(data);
      });
      if (safePath === '/' || safePath === '') {
        return sendFile(path.join(rootDir, 'index.html'));
      }
      fs.stat(candidate, (err, stats) => {
        if (!err && stats && stats.isFile()) return sendFile(candidate);
        return sendFile(path.join(rootDir, 'index.html'));
      });
    });

    server.on('error', reject);
    server.listen(0, '127.0.0.1', () => {
      const addr = server.address();
      if (!addr || typeof addr === 'string') {
        reject(new Error('failed to get server address'));
        return;
      }
      resolve({ server, url: `http://127.0.0.1:${addr.port}/` });
    });
  });
}

function findBrowserPath() {
  const envPaths = [
    process.env.PLAYWRIGHT_EXECUTABLE_PATH,
    process.env.CHROMIUM_BIN,
    process.env.CHROME_BIN,
    process.env.GOOGLE_CHROME_BIN,
  ].filter(Boolean);
  const candidates = envPaths.concat([
    '/usr/bin/chromium-browser',
    '/usr/bin/chromium',
    '/usr/bin/google-chrome',
    '/usr/bin/google-chrome-stable',
    '/snap/bin/chromium',
  ]);
  for (const p of candidates) {
    try { if (p && fs.existsSync(p)) return p; } catch (_) {}
  }
  return null;
}

function readJSONSafe(p, def = null) {
  try { return JSON.parse(fs.readFileSync(p, 'utf-8')); } catch (_) { return def; }
}

function readEnvFile(p) {
  try {
    const txt = fs.readFileSync(p, 'utf-8');
    const out = {};
    for (const line of txt.split(/\r?\n/)) {
      const m = line.match(/^([A-Z0-9_]+)=(.*)$/);
      if (m) out[m[1]] = m[2];
    }
    return out;
  } catch (_) {
    return {};
  }
}

(async () => {
  const repoRoot = path.resolve(__dirname, '../../');
  const logsDir = path.join(repoRoot, 'build-ci-logs');
  const screenshotsDir = path.join(logsDir, 'screenshots');
  if (!fs.existsSync(screenshotsDir)) fs.mkdirSync(screenshotsDir, { recursive: true });

  const outputRoot = 'dist-fixed';
  const distRoot = path.join(repoRoot, 'wx-fe', outputRoot);

  let previewURL = process.env.PREVIEW_URL || '';
  let localServer = null;
  let browser = null;
  let exitCode = 0;

  if (!previewURL) {
    const indexHtml = path.join(distRoot, 'index.html');
    if (!fs.existsSync(indexHtml)) {
      throw new Error(`PREVIEW_URL not set and ${indexHtml} not found. Build wx-fe first or set PREVIEW_URL.`);
    }
    const started = await startStaticServer(distRoot);
    localServer = started.server;
    previewURL = started.url;
    console.log('[wx-fe-verify] Started local static server at', previewURL, `(root=${outputRoot})`);
  }

  const apiBase = process.env.WX_API_BASE_URL || 'http://127.0.0.1:9292';
  const base = previewURL.replace(/#.*$/, '').replace(/\/?$/, '/');

  const tokenFromEnv = process.env.AUTH_TOKEN || '';
  const tokenFromTokensEnv = readEnvFile(path.join(logsDir, 'tokens.env')).AUTH_TOKEN || '';
  const tokenFromLogs = (readJSONSafe(path.join(logsDir, 'admin_login_response.json'), null)?.data?.token) || '';
  const authToken = String(tokenFromEnv || tokenFromTokensEnv || tokenFromLogs || '').trim();

  function withTkInjected(inputUrl) {
    if (!authToken) return inputUrl;
    try {
      const u = new URL(inputUrl);
      u.searchParams.set('tk', authToken);
      return u.toString();
    } catch (_) {
      return inputUrl;
    }
  }

  function buildH5Url(pathAndQuery) {
    const u = new URL(base);
    if (authToken) u.searchParams.set('tk', authToken);
    const h = String(pathAndQuery || '').replace(/^#/, '');
    u.hash = h.startsWith('/') ? h : `/${h}`;
    return u.toString();
  }

  const executablePath = findBrowserPath();
  const launchOpts = { headless: true, args: ['--no-sandbox', '--disable-setuid-sandbox'] };
  if (executablePath) launchOpts.executablePath = executablePath;
  const result = {
    base: previewURL,
    checks: [],
    debug: {
      apiBase,
      tokenSource: tokenFromEnv ? 'env' : (tokenFromTokensEnv ? 'tokens.env' : (tokenFromLogs ? 'admin_login_response.json' : 'none')),
      authTokenLen: authToken ? authToken.length : 0,
      authTokenHead: authToken ? authToken.slice(0, 16) : '',
    },
  };
  result.apiProbe = {};

  try {
    console.log('[wx-fe-verify] Launching Chromium at', executablePath || '(bundled/default)');
    browser = await chromium.launch(launchOpts);
    const context = await browser.newContext({ viewport: { width: 420, height: 880, deviceScaleFactor: 1 } });
    if (authToken) {
      await context.addInitScript((t) => { try { localStorage.setItem('token', t); } catch(_) {} }, authToken);
    }
    // Ensure H5-originated API calls carry Authorization even if app-side token init is flaky.
    if (authToken) {
      const shouldRewriteDockerInternal = process.env.WX_FE_VERIFY_REWRITE_DOCKER_INTERNAL === '1';
      await context.route('**/api/v1/**', async (route) => {
        try {
          const req = route.request();
          const reqUrl = req.url();
          const headers = { ...req.headers() };
          if (!headers.authorization && !headers.Authorization) {
            headers.authorization = `Bearer ${authToken}`;
          }

          // Some builds bake in host.docker.internal which is not resolvable on Linux.
          // Optional rewrite to apiBase for local verification.
          let newUrl = undefined;
          try {
            const u = new URL(reqUrl);
            if (shouldRewriteDockerInternal && u.hostname === 'host.docker.internal') {
              const target = new URL(apiBase);
              u.protocol = target.protocol;
              u.hostname = target.hostname;
              u.port = target.port;
              newUrl = u.toString();
            }
          } catch (_) {}

          await route.continue(newUrl ? { headers, url: newUrl } : { headers });
        } catch (_) {
          await route.continue();
        }
      });
    }
    const page = await context.newPage();

  async function gotoHash(pathAndQuery) {
    const url = buildH5Url(pathAndQuery);
    await page.goto(url, { waitUntil: 'load' });
    return url;
  }

  async function checkLogin() {
    const url = await gotoHash('/pages/login/index');
    let ok = false; let msg = '';
    try {
      await page.getByText('微信一键登录').waitFor({ timeout: 15000 });
      ok = true;
    } catch (e) {
      msg = 'Login heading not found';
    }
    const shot = path.join(screenshotsDir, 'login.png');
    await page.screenshot({ path: shot, fullPage: true }).catch(() => {});
    result.checks.push({ page: 'login', url, ok, shot, msg });
  }

  async function checkProfile() {
    const url = await gotoHash('/pages/profile/index');
    let ok = false; let msg = '';
    try {
      await page.getByText('我的服务').waitFor({ timeout: 15000 });
      // Verify the presence of "我的工单"
      await page.getByText('我的工单').waitFor({ timeout: 15000 });
      ok = true;
    } catch (e) {
      msg = 'Profile service block or 我的工单 missing';
    }
    const shot = path.join(screenshotsDir, 'profile.png');
    await page.screenshot({ path: shot, fullPage: true }).catch(() => {});
    result.checks.push({ page: 'profile', url, ok, shot, msg });
  }

  async function checkFeedback() {
    const url = await gotoHash('/pages/feedback/index');
    let ok = false; let msg = '';
    try {
      await page.getByText('意见反馈').waitFor({ timeout: 15000 });
      // Button text (Taro Button may not map to semantic role button reliably)
      await page.getByText('提交反馈').waitFor({ timeout: 15000 });
      ok = true;
    } catch (e) {
      msg = 'Feedback page or submit button missing';
    }
    const shot = path.join(screenshotsDir, 'feedback.png');
    await page.screenshot({ path: shot, fullPage: true }).catch(() => {});
    result.checks.push({ page: 'feedback', url, ok, shot, msg });
  }

  async function checkStoreDetail() {
    // Use explicit STORE_ID if provided; else pick from logs if available
    const envStoreId = Number(process.env.STORE_ID || 0);
    const storeInfo = readJSONSafe(path.join(logsDir, 'activity_demo_create_store.json'), null);
    const storeId = envStoreId > 0 ? envStoreId : Number(storeInfo?.data?.id || 0);
    const q = storeId > 0 ? `?store_id=${storeId}` : '';
    const url = await gotoHash('/pages/store-detail/index' + q);
    let ok = false; let msg = '';

    const storeDetailDiagnostics = { uiState: null, bodyTextLen: 0, bodyTextHead: '', consoleErrors: [], requestFailures: [] };
    const onConsole = (m) => {
      try {
        if (m.type && m.type() === 'error') {
          storeDetailDiagnostics.consoleErrors.push({ text: m.text(), location: m.location ? m.location() : undefined });
          if (storeDetailDiagnostics.consoleErrors.length > 20) storeDetailDiagnostics.consoleErrors.shift();
        }
      } catch (_) {}
    };
    const onRequestFailed = (r) => {
      try {
        const f = r.failure && r.failure();
        storeDetailDiagnostics.requestFailures.push({ url: r.url(), errorText: f?.errorText || '' });
        if (storeDetailDiagnostics.requestFailures.length > 20) storeDetailDiagnostics.requestFailures.shift();
      } catch (_) {}
    };
    page.on('console', onConsole);
    page.on('requestfailed', onRequestFailed);

    // API probe for store-detail dependency
    async function probeStoreApi() {
      const probe = { storeId, url: null, ok: false, status: 0, dataPreview: null, error: null };
      try {
        const apiUrl = new URL(apiBase);
        apiUrl.pathname = `/api/v1/stores/${storeId}`;
        probe.url = apiUrl.toString();
        const headers = authToken ? { Authorization: `Bearer ${authToken}` } : {};
        const resp = await fetch(probe.url, { method: 'GET', headers });
        probe.status = resp.status;
        const json = await resp.json().catch(() => null);
        if (resp.ok && json) {
          probe.ok = true;
          const d = json.data || json;
          probe.dataPreview = {
            id: d?.id,
            name: d?.name,
            address: d?.address,
            phone: d?.phone,
          };
        } else {
          probe.error = (json && json.message) || `HTTP ${resp.status}`;
        }
      } catch (e) {
        probe.error = e?.message || String(e);
      }
      result.apiProbe.storeDetail = probe;
    }
    async function probeStoreAccountsApi() {
      const probe = { storeId, url: null, ok: false, status: 0, count: 0, firstId: null, error: null };
      try {
        const apiUrl = new URL(apiBase);
        apiUrl.pathname = `/api/v1/stores/${storeId}/accounts`;
        probe.url = apiUrl.toString();
        const headers = authToken ? { Authorization: `Bearer ${authToken}` } : {};
        const resp = await fetch(probe.url, { method: 'GET', headers });
        probe.status = resp.status;
        const json = await resp.json().catch(() => null);
        if (resp.ok && json) {
          probe.ok = true;
          const list = json.data || json || [];
          if (Array.isArray(list)) {
            probe.count = list.length;
            probe.firstId = list[0]?.id ?? null;
          }
        } else {
          probe.error = (json && json.message) || `HTTP ${resp.status}`;
        }
      } catch (e) {
        probe.error = e?.message || String(e);
      }
      result.apiProbe.storeAccounts = probe;
    }
    async function probeStoreFinanceApi() {
      const probe = { storeId, url: null, ok: false, status: 0, count: 0, error: null };
      try {
        const apiUrl = new URL(apiBase);
        apiUrl.pathname = `/api/v1/stores/${storeId}/finance/transactions`;
        probe.url = apiUrl.toString();
        const headers = authToken ? { Authorization: `Bearer ${authToken}` } : {};
        const resp = await fetch(probe.url, { method: 'GET', headers });
        probe.status = resp.status;
        const json = await resp.json().catch(() => null);
        if (resp.ok && json) {
          probe.ok = true;
          const rows = json.data || json || [];
          if (Array.isArray(rows)) {
            probe.count = rows.length;
          }
        } else {
          probe.error = (json && json.message) || `HTTP ${resp.status}`;
        }
      } catch (e) {
        probe.error = e?.message || String(e);
      }
      result.apiProbe.storeFinance = probe;
    }
    await Promise.all([probeStoreApi(), probeStoreAccountsApi(), probeStoreFinanceApi()]);
    try {
      // Wait deterministically for either success state or failure state to render.
      const okSentinel = page.getByText(/当前门店：|当前门店:/);
      const notFoundSentinel = page.getByText('未找到门店信息');
      const toastSentinel = page.getByText('门店加载失败');

      const sentinel = await Promise.race([
        okSentinel.waitFor({ timeout: 20000 }).then(() => 'ok'),
        notFoundSentinel.waitFor({ timeout: 20000 }).then(() => 'notFound'),
        toastSentinel.waitFor({ timeout: 20000 }).then(() => 'toastFail'),
      ]).catch(() => 'timeout');

      storeDetailDiagnostics.uiState = sentinel;

      const bodyText = await page.evaluate(() => {
        try { return (document.body && (document.body.innerText || '')) || ''; } catch (_) { return ''; }
      });
      storeDetailDiagnostics.bodyTextLen = bodyText ? bodyText.length : 0;
      storeDetailDiagnostics.bodyTextHead = bodyText ? bodyText.slice(0, 200) : '';

      const apiOk = !!result.apiProbe.storeDetail?.ok;
      const apiName = result.apiProbe.storeDetail?.dataPreview?.name;

      if (sentinel === 'ok') {
        if (apiName && typeof apiName === 'string' && apiName.trim()) {
          if (bodyText.includes(apiName)) ok = true;
          else msg = 'UI reached ok state but store name not found in body text';
        } else {
          ok = true;
        }
      } else if (sentinel === 'notFound' || sentinel === 'toastFail') {
        if (apiOk) {
          msg = `UI shows failure state (${sentinel}) but store API probe is ok; likely H5 request auth/cors/baseURL mismatch`;
        } else {
          // If API also fails, accept failure-state UI as "detected".
          ok = true;
        }
      } else {
        msg = 'Store-detail UI did not reach a known state (timeout)';
      }
    } catch (e) {
      msg = 'Store-detail load failed';
    }

    page.off('console', onConsole);
    page.off('requestfailed', onRequestFailed);

    const shot = path.join(screenshotsDir, 'store-detail.png');
    await page.screenshot({ path: shot, fullPage: true }).catch(() => {});
    result.checks.push({ page: 'store-detail', url, ok, shot, msg, diagnostics: storeDetailDiagnostics });
  }

  try {
    await page.goto(withTkInjected(previewURL), { waitUntil: 'load' });
  } catch (_) {}

  await checkLogin();
  await checkProfile();
  await checkFeedback();
  await checkStoreDetail();

    const out = path.join(logsDir, 'wx-fe-verify-result.json');
    fs.writeFileSync(out, JSON.stringify(result, null, 2));

    const total = result.checks.length;
    const failed = result.checks.filter((c) => !c.ok);
    const passed = total - failed.length;

    const probeEntries = Object.entries(result.apiProbe || {});
    const failedProbes = probeEntries
      .map(([name, probe]) => ({ name, probe }))
      .filter(({ probe }) => probe && probe.ok === false);

    console.log('[wx-fe-verify] Saved result:', out);
    console.log('[wx-fe-verify] Screenshots:', screenshotsDir);
    console.log(`[wx-fe-verify] Summary: ${passed}/${total} checks passed`);

    const hasFailures = failed.length > 0 || failedProbes.length > 0;
    if (hasFailures) {
      if (failed.length) {
        console.log('[wx-fe-verify] Failed checks:');
        for (const c of failed) {
          const msg = (c.msg || '').trim();
          console.log(`- ${c.page}: ${msg || 'failed'}`);
          if (c.diagnostics && (c.diagnostics.uiState || c.diagnostics.bodyTextHead)) {
            const uiState = c.diagnostics.uiState ? String(c.diagnostics.uiState) : '';
            const head = c.diagnostics.bodyTextHead ? String(c.diagnostics.bodyTextHead).replace(/\s+/g, ' ').slice(0, 120) : '';
            console.log(`  diagnostics: uiState=${uiState} bodyHead=${head}`);
          }
        }
      }

      if (failedProbes.length) {
        console.log('[wx-fe-verify] Failed api probes:');
        for (const { name, probe } of failedProbes) {
          const status = probe && typeof probe.status === 'number' ? probe.status : 0;
          const url = probe && probe.url ? String(probe.url) : '';
          const err = probe && probe.error ? String(probe.error) : '';
          console.log(`- ${name}: ok=false status=${status} ${url}`);
          if (err) console.log(`  error: ${err}`);
        }
      }

      exitCode = 2;
    }
  } finally {
    try {
      if (browser) await browser.close();
    } catch (_) {}
    browser = null;
    try {
      if (localServer) await new Promise((r) => localServer.close(r));
    } catch (_) {}
    localServer = null;
  }

  if (exitCode) process.exit(exitCode);
})().catch((err) => {
  console.error('[wx-fe-verify] failed:', err);
  process.exit(1);
});
