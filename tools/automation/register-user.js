#!/usr/bin/env node
/*
 Register a new user via unified auth login using phone+code.
 Since current auth handler validates only lengths, we generate a synthetic phone and code.
 Outputs:
  - build-ci-logs/new_user_credentials.txt: phone and sms code
  - build-ci-logs/user_login_response_latest.json: full login response
  - build-ci-logs/tokens.env: AUTH_TOKEN updated (when token present in response)
 Env:
  - WX_API_BASE_URL (default http://127.0.0.1:9292)
  - AUTH_PHONE / AUTH_SMS_CODE (optional; if absent, generated)
*/
const fs = require('fs');
const path = require('path');
const http = require('http');
const https = require('https');

function writeFile(p, content) { fs.mkdirSync(path.dirname(p), { recursive: true }); fs.writeFileSync(p, content); }

async function main() {
  const repoRoot = path.resolve(__dirname, '../../');
  const logsDir = path.join(repoRoot, 'build-ci-logs');
  const apiBase = process.env.WX_API_BASE_URL || 'http://127.0.0.1:9292';
  let phone = process.env.AUTH_PHONE || '';
  let smsCode = process.env.AUTH_SMS_CODE || '';

  // Generate synthetic credentials if not provided
  const now = Date.now();
  if (!phone) phone = '171' + String(now).slice(-8); // e.g., 171xxxxxx (>= 6 digits)
  if (!smsCode) smsCode = String(now).slice(-4);      // 4-digit code

  const url = new URL(apiBase);
  url.pathname = '/api/v1/auth/login';
  const payload = { phone, code: smsCode };

  const text = await postJSON(url, payload);
  let json = null;
  try { json = JSON.parse(text.body); } catch (_) {}

  if (text.status !== 200 || !json) {
    console.error('[register-user] login/register failed:', text.status, text.body);
    process.exit(2);
  }

  const token = (json.data && json.data.token) || '';
  // Persist credentials
  const credFile = path.join(logsDir, 'new_user_credentials.txt');
  writeFile(credFile, `AUTH_PHONE=${phone}\nAUTH_SMS_CODE=${smsCode}\n`);

  // Persist login response
  const outJson = path.join(logsDir, 'user_login_response_latest.json');
  writeFile(outJson, JSON.stringify(json, null, 2));

  // Update tokens.env if token exists
  if (token) {
    const envFile = path.join(logsDir, 'tokens.env');
    let existing = '';
    try { if (fs.existsSync(envFile)) existing = fs.readFileSync(envFile, 'utf-8'); } catch (_) {}
    const lines = existing.split(/\r?\n/).filter(Boolean).filter((l) => !/^AUTH_TOKEN=/.test(l));
    lines.push(`AUTH_TOKEN=${token}`);
    writeFile(envFile, lines.join('\n') + '\n');
  }

  console.log('[register-user] created credentials at', credFile);
  if (token) console.log('[register-user] token captured and tokens.env updated');
}

main().catch((err) => { console.error('[register-user] failed:', err); process.exit(1); });

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
