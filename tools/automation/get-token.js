#!/usr/bin/env node
/*
 Get API token automatically and record it for later use.
 Modes:
  - Use env credentials to login: AUTH_USERNAME/AUTH_PASSWORD or AUTH_CODE or AUTH_OPENID
  - Fallback: reuse existing login response in build-ci-logs (admin_login_response.json or user_login_response_latest.json)
 Outputs:
  - Writes token to build-ci-logs/tokens.env (AUTH_TOKEN=...)
  - Writes token to role-specific files: admin_token.txt or user_token.txt
  - Writes full response JSON to build-ci-logs/{admin|user}_login_response.json
 Env:
  - WX_API_BASE_URL (default http://127.0.0.1:9292)
  - AUTH_USERNAME / AUTH_PASSWORD
  - AUTH_CODE
  - AUTH_OPENID
  - ROLE=admin|user (controls output file naming; default admin when username provided, else user)
*/
const fs = require('fs');
const path = require('path');

function readJSONSafe(p, def = null) { try { return JSON.parse(fs.readFileSync(p, 'utf-8')); } catch (_) { return def; } }
function writeFile(p, content) { fs.mkdirSync(path.dirname(p), { recursive: true }); fs.writeFileSync(p, content); }

const http = require('http');
const https = require('https');

async function main() {
  const repoRoot = path.resolve(__dirname, '../../');
  const logsDir = path.join(repoRoot, 'build-ci-logs');
  const apiBase = process.env.WX_API_BASE_URL || 'http://127.0.0.1:9292';
  const username = process.env.AUTH_USERNAME || '';
  const password = process.env.AUTH_PASSWORD || '';
  const code = process.env.AUTH_CODE || '';
  const wechatCode = process.env.AUTH_WECHAT_CODE || '';
  const phone = process.env.AUTH_PHONE || '';
  const smsCode = process.env.AUTH_SMS_CODE || '';
  let role = (process.env.ROLE || '').toLowerCase();
  if (!role) role = username ? 'admin' : 'user';

  const payload = {};
  if (username && password) {
    payload.username = username; payload.password = password;
  } else if (phone && smsCode) {
    payload.phone = phone; payload.code = smsCode;
  } else if (code) {
    payload.code = code;
  } else if (wechatCode) {
    payload.wechat_code = wechatCode;
  }

  async function tryLogin(urlPath) {
    const url = new URL(apiBase);
    url.pathname = urlPath;
    const res = await postJSON(url, payload);
    let json = null;
    try { json = JSON.parse(res.body); } catch (_) {}
    return { ok: res.status >= 200 && res.status < 300, status: res.status, json, text: res.body };
  }

  let token = '';
  let respBundle = null;
  if (Object.keys(payload).length > 0) {
    // Prefer unified auth endpoint
    let r = await tryLogin('/api/v1/auth/login');
    if (!r.ok) {
      // Fallback to legacy user login
      r = await tryLogin('/api/v1/user/login');
    }
    if (r.ok && r.json) {
      const data = r.json.data || r.json;
      token = data?.token || '';
      respBundle = r.json;
    } else {
      console.error('[get-token] login failed:', r.status, (r.json && r.json.message) || r.text);
    }
  }

  // Fallback: reuse existing login response
  if (!token) {
    const adminResp = readJSONSafe(path.join(logsDir, 'admin_login_response.json'), null);
    const userRespA = readJSONSafe(path.join(logsDir, 'user_login_response_latest.json'), null);
    const userRespB = readJSONSafe(path.join(logsDir, 'user_login_response.json'), null);
    const candidate = adminResp || userRespA || userRespB;
    const data = candidate?.data || candidate;
    const existing = data?.token || '';
    if (existing) {
      token = existing;
      respBundle = candidate;
      if (!role) role = adminResp ? 'admin' : 'user';
      console.log('[get-token] reused existing token from logs');
    }
  }

  if (!token) {
    console.error('[get-token] no token acquired. Provide AUTH_USERNAME/AUTH_PASSWORD or AUTH_CODE/OPENID, or ensure logs have a previous login response.');
    process.exit(2);
  }

  // Persist token and response
  const tokenFile = path.join(logsDir, role === 'admin' ? 'admin_token.txt' : 'user_token.txt');
  writeFile(tokenFile, token + '\n');
  const envFile = path.join(logsDir, 'tokens.env');
  try {
    let existing = '';
    if (fs.existsSync(envFile)) existing = fs.readFileSync(envFile, 'utf-8');
    const lines = existing.split(/\r?\n/).filter(Boolean).filter((l) => !/^AUTH_TOKEN=/.test(l));
    lines.push(`AUTH_TOKEN=${token}`);
    writeFile(envFile, lines.join('\n') + '\n');
  } catch (e) {
    console.error('[get-token] write tokens.env failed:', e?.message || e);
  }

  const outJson = path.join(logsDir, role === 'admin' ? 'admin_login_response.json' : 'user_login_response_latest.json');
  if (respBundle) {
    writeFile(outJson, JSON.stringify(respBundle, null, 2));
  }

  console.log('[get-token] token saved:', tokenFile);
  console.log('[get-token] tokens.env updated at', envFile);
}

main().catch((err) => { console.error('[get-token] failed:', err); process.exit(1); });

function postJSON(url, payload) {
  return new Promise((resolve, reject) => {
    const data = JSON.stringify(payload);
    const isHttps = url.protocol === 'https:';
    const lib = isHttps ? https : http;
    const options = {
      hostname: url.hostname,
      port: url.port || (isHttps ? 443 : 80),
      path: url.pathname + (url.search || ''),
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(data),
      },
    };
    const req = lib.request(options, (res) => {
      let body = '';
      res.setEncoding('utf8');
      res.on('data', (chunk) => { body += chunk; });
      res.on('end', () => resolve({ status: res.statusCode, body }));
    });
    req.on('error', reject);
    req.write(data);
    req.end();
  });
}
