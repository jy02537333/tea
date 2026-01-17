#!/usr/bin/env node
/**
 * Bootstrap a fresh AUTH_TOKEN for wx-fe verification (host-side Playwright + H5 auth guard).
 *
 * What it does:
 * 1) Call /api/v1/user/dev-login with an OpenID (default: wxfe_store_manager_openid)
 * 2) Save response to build-ci-logs/admin_login_response.json (used by verifier fallbacks)
 * 3) Save AUTH_TOKEN=... to build-ci-logs/tokens.env
 * 4) Run tools/automation/setup-store-manager.js to ensure RBAC role/permissions and assign to this user
 *
 * Env:
 * - WX_API_BASE_URL or API_BASE (default: http://127.0.0.1:9292)
 * - AUTH_OPENID (default: wxfe_store_manager_openid)
 */

const fs = require('fs');
const path = require('path');
const http = require('http');
const https = require('https');
const { spawnSync } = require('child_process');

function getApiBase() {
  const v = (process.env.API_BASE || process.env.WX_API_BASE_URL || '').trim();
  return v || 'http://127.0.0.1:9292';
}

function parseEnvFile(p) {
  const out = {};
  try {
    const txt = fs.readFileSync(p, 'utf-8');
    for (const line of txt.split(/\r?\n/)) {
      const m = line.match(/^([A-Z0-9_]+)=(.*)$/);
      if (m) out[m[1]] = m[2];
    }
  } catch (_) {}
  return out;
}

function writeEnvKV(p, kv) {
  fs.mkdirSync(path.dirname(p), { recursive: true });
  const existing = fs.existsSync(p) ? fs.readFileSync(p, 'utf-8') : '';
  const lines = existing.split(/\r?\n/).filter((l) => l.trim().length > 0);
  const kept = [];
  const replaced = new Set(Object.keys(kv));
  for (const line of lines) {
    const k = (line.split('=')[0] || '').trim();
    if (!k) continue;
    if (replaced.has(k)) continue;
    kept.push(line);
  }
  for (const [k, v] of Object.entries(kv)) {
    kept.push(`${k}=${v}`);
  }
  fs.writeFileSync(p, kept.join('\n') + '\n');
}

function requestJson(url, { method = 'GET', headers = {}, body = null } = {}) {
  return new Promise((resolve, reject) => {
    const u = new URL(url);
    const isHttps = u.protocol === 'https:';
    const lib = isHttps ? https : http;
    const data = body ? JSON.stringify(body) : null;

    const req = lib.request(
      {
        hostname: u.hostname,
        port: u.port || (isHttps ? 443 : 80),
        path: u.pathname + (u.search || ''),
        method,
        headers: {
          Accept: 'application/json',
          ...(data ? { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(data) } : {}),
          ...headers,
        },
      },
      (res) => {
        let buf = '';
        res.setEncoding('utf8');
        res.on('data', (chunk) => (buf += chunk));
        res.on('end', () => {
          const status = res.statusCode || 0;
          try {
            const json = buf ? JSON.parse(buf) : null;
            resolve({ status, data: json, raw: buf });
          } catch (_) {
            resolve({ status, data: null, raw: buf });
          }
        });
      },
    );
    req.on('error', reject);
    if (data) req.write(data);
    req.end();
  });
}

async function main() {
  const repoRoot = path.resolve(__dirname, '../../');
  const logsDir = path.join(repoRoot, 'build-ci-logs');
  const tokensPath = path.join(logsDir, 'tokens.env');
  const apiBase = getApiBase();
  const openid = String(process.env.AUTH_OPENID || 'wxfe_store_manager_openid').trim();

  if (!openid) {
    console.error('[bootstrap-wx-fe-token] missing AUTH_OPENID');
    process.exit(2);
  }

  fs.mkdirSync(logsDir, { recursive: true });

  // 1) dev-login to obtain a fresh token
  const loginResp = await requestJson(`${apiBase}/api/v1/user/dev-login`, {
    method: 'POST',
    body: { openid },
  });

  if (loginResp.status !== 200 || !loginResp.data) {
    console.error('[bootstrap-wx-fe-token] dev-login failed:', loginResp.status, loginResp.raw ? loginResp.raw.slice(0, 200) : '');
    process.exit(3);
  }

  const token = loginResp.data?.data?.token || loginResp.data?.token || '';
  if (!token) {
    console.error('[bootstrap-wx-fe-token] dev-login returned no token');
    process.exit(4);
  }

  // 2) write admin_login_response.json (verifier fallback)
  fs.writeFileSync(path.join(logsDir, 'admin_login_response.json'), JSON.stringify(loginResp.data, null, 2));

  // 3) update tokens.env with AUTH_TOKEN and API_BASE for other tools
  writeEnvKV(tokensPath, { AUTH_TOKEN: token, API_BASE: apiBase });

  console.log('[bootstrap-wx-fe-token] AUTH_TOKEN written to build-ci-logs/tokens.env');
  console.log('[bootstrap-wx-fe-token] login response written to build-ci-logs/admin_login_response.json');

  // 4) ensure RBAC role + permissions and assign to this user
  const child = spawnSync(process.execPath, [path.join(__dirname, 'setup-store-manager.js')], {
    stdio: 'inherit',
    env: { ...process.env, API_BASE: apiBase, WX_API_BASE_URL: apiBase, AUTH_TOKEN: token },
  });
  if (child.status !== 0) {
    console.error('[bootstrap-wx-fe-token] setup-store-manager failed with exit', child.status);
    process.exit(child.status || 5);
  }

  // Sanity: ensure tokens.env still has AUTH_TOKEN
  const kv = parseEnvFile(tokensPath);
  if (!kv.AUTH_TOKEN) {
    console.error('[bootstrap-wx-fe-token] tokens.env missing AUTH_TOKEN after setup');
    process.exit(6);
  }

  console.log('[bootstrap-wx-fe-token] done');
}

main().catch((err) => {
  console.error('[bootstrap-wx-fe-token] failed:', err?.message || err);
  process.exit(1);
});
